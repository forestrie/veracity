package veracity

// Watch for log changes, relying on the blob last idtimestamps to do so
// efficiently.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"time"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	azwatcher "github.com/robinbryce/go-merklelog-azure/watcher"

	// "github.com/datatrails/go-datatrails-common/azblob"
	"github.com/urfave/cli/v2"
)

const (
	flagCount    = "count"
	flagHorizon  = "horizon"
	flagIDSince  = "idsince"
	flagInterval = "interval"
	flagLatest   = "latest"
	flagSince    = "since"

	currentEpoch   = uint8(1) // good until the end of the first unix epoch
	tenantPrefix   = "tenant/"
	sealIDNotFound = "NOT-FOUND"
	// maxPollCount is the maximum number of times to poll for *some* activity.
	// Polling always terminates as soon as the first activity is detected.
	maxPollCount = 15
	// More than this over flows the epoch which is half the length of the unix time epoch
	maxHorizon                       = time.Hour * 100000
	horizonAliasMax                  = "max"    // short hand for the largest supported duration
	sinceAliasLatest                 = "latest" // short hand for obtaining the latest change for all watched tenants
	rangeDurationParseErrorSubString = "time: invalid duration "
	threeSeconds                     = 3 * time.Second
)

var (
	ErrNoChanges = errors.New("no changes found")
)

type WatchConfig struct {
	azwatcher.WatchConfig
}

// watchReporter abstracts the output interface for WatchForChanges to facilitate unit testing.
type watchReporter interface {
	Logf(message string, args ...any)
	Outf(message string, args ...any)
}

type defaultReporter struct {
	log logger.Logger
}

func (r defaultReporter) Logf(message string, args ...any) {
	if r.log == nil {
		return
	}
	r.log.Infof(message, args...)
}
func (r defaultReporter) Outf(message string, args ...any) {
	fmt.Printf(message, args...)
}

// NewLogWatcherCmd watches for changes on any log
func NewLogWatcherCmd() *cli.Command {
	return &cli.Command{Name: "watch",
		Usage: `discover recently active logs
		
		Provide --horizon OR provide either of --since or --idsince

		horizon is always inferred from the since arguments if they are provided
		`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  flagLatest,
				Usage: `find the latest changes for each requested tenant (no matter how long ago they occurred). This is mutually exclusive with --since, --idsince and --horizon.`,
				Value: false,
			},

			&cli.TimestampFlag{
				Name:   flagSince,
				Usage:  "RFC 3339 time stamp, only logs with changes after this are considered, defaults to now. idsince takes precendence if also supplied.",
				Layout: time.RFC3339,
			},
			&cli.StringFlag{
				Name: flagIDSince, Aliases: []string{"s"},
				Usage: "Start time as an idtimestamp. Start time defaults to now. All results are >= this hex string. If provided, it is used exactly as is. Takes precedence over since",
			},
			&cli.StringFlag{
				Name:    flagHorizon,
				Aliases: []string{"z"},
				Value:   "24h",
				Usage:   "Infer since as now - horizon. Use the alias --horizon=max to force the highest supported value. Otherwise, the format is {number}{units} eg 1h to only see things in the last hour. If watching (count=0), since is re-calculated every interval",
			},
			&cli.DurationFlag{
				Name: flagInterval, Aliases: []string{"d"},
				Value: threeSeconds,
				Usage: "The default polling interval is once every three seconds, setting the interval to zero disables polling",
			},
			&cli.IntFlag{
				Name: flagCount, Usage: fmt.Sprintf(
					"Number of intervals to poll. Polling is terminated once the first activity is seen or after %d attempts regardless", maxPollCount),
				Value: 1,
			},
		},
		Action: func(cCtx *cli.Context) error {

			var err error
			cmd := &CmdCtx{}
			ctx := context.Background()

			if err = cfgLogging(cmd, cCtx); err != nil {
				return err
			}
			reporter := &defaultReporter{log: cmd.Log}

			cfg, err := NewWatchConfig(cCtx, cmd)
			if err != nil {
				return err
			}

			dataUrl := cCtx.String("data-url")
			if dataUrl == "" && !IsStorageEmulatorEnabled(cCtx) {
				dataUrl = DefaultRemoteMassifURL
			}


			reader, err := cfgReader(cmd, cCtx, dataUrl)
			if err != nil {
				return err
			}

			collator := azwatcher.NewLogTailCollator(
				func(storagePath string) storage.LogID {
					return storage.ParsePrefixedLogID("tenant/", storagePath)
				},
				storage.ObjectIndexFromPath,
			)
			watcher, err := azwatcher.NewWatcher(cfg.WatchConfig)
			if err != nil {
				return err
			}
			wc := &WatcherCollator{
				Watcher:         watcher,
				LogTailCollator: collator,
			}

			return azwatcher.WatchForChanges(ctx, cfg.WatchConfig, wc, reader, reporter)
		},
	}
}

