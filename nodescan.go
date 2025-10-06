package veracity

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/forestrie/go-merklelog/massifs"
	"github.com/urfave/cli/v2"
)

// NewNodeScan implements a sub command which linearly scans for a node in a blob
// This is a debugging tool
func NewNodeScanCmd() *cli.Command {
	return &cli.Command{Name: "nodescan",
		Usage: "scan a log for a particular node value. this is a debugging tool",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name: "massif", Aliases: []string{"m"},
			},
			&cli.StringFlag{
				Name: "value", Aliases: []string{"v"},
			},
			&cli.BoolFlag{Name: "massif-relative", Aliases: []string{"r"}},
		},
		Action: func(cCtx *cli.Context) error {
			cmd := &CmdCtx{}

			var err error
			var massif *massifs.MassifContext
			if massif, err = cfgMassif(context.Background(), cmd, cCtx); err != nil {
				return err
			}

			targetValue, err := hex.DecodeString(cCtx.String("value"))
			if err != nil {
				return err
			}
			start := massif.LogStart()
			count := massif.Count()
			for i := range count {
				entry := massif.Data[start+i*massifs.ValueBytes : start+i*massifs.ValueBytes+massifs.ValueBytes]
				if bytes.Equal(entry, targetValue) {
					fmt.Printf("%d\n", i+massif.Start.FirstIndex)
					return nil
				}
			}
			return fmt.Errorf("'%s' not found", cCtx.String("value"))
		},
	}
}
