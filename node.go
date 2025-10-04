package veracity

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"
)

// NewNodeCmd prints out the identified mmr node
func NewNodeCmd() *cli.Command {
	return &cli.Command{Name: "node",
		Usage: `read a merklelog node

			provide --mmrindex or -i to specify the node to read
		`,
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name: "mmrindex", Aliases: []string{"i"},
			},
		},
		Action: func(cCtx *cli.Context) error {
			cmd := &CmdCtx{}
			massif, err := cfgMassif(context.Background(), cmd, cCtx)
			if err != nil {
				return err
			}

			mmrIndex := cCtx.Uint64("mmrindex")

			value, err := massif.Get(mmrIndex)
			if err != nil {
				return err
			}
			fmt.Printf("%x\n", value)
			return nil
		},
	}
}
