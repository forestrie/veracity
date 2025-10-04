package veracity

import (
	"github.com/urfave/cli/v2"
)

// cfgMassifReader establishes the blob read only data accessor
// only azure blob storage is supported. Both emulated and production.
func cfgMassifReader(cmd *CmdCtx, cCtx *cli.Context) (readerSelector, error) {
	var err error
	if cmd.Log == nil {
		if err = cfgLogging(cmd, cCtx); err != nil {
			return nil, err
		}
	}
	if err = cfgMassifFmt(cmd, cCtx); err != nil {
		return nil, err
	}

	reader, err := newMassifReader(cmd, cCtx)
	if err != nil {
		return nil, err
	}

	return reader, nil
}
