package veracity

import (
	"context"
	"fmt"

	"github.com/forestrie/go-merklelog/massifs"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
	"github.com/urfave/cli/v2"
)

// cfgMassif configures a massif reader and reads a massif
func cfgMassif(ctx context.Context, cmd *CmdCtx, cCtx *cli.Context) (*massifs.MassifContext, error) {

	var massif massifs.MassifContext
	var reader readerSelector
	var err error

	if reader, err = cfgMassifReader(cmd, cCtx); err != nil {
		return nil, err
	}

	tenant := CtxGetOneTenantOption(cCtx)
	if tenant == "" {
		return nil, fmt.Errorf("tenant must be provided for this command")
	}

	logID := datatrails.TenantID2LogID(tenant)
	if logID == nil {
		return nil, fmt.Errorf("invalid tenant id '%s'", tenant)
	}
	
	if err = reader.SelectLog(ctx, logID); err != nil {
		return nil, err
	}

	mmrIndex := cCtx.Uint64("mmrindex")
	massifIndex := uint32(cCtx.Uint64("massif"))

	// mmrIndex zero is always going to be massifIndex 0 so we treat this the
	// same as though the massif option had been supplied as 0
	if massifIndex == uint32(0) && mmrIndex == uint64(0) {
		massif, err = massifs.GetMassifContext(ctx, reader, massifIndex)
		if err != nil {
			return nil, err
		}
		return &massif, nil
	}

	// now, if we have a non zero mmrIndex, use it to (re)compute the massifIndex
	if mmrIndex > uint64(0) {
		massifIndex = uint32(massifs.MassifIndexFromMMRIndex(cmd.MassifFmt.MassifHeight, mmrIndex))

		massif, err = massifs.GetMassifContext(ctx, reader, massifIndex)
		if err != nil {
			return nil, err
		}
		return &massif, nil
	}

	// If massifIndex is not provided it will be zero here, and that is a good
	// default.
	massif, err = massifs.GetMassifContext(ctx, reader, massifIndex)
	if err != nil {
		return nil, err
	}
	return &massif, nil
}
