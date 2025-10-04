package veracity

import (
	"context"
	"fmt"
	"time"

	commoncbor "github.com/datatrails/go-datatrails-common/cbor"
	"github.com/datatrails/go-datatrails-common/cose"
	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/datatrails/go-datatrails-merklelog/massifs/snowflakeid"
	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	"github.com/datatrails/go-datatrails-merklelog/massifs/watcher"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

type TailConfig struct {
	// Interval defines the wait period between repeated tail checks if many
	// checks have been asked for.
	Interval time.Duration
	LogID    storage.LogID
}

// LogTailActivity can represent either the seal or the massif that has most recently
// been updated for the log.
type LogTailActivity struct {
	watcher.LogTail
	LastIDEpoch     uint8
	LastIDTimestamp uint64
	LogActivity     time.Time
}

// MassifTail contains the massif specific tail information
type MassifTail struct {
	LogTailActivity
	FirstIndex uint64
}

// SealTail contains the seal specific tail information
type SealTail struct {
	LogTailActivity
	Count  uint64
	Signed cose.CoseSign1Message
	State  massifs.MMRState
}

// String returns a printable pretty rendering of the tail
func (st SealTail) String() string {

	s := fmt.Sprintf(
		"seal: %d, mmrSize: %d, lastid: %s, seal time: %v, log activity: %v",
		st.Number, st.State.MMRSize, st.LastID,
		time.UnixMilli(st.State.Timestamp).UTC().Format(time.RFC3339),
		st.LogActivity.UTC().Format(time.RFC3339),
	)
	return fmt.Sprintf(
		"%s, log activity: %v",
		s, st.LogActivity.UTC().Format(time.RFC3339),
	)
}

// NewTailConfig derives a configuration from the supplied command line options context
func NewTailConfig(cCtx *cli.Context, cmd *CmdCtx) (TailConfig, error) {

	if cCtx.String("logid") == "" {
		return TailConfig{}, fmt.Errorf("a logid is required")
	}

	cfg := TailConfig{}
	// note: the cli defaults to 1 second interval and count = 1. so by default
	// the interval is ignored. If count is 0 or > 1, we get a single second
	// sleep by default.
	cfg.Interval = cCtx.Duration("interval")

	// transitional allow regular tenant id's from the datatrails era
	cfg.LogID = storage.ParsePrefixedLogID("tenant/", cCtx.String("logid"))
	if cfg.LogID == nil {
		var err error
		uid, err := uuid.Parse(cCtx.String("logid"))
		if err != nil {
			return TailConfig{}, err
		}
		cfg.LogID = uid[:]
	}
	return cfg, nil
}

// String returns a printable pretty rendering of the tail
func (lt MassifTail) String() string {

	s := fmt.Sprintf(
		"massif: %d, lastid: %s, log activity: %v",
		lt.Number, lt.LastID,
		lt.LogActivity.UTC().Format(time.RFC3339),
	)
	return fmt.Sprintf(
		"%s, log activity: %v",
		s, lt.LogActivity.UTC().Format(time.RFC3339),
	)
}

// TailSeal returns the most recently added seal for the log
func TailSeal(
	ctx context.Context,
	reader massifs.ObjectReader,
	codec commoncbor.CBORCodec,
	logID storage.LogID,
) (SealTail, error) {
	var err error

	st := SealTail{
		LogTailActivity: LogTailActivity{
			LogTail: watcher.LogTail{
				LogID: logID,
			},
		},
	}

	headIndex, err := reader.HeadIndex(ctx, storage.ObjectCheckpoint)
	if err != nil {
		return SealTail{}, fmt.Errorf("error reading head massif index: %w", err)
	}

	checkpt, err := massifs.GetCheckpoint(ctx, reader, codec, headIndex)
	if err != nil {
		return SealTail{}, fmt.Errorf("error reading checkpoint for tenant %x: %w", logID, err)
	}
	st.Signed = checkpt.Sign1Message
	st.State = checkpt.MMRState

	// The log activity as it stood when the seal was made is on the state
	lastMS, err := snowflakeid.IDUnixMilli(st.State.IDTimestamp, uint8(st.State.CommitmentEpoch))
	if err != nil {
		return SealTail{}, err
	}

	st.LogActivity = time.UnixMilli(lastMS)
	st.LastIDTimestamp = st.State.IDTimestamp

	return st, err
}

type storageTailer interface {
	HeadIndex(ctx context.Context, ty storage.ObjectType) (uint32, error)
}

type massifTailer interface {
	storageTailer
}

// TailMassif returns the active massif for the tenant
func TailMassif(
	ctx context.Context,
	reader massifs.ObjectReader,
	logID storage.LogID,
) (MassifTail, error) {
	var err error
	lt := MassifTail{
		LogTailActivity: LogTailActivity{
			LogTail: watcher.LogTail{
				LogID: logID,
			},
		},
	}

	headIndex, err := reader.HeadIndex(ctx, storage.ObjectMassifStart)
	if err != nil {
		return MassifTail{}, fmt.Errorf("error reading head massif index: %w", err)
	}

	start, err := massifs.GetMassifStart(ctx, reader, headIndex)

	lt.Number = headIndex
	lt.OType = storage.ObjectMassifData

	logActivityMS, err := snowflakeid.IDUnixMilli(start.LastID, uint8(start.CommitmentEpoch))
	if err != nil {
		return MassifTail{}, fmt.Errorf(
			"error reading last activity time from head massif for log %x: %w",
			logID, err)
	}
	lt.LogActivity = time.UnixMilli(logActivityMS)

	lt.LastIDTimestamp = start.LastID
	lt.LastIDEpoch = uint8(start.CommitmentEpoch)

	return lt, nil
}

func NewLogTailCmd() *cli.Command {
	return &cli.Command{Name: "tail",
		Usage: `report the current tail (most recent end) of the log

		if --count is > 1, re-check every interval seconds until the count is exhausted
		if --count is explicitly zero, check forever
		`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "mode",
				Usage: "Any of [massif, seal, both], defaults to both",
				Value: "massif",
			},

			&cli.StringFlag{
				Name:     "logid",
				Usage:    "log identifier, as a string encoded uuid",
				Required: true,
			},

			&cli.DurationFlag{
				Name: "interval", Aliases: []string{"d"},
				Value: time.Second,
				Usage: "The default polling interval is one second, setting the interval to zero disables polling",
			},
			&cli.IntFlag{
				Name: "count", Usage: "Number of intervals. Zero means forever. Defaults to single poll",
				Value: 1,
			},
		},
		Action: func(cCtx *cli.Context) error {
			var err error
			cmd := &CmdCtx{}
			ctx := context.Background()

			if err = cfgMassifFmt(cmd, cCtx); err != nil {
				return err
			}

			reader, err := newMassifReader(cmd, cCtx)
			if err != nil {
				return err
			}

			cfg, err := NewTailConfig(cCtx, cmd)
			if err != nil {
				return err
			}

			codec, err := massifs.NewCBORCodec()
			if err != nil {
				return err
			}

			count := cCtx.Int("count")
			mode := cCtx.String("mode")
			for {

				var lt MassifTail
				var st SealTail
				if mode == "both" || mode == "massif" {
					lt, err = TailMassif(ctx, reader, cfg.LogID)
					if err != nil {
						return err
					}
					fmt.Printf("%s\n", lt.String())
				}
				if mode == "both" || mode == "seal" {
					st, err = TailSeal(ctx, reader, codec, cfg.LogID)
					if err != nil {
						return err
					}
					fmt.Printf("%s\n", st.String())
				}

				// Note we don't allow a zero interval
				if count == 1 || cfg.Interval == 0 {
					break
				}
				// count == 0 is infinite
				if count > 1 {
					count--
				}
				time.Sleep(cfg.Interval)
			}
			return nil
		},
	}
}
