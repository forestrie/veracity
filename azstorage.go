package veracity

import (
	"context"
	"fmt"

	"github.com/datatrails/go-datatrails-common/azblob"
	azstorage "github.com/robinbryce/go-merklelog-azure/storage"
	"github.com/urfave/cli/v2"
)

func IsStorageEmulatorEnabled(cCtx *cli.Context) bool {
	return cCtx.String("account") == AzuriteStorageAccount
}

func NewCmdStorageProviderAzure(
	ctx context.Context,
	cCtx *cli.Context, cmd *CmdCtx,
	dataUrl string,
	reader azblob.Reader,
) (*azstorage.CachingStore, error) {

	var err error

	if reader == nil {

		// If we had no url and no local data supplied we default to the production data source.
		reader, err = cfgReader(cmd, cCtx, dataUrl)
		if err != nil {
			return nil, err
		}
	}

	/*
		logID := datatrails.TenantID2LogID(dataUrl)
		if logID == nil {
			logID = datatrails.TenantID2LogID(cCtx.String("tenant"))
			// if logID == nil {
			// 	return nil, fmt.Errorf("tenant must be provided for this command (todo: remove this requirement)")
			// }
		}*/
	opts := azstorage.Options{}
	opts.Store = reader

	store, err := azstorage.NewStore(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("could not create Azure object store: %w", err)
	}
	return store, nil
}
