package veracity

import (
	"context"
	"fmt"

	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	"github.com/urfave/cli/v2"
)

// omniMassifReader is the union of all interfaces needed by veracity commands
type omniMassifReader interface {
	SelectLog(ctx context.Context, logId storage.LogID) error
	massifs.ObjectReader
	massifs.ObjectWriter
}

type readerSelector interface {
	SelectLog(ctx context.Context, logId storage.LogID) error
	massifs.ObjectReader
}

func newReaderSelector(cmd *CmdCtx, cCtx *cli.Context) (readerSelector, error) {
	return newMassifStore(cmd, cCtx)
}

func newMassifReader(cmd *CmdCtx, cCtx *cli.Context) (readerSelector, error) {
	return newMassifStore(cmd, cCtx)
}

func localDataOptionsSet(cCtx *cli.Context) bool {
	if cCtx.IsSet("data-local") && cCtx.String("data-local") != "" {
		return true
	}
	if cCtx.IsSet("massif-file") && cCtx.String("massif-file") != "" {
		return true
	}
	if cCtx.IsSet("checkpoint-file") && cCtx.String("checkpoint-file") != "" {
		return true
	}
	return false
}

func newMassifStore(cmd *CmdCtx, cCtx *cli.Context) (omniMassifReader, error) {
	var err error

	localSet := localDataOptionsSet(cCtx)
	remoteLog := cCtx.String("data-url")

	if !localSet && remoteLog != "" {
		return nil, fmt.Errorf("can't use data-local and data-url at the same time")
	}

	if !localSet && remoteLog == "" && !IsStorageEmulatorEnabled(cCtx) {
		remoteLog = DefaultRemoteMassifURL
	}

	var reader omniMassifReader

	if remoteLog != "" || IsStorageEmulatorEnabled(cCtx) {
		reader, err = NewCmdStorageProviderAzure(context.Background(), cCtx, cmd, remoteLog, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create massif reader: %w", err)
		}
		return reader, nil
	}
	if localSet {

		reader, err := NewCmdStorageProviderFS(context.Background(), cCtx, cmd, cCtx.String("data-local"), false)
		if err != nil {
			return nil, fmt.Errorf("could not create massif reader: %w", err)
		}
		return reader, nil
	}
	return nil, fmt.Errorf("no massif reader configured, use either data-local or data-url or both")
}
