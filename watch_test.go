package veracity

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type checkWatchConfig func(t *testing.T, cfg WatchConfig)

func TestNewWatchConfig(t *testing.T) {

	hourSince := time.Now().Add(-time.Hour)

	type args struct {
		cCtx *mockContext
		cmd  *CmdCtx
	}
	tests := []struct {
		name      string
		args      args
		want      *WatchConfig
		check     checkWatchConfig
		errPrefix string
	}{

		{
			name: "max horizon alias",
			args: args{
				cCtx: &mockContext{
					horizon: "max",
				},
				cmd: new(CmdCtx),
			},
		},

		{
			name: "latest flag for casual replicators",
			args: args{
				cCtx: &mockContext{
					latest: true,
				},
				cmd: new(CmdCtx),
			},
		},

		{
			name: "latest mutualy exclusive with horizon",
			args: args{
				cCtx: &mockContext{
					latest:  true,
					horizon: "1h",
				},
				cmd: new(CmdCtx),
			},
			errPrefix: "the latest flag is mutually exclusive",
		},

		{
			name: "interval too small",
			args: args{
				cCtx: &mockContext{
					horizon: "1h",
					// just under a second
					interval: time.Millisecond * 999,
				},
				cmd: new(CmdCtx),
			},
			errPrefix: "polling more than once per second is not",
		},

		{
			name: "horizon or since options are required",
			args: args{
				cCtx: &mockContext{},
				cmd:  new(CmdCtx),
			},
			errPrefix: "provide the latest flag, horizon on its own or either of the since parameters",
		},

		{
			name: "poll count is at least one",
			args: args{
				cCtx: &mockContext{
					since: &hourSince,
				},
				cmd: new(CmdCtx),
			},
			check: func(t *testing.T, cfg WatchConfig) {
				assert.Equal(t, 1, cfg.WatchCount)
			},
		},

		{
			name: "poll count is capped",
			args: args{
				cCtx: &mockContext{
					since: &hourSince,
					count: 100,
				},
				cmd: new(CmdCtx),
			},
			check: func(t *testing.T, cfg WatchConfig) {
				assert.Equal(t, maxPollCount, cfg.WatchCount)
			},
		},
		{
			name: "poll with since an hour in the past",
			args: args{
				cCtx: &mockContext{
					since: &hourSince,
				},
				cmd: new(CmdCtx),
			},
			check: func(t *testing.T, cfg WatchConfig) {
				assert.Equal(t, hourSince, cfg.Since)
				assert.Equal(t, time.Second*3, cfg.Interval)
				assert.NotEqual(t, "", cfg.IDSince) // should be set to IDTimeHex
			},
		},
		{
			name: "bad hex string for idtimestamp errors",
			args: args{
				cCtx: &mockContext{
					idsince: "thisisnothex",
				},
				cmd: new(CmdCtx),
			},
			errPrefix: "encoding/hex: invalid byte",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWatchConfig(tt.args.cCtx, tt.args.cmd)
			if err != nil {
				if tt.errPrefix == "" {
					t.Errorf("NewWatchConfig() unexpected error = %v", err)
				}
				if !strings.HasPrefix(err.Error(), tt.errPrefix) {
					t.Errorf("NewWatchConfig() unexpected error = %v, expected prefix: %s", err, tt.errPrefix)
				}
			} else {
				if tt.errPrefix != "" {
					t.Errorf("NewWatchConfig() expected error prefix = %s", tt.errPrefix)
				}
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWatchConfig() = %v, want %v", got, tt.want)
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

type mockContext struct {
	since    *time.Time
	latest   bool
	mode     string
	idsince  string
	horizon  string
	interval time.Duration
	count    int
	tenant   string
}

func (c mockContext) IsSet(n string) bool {
	switch n {
	case "since":
		return c.since != nil
	case "latest":
		return c.latest == true
	case "mode":
		return c.mode != ""
	case "idsince":
		return c.idsince != ""
	case "horizon":
		return c.horizon != ""
	case "interval":
		return c.interval != 0
	case "count":
		return c.count != 0
	case "tenant":
		return c.tenant != ""
	default:
		return false
	}
}

func (c mockContext) String(n string) string {
	switch n {
	case "mode":
		return c.mode
	case "idsince":
		return c.idsince
	case "tenant":
		return c.tenant
	case "horizon":
		return c.horizon
	default:
		return ""
	}
}

func (c mockContext) Bool(n string) bool {
	switch n {
	case "latest":
		return c.latest
	default:
		return false
	}
}

func (c mockContext) Int(n string) int {
	switch n {
	case "count":
		return c.count
	default:
		return 0
	}
}

func (c mockContext) Duration(n string) time.Duration {
	switch n {
	case "interval":
		return c.interval
	default:
		return time.Duration(0)
	}
}

func (c mockContext) Timestamp(n string) *time.Time {
	switch n {
	case "since":
		return c.since
	default:
		return nil
	}
}
