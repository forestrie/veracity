package veracity

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	"github.com/datatrails/go-datatrails-merklelog/massifs/watcher"
	"github.com/gosuri/uiprogress"
	azblobs "github.com/robinbryce/go-merklelog-azure/blobs"
	azwatcher "github.com/robinbryce/go-merklelog-azure/watcher"
	"github.com/urfave/cli/v2"
	"github.com/veraison/go-cose"
	"golang.org/x/exp/rand"
)

const (
	// baseDefaultRetryDelay is the base delay for retrying transient errors. A little jitter is added.
	// 429 errors which provide a valid Retry-After header will honor that header rather than use this.
	baseDefaultRetryDelay = 2 * time.Second
	defaultConcurrency    = 5

	// The default data retention policy is 2 years, so this is a generous default for "all data".
	tenYearsOfHours = 10 * 365 * 24 * time.Hour

	// jitterRangeMS is the range from 0 to jitter in milliseconds
	jitterRangeMS = 100

	// massifHeightMax is the maximum massif height
	massifHeightMax = 255
)

var (
	ErrChangesFlagIsExclusive          = errors.New("use --changes Or --massif and --tenant, not both")
	ErrNewReplicaNotEmpty              = errors.New("the local directory for a new replica already exists")
	ErrSealNotFound                    = errors.New("seal not found")
	ErrSealVerifyFailed                = errors.New("the seal signature verification failed")
	ErrFailedCheckingConsistencyProof  = errors.New("failed to check a consistency proof")
	ErrFailedToCreateReplicaDir        = errors.New("failed to create a directory needed for local replication")
	ErrRequiredOption                  = errors.New("a required option was not provided")
	ErrRemoteLogTruncated              = errors.New("the local replica indicates the remote log has been truncated")
	ErrRemoteLogInconsistentRootState  = errors.New("the local replica root state disagrees with the remote")
	ErrInconsistentUseOfPrefetchedSeal = errors.New("prefetching signed root reader used inconsistently")
)

