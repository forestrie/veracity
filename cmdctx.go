package veracity

import (
	"fmt"

	"github.com/forestrie/go-merklelog/massifs/cbor"
	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/forestrie/go-merklelog/massifs/snowflakeid"
	"github.com/datatrails/veracity/keyio"
)

// CmdCtx holds shared config and config derived state for all commands
type CmdCtx struct {
	Log logger.Logger

	CheckpointPublic keyio.DecodedPublic

	RemoteURL string
	CBORCodec cbor.CBORCodec

	// cfgMassifFmt sets the massif format options and the IDState
	MassifFmt MassifFormatOptions
	IDState   *snowflakeid.IDState

	Bugs map[string]bool
}

func (cmd *CmdCtx) NextID() (uint64, error) {
	if cmd.IDState == nil {
		return 0, fmt.Errorf("idState not initialized, cannot generate next ID")
	}
	return cmd.IDState.NextID()
}

// Clone returns a safe copy of the CmdCtx.
func (c *CmdCtx) Clone() *CmdCtx {
	return &CmdCtx{
		RemoteURL: c.RemoteURL,
		CBORCodec: c.CBORCodec,
		MassifFmt: c.MassifFmt,
		IDState:   c.IDState,
		Log:       c.Log,
	}
}
