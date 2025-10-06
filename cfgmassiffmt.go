package veracity

import (
	"github.com/forestrie/go-merklelog/massifs"
	"github.com/forestrie/go-merklelog/massifs/snowflakeid"
	"github.com/urfave/cli/v2"
)

type MassifFormatOptions struct {
	MassifHeight    uint8
	CommitmentEpoch uint8
	WorkerCIDR      string
	PodIP           string
}

// cfgMassifFmt initialises an idTimestamp generator
// of enabled workarounds.
func cfgMassifFmt(cmd *CmdCtx, cCtx *cli.Context) error {
	var err error
	cmd.MassifFmt.CommitmentEpoch = 1 // the default, and correct until the unix epoch changes
	if cCtx.IsSet("commitment-epoch") {
		cmd.MassifFmt.CommitmentEpoch = uint8(cCtx.Uint64("commitment-epoch"))
		if cmd.MassifFmt.CommitmentEpoch == 0 {
			cmd.MassifFmt.CommitmentEpoch = uint8(massifs.Epoch2038)
		}
	}
	cmd.MassifFmt.MassifHeight = uint8(cCtx.Uint("height"))
	if cmd.MassifFmt.MassifHeight == 0 {
		cmd.MassifFmt.MassifHeight = defaultMassifHeight
	}

	cmd.MassifFmt.WorkerCIDR = "0.0.0.0/16"
	if cCtx.IsSet("worker-cidr") {
		cmd.MassifFmt.WorkerCIDR = cCtx.String("worker-cidr")
	}
	cmd.MassifFmt.PodIP = "10.0.0.127"
	if cCtx.IsSet("pod-ip") {
		cmd.MassifFmt.PodIP = cCtx.String("pod-ip")
	}

	cmd.IDState, err = snowflakeid.NewIDState(snowflakeid.Config{
		CommitmentEpoch: cmd.MassifFmt.CommitmentEpoch,
		// There is no reason to override these for local use.
		WorkerCIDR: cmd.MassifFmt.WorkerCIDR,
		PodIP:      cmd.MassifFmt.PodIP,
	})

	return err
}