func checkCompatibleFlags(cCtx cliContext) error {
	if !cCtx.IsSet(flagLatest) {
		return nil
	}

	latestExcludes := []string{flagHorizon, flagSince, flagIDSince}

	for _, excluded := range latestExcludes {
		if cCtx.IsSet(excluded) {
			return fmt.Errorf("the %s flag is mutually exclusive with %s", flagLatest, strings.Join(latestExcludes, ", "))
		}
	}
	return nil
}

type cliContext interface {
	IsSet(string) bool
	Bool(string) bool
	Duration(name string) time.Duration
	Timestamp(name string) *time.Time
	String(name string) string
	Int(name string) int
}

// parseHorizon parses a duration string from the command line In accordance
// with the most common reason for parse failure (specifying a large number), On
// an error that looks like a range to large issue, we coerce to the maximum
// hours and ignore the error. Errors that don't contain the marker substring
// are returned as is.
func parseHorizon(horizon string) (time.Duration, error) {

	if horizon == horizonAliasMax {
		return maxHorizon, nil
	}

	d, err := time.ParseDuration(horizon)
	if err == nil {

		if d > maxHorizon {
			return 0, fmt.Errorf("the maximum supported duration is --horizon=%v, which has the alias --horizon=max. also consider using --latest", maxHorizon)
		}
		if d < 0 {
			return 0, fmt.Errorf("negative horizon value:%s", horizon)
		}

		return d, nil
	}

	if strings.HasPrefix(err.Error(), rangeDurationParseErrorSubString) {
		return 0, fmt.Errorf("the supplied horizon was invalid. the maximum supported duration is --horizon=%v, which has the alias --horizon=max. also consider using --latest", maxHorizon)
	}

	return d, fmt.Errorf("the horizon '%s' is out of range or otherwise invalid. Use --horizon=max to get the largest supported value %v. also consider using --latest", horizon, maxHorizon)
}

// NewWatchConfig derives a configuration from the options set on the command line context
func NewWatchConfig(cCtx cliContext, cmd *CmdCtx) (WatchConfig, error) {

	var err error

	// --latest is mutualy exclusive with the horizon, since, idsince flags.
	if err = checkCompatibleFlags(cCtx); err != nil {
		return WatchConfig{}, err
	}

	cfg := WatchConfig{
		WatchConfig: azwatcher.WatchConfig{
			Latest:   cCtx.Bool(flagLatest),
			Interval: cCtx.Duration(flagInterval),
		},
	}
	if cfg.Interval == 0 {
		cfg.Interval = threeSeconds
	}

	if cCtx.IsSet(flagHorizon) {
		cfg.Horizon, err = parseHorizon(cCtx.String(flagHorizon))
		if err != nil {
			return WatchConfig{}, err
		}
	}

	if cCtx.IsSet(flagSince) {
		cfg.Since = *cCtx.Timestamp(flagSince)
	}
	if cCtx.IsSet(flagIDSince) {
		cfg.IDSince = cCtx.String(flagIDSince)
	}

	err = azwatcher.ConfigDefaults(&cfg.WatchConfig)
	if err != nil {
		return WatchConfig{}, err
	}
	if cfg.Interval < time.Second {
		return WatchConfig{}, fmt.Errorf("polling more than once per second is not currently supported")
	}

	cfg.WatchCount = min(max(1, cCtx.Int(flagCount)), maxPollCount)

	cfg.ObjectPrefixURL = cmd.RemoteURL

	logs := CtxGetLogOptions(cCtx)
	if len(logs) == 0 {
		return cfg, nil
	}

	cfg.WatchLogs = make(map[string]bool)
	for _, lid := range logs {
		cfg.WatchLogs[string(lid)] = true
	}
	return cfg, nil
}

type WatcherCollator struct {
	azwatcher.Watcher
	azwatcher.LogTailCollator
}