// NewReplicateLogsCmd updates a local replica of a remote log, verifying the mutual consistency of the two before making any changes.
//
//nolint:gocognit
func NewReplicateLogsCmd() *cli.Command {
	return &cli.Command{
		Name:    "replicate-logs",
		Aliases: []string{"replicate"},
		Usage:   `verifies the remote log and replicates it locally, ensuring the remote changes are consistent with the trusted local replica.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: skipUncommittedFlagName, Value: false},
			&cli.IntFlag{
				Name: "massif", Aliases: []string{"m"},
			},
			&cli.UintFlag{
				Usage: `The number of massif 'ancestors' to retain in the local replica.
This many massifs, prior to the requested, will be verified and retained localy.
If more exist locally, they are not removed (or reverified). If set to 0, a full replica is requested.`,
				Value: 0,
				Name:  "ancestors", Aliases: []string{"a"},
			},
			&cli.StringFlag{
				Name: "replicadir",
				Usage: `the root directory for all tenant log replicas,
the structure under this directory mirrors the /verifiabledata/merklelogs paths
in the publicly accessible remote storage`,
				Aliases: []string{"d"},
				Value:   ".",
			},
			&cli.StringFlag{
				Name:    "checkpoint-public",
				Usage:   `A COSE Key format file, containing the key to use to verify checkpoint signatures, ES2 only.`,
				Aliases: []string{"pub"},
			},
			&cli.StringFlag{
				Name:    "checkpoint-jwks",
				Usage:   `A JWKS format file, whose *last* entry is the key to use to verify checkpoint signatures. ES only`,
				Aliases: []string{"jwks"},
			},

			&cli.StringFlag{
				Name: "changes",
				Usage: `
provide the path to a file enumerating the tenant massifs with changes you want
to verify and replicate.  This is mutually exclusive with the --massif and
--tenant flags. If none of --massif, --tenant or --changes are provided, the
changes are read from standard input.`,
			},
			&cli.BoolFlag{
				Name:    "progress",
				Usage:   `show progress of the replication process`,
				Value:   false,
				Aliases: []string{"p"},
			},
			&cli.BoolFlag{
				Name:  "latest",
				Usage: `find the latest changes automaticaly. When --latest is set, a list of tenants can be provided to --tenant to limit the tenant logs to be replicated.`,
				Value: false,
			},
			&cli.IntFlag{
				Name:    "retries",
				Aliases: []string{"r"},
				Value:   -1, // -1 means no limit
				Usage: `
Set a maximum number of retries for transient error conditions. Set 0 to disable retries.
By default transient errors are re-tried without limit, and if the error is 429, the Retry-After header is honored.`,
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Value:   defaultConcurrency,
				Aliases: []string{"c"},
				Usage: fmt.Sprintf(
					`The number of concurrent replication operations to run, defaults to %d. A high number is a sure way to get rate limited`, defaultConcurrency),
			},
		},
		Action: func(cCtx *cli.Context) error {
			cmd := &CmdCtx{}

			var err error
			// The loggin configuration is safe to share accross go routines.
			if err = cfgLogging(cmd, cCtx); err != nil {
				return err
			}

			if err = CfgKeys(cmd, cCtx); err != nil {
				return err
			}

			dataUrl := cCtx.String("data-url")
			if dataUrl == "" && !IsStorageEmulatorEnabled(cCtx) {
				dataUrl = DefaultRemoteMassifURL
			}
			if dataUrl == "" {
				return fmt.Errorf("%w: remote-url is required", ErrRequiredOption)
			}
			cmd.RemoteURL = dataUrl

			// There isn't really a better context. We could implement user
			// defined timeouts for "lights out/ci" use cases in future. Humans can ctrl-c
			changes, err := readTenantMassifChanges(context.Background(), cCtx, cmd)
			if err != nil {
				return err
			}

			if cCtx.Bool("progress") {
				uiprogress.Start()
			}
			progress := newProgressor(cCtx, "tenants", len(changes))

			concurency := min(len(changes), max(1, cCtx.Int("concurrency")))
			for i := 0; i < len(changes); i += concurency {
				err = replicateChanges(cCtx, cmd, changes[i:min(i+concurency, len(changes))], progress)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// replicateChanges replicate the changes for the provided slice of tenants.
// Paralelism is limited by breaking the total changes into smaller slices and calling this function
func replicateChanges(cCtx *cli.Context, cmd *CmdCtx, changes []watcher.LogMassif, progress Progresser) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(changes)) // buffered so it doesn't block

	for _, change := range changes {
		wg.Add(1)
		go func(change watcher.LogMassif, errChan chan<- error) {
			defer wg.Done()
			defer progress.Completed()

			retries := max(-1, cCtx.Int("retries"))
			for {

				replicator, startMassif, endMassif, err := initReplication(cCtx, cmd, change)
				if err != nil {
					errChan <- err
					return
				}

				// There isn't really a better context. We could implement user
				// defined timeouts for "lights out/ci" use cases in future. Humans can ctrl-c
				err = replicator.ReplicateVerifiedUpdates(
					context.Background(),
					startMassif, endMassif,
				)
				if err == nil {
					return
				}

				// 429 is the only transient error we currently re-try
				var retryDelay time.Duration
				retryDelay, ok := azblobs.IsRateLimiting(err)
				if !ok || retries == 0 {
					// not transient
					errChan <- err
					return
				}
				if retryDelay == 0 {
					retryDelay = defaultRetryDelay(err)
				}

				// underflow will actually terminate the loop, but that would have been running for an infeasible amount of time
				retries--
				// in the default case, remaining is always reported as -1
				cmd.Log.Infof("retrying in %s, remaining: %d", retryDelay, max(-1, retries))
			}
		}(change, errChan)
	}

	// the error channel is buffered enough for each tenant, so this will not get deadlocked
	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		cmd.Log.Infof("%v", err)
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs[0]
	}
	if len(changes) == 1 {
		cmd.Log.Infof("replication complete for log %x", changes[0].LogID)
	} else {
		cmd.Log.Infof("replication complete for %d logs", len(changes))
	}
	return nil
}

func initReplication(cCtx *cli.Context, cmd *CmdCtx, change watcher.LogMassif) (*VerifiedReplica, uint32, uint32, error) {
	replicator, err := NewVerifiedReplica(cCtx, cmd.Clone(), change.LogID)
	if err != nil {
		return nil, 0, 0, err
	}
	endMassif := uint32(change.Massif)
	startMassif := uint32(0)
	if cCtx.IsSet("ancestors") && uint32(cCtx.Int("ancestors")) < endMassif {
		startMassif = endMassif - uint32(cCtx.Int("ancestors"))
	}
	return replicator, startMassif, endMassif, nil
}

func defaultRetryDelay(_ error) time.Duration {
	// give the delay some jitter, this is universally a good practice
	return baseDefaultRetryDelay + time.Duration(rand.Intn(jitterRangeMS))*time.Millisecond
}

func newProgressor(cCtx *cli.Context, barName string, increments int) Progresser {
	if !cCtx.Bool("progress") {
		return NewNoopProgress()
	}
	return NewStagedProgress(barName, increments)
}

type VerifiedReplica struct {
	massifs.VerifyingReplicator
	cCtx *cli.Context
	log  logger.Logger
}

func NewVerifiedReplica(
	cCtx *cli.Context, cmd *CmdCtx, logID storage.LogID,
) (*VerifiedReplica, error) {

	var err error

	if err := cfgMassifFmt(cmd, cCtx); err != nil {
		return nil, err
	}

	if cmd.MassifFmt.MassifHeight > massifHeightMax {
		return nil, fmt.Errorf("massif height must be less than 256")
	}
	if cmd.MassifFmt.MassifHeight == 0 {
		return nil, fmt.Errorf("massif height must be initialized")
	}

	if cmd.RemoteURL == "" {
		return nil, fmt.Errorf("%w: remote-url is required", ErrRequiredOption)
	}

	reader, err := cfgReader(cmd, cCtx, cmd.RemoteURL)
	if err != nil {
		return nil, err
	}

	dataUrl := cmd.RemoteURL // may be azurite in emulator mode, which overrides

	remoteReader, err := NewCmdStorageProviderAzure(context.Background(), cCtx, cmd, dataUrl, reader)
	if err != nil {
		return nil, err
	}
	if err = remoteReader.SelectLog(context.Background(), logID); err != nil {
		return nil, fmt.Errorf("failed to select remote log %s: %w", logID, err)
	}
	localReader, err := NewCmdStorageProviderFS(
		context.Background(), cCtx, cmd, cCtx.String("replicadir"), true)
	if err != nil {
		return nil, err
	}

	if err = localReader.SelectLog(context.Background(), logID); err != nil {
		return nil, fmt.Errorf("failed to select local log %s: %w", logID, err)
	}

	var verifier cose.Verifier

	if cmd.CheckpointPublic.Public != nil {
		verifier, err = cose.NewVerifier(cmd.CheckpointPublic.Alg, cmd.CheckpointPublic.Public)
		if err != nil {
			return nil, err
		}
	}

	return &VerifiedReplica{
		cCtx: cCtx,
		log:  logger.Sugar,
		VerifyingReplicator: massifs.VerifyingReplicator{
			CBORCodec:    cmd.CBORCodec,
			COSEVerifier: verifier,
			Sink:         localReader,
			Source:       remoteReader,
		},
	}, nil
}

type changeCollector struct {
	log         logger.Logger
	watchOutput string
}

func (c *changeCollector) Logf(msg string, args ...any) {
	if c.log == nil {
		return
	}
	c.log.Infof(msg, args...)
}

func (c *changeCollector) Outf(msg string, args ...any) {
	c.watchOutput += fmt.Sprintf(msg, args...)
}

func newWatchConfig(cCtx *cli.Context, cmd *CmdCtx) (WatchConfig, error) {
	cfg := WatchConfig{
		WatchConfig: azwatcher.WatchConfig{
			// Latest:     cCtx.Bool("latest"),
			WatchCount: 1,
			Horizon:    tenYearsOfHours,
		},
	}
	err := azwatcher.ConfigDefaults(&cfg.WatchConfig)
	if err != nil {
		return WatchConfig{}, err
	}
	cfg.ObjectPrefixURL = cmd.RemoteURL

	logids := CtxGetLogOptions(cCtx)
	if len(logids) == 0 {
		return cfg, nil
	}

	cfg.WatchLogs = make(map[string]bool)
	for _, lid := range logids {
		cfg.WatchLogs[string(lid)] = true
	}
	return cfg, nil
}

func readTenantMassifChanges(ctx context.Context, cCtx *cli.Context, cmd *CmdCtx) ([]watcher.LogMassif, error) {
	if cCtx.IsSet("latest") {
		// This is because people get tripped up with the `veracity watch -z 90000h | veracity replicate-logs` idiom,
		// Its such a common use case that we should just make it work.
		cfg, err := newWatchConfig(cCtx, cmd)
		if err != nil {
			return nil, err
		}

		if cmd.RemoteURL == "" {
			return nil, fmt.Errorf("%w: remote-url is required", ErrRequiredOption)
		}

		reader, err := cfgReader(cmd, cCtx, cmd.RemoteURL)
		if err != nil {
			return nil, err
		}
		collator := azwatcher.NewLogTailCollator(
			func(storagePath string) storage.LogID {
				return storage.ParsePrefixedLogID("tenant/", storagePath)
			},
			storage.ObjectIndexFromPath,
		)
		watcher, err := azwatcher.NewWatcher(cfg.WatchConfig)
		if err != nil {
			return nil, err
		}
		wc := &WatcherCollator{
			Watcher:         watcher,
			LogTailCollator: collator,
		}

		collector := &changeCollector{log: cmd.Log}
		err = azwatcher.WatchForChanges(ctx, cfg.WatchConfig, wc, reader, collector)
		if err != nil {
			return nil, err
		}

		return scannerToLogMassifs(bufio.NewScanner(strings.NewReader(collector.watchOutput)))
	}

	logs := CtxGetLogOptions(cCtx)
	if len(logs) == 1 {
		return []watcher.LogMassif{{LogID: logs[0], Massif: cCtx.Int("massif")}}, nil
	}
	if len(logs) > 1 {
		return nil, fmt.Errorf("multiple logs may only be used with --latest")
	}

	// If --changes is set the logs and massif indices are read from the identified file
	changesFile := cCtx.String("changes")
	if changesFile != "" {
		return filePathToLogMassifs(changesFile)
	}

	// No explicit config and --all not set, read from stdin
	return stdinToDecodedLogMassifs()
}
