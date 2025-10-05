package veracity

import (
	"context"

	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	fsstorage "github.com/robinbryce/go-merklelog-fs/storage"
	"github.com/urfave/cli/v2"
	"github.com/veraison/go-cose"
)

const (
	defaultMassifHeight = uint8(14)
)

func NewCmdStorageProviderFS(
	ctx context.Context,
	cCtx *cli.Context, cmd *CmdCtx,
	dataLocal string,
	createRootDir bool,
) (*fsstorage.CachingStore, error) {
	var err error
	massifExt := storage.V1MMRExtSep + storage.V1MMRMassifExt
	if cCtx.IsSet("massif-ext") {
		massifExt = cCtx.String("massif-ext")
	}

	opts := fsstorage.Options{
		FSOptions: fsstorage.FSOptions{
			MassifFile:      cCtx.String("massif-file"),
			CheckpointFile:  cCtx.String("checkpoint-file"),
			RootDir:         dataLocal,
			CreateRootDir:  createRootDir,
			MassifExtension: massifExt,
		},
	}

	opts.MassifHeight = cmd.MassifFmt.MassifHeight

	if cmd.CheckpointPublic.Public != nil {
		verifier, err := cose.NewVerifier(cmd.CheckpointPublic.Alg, cmd.CheckpointPublic.Public)
		if err != nil {
			return nil, err
		}
		opts.COSEVerifier = verifier
	}

	// Create Filesystem ObjectStore (replaces MassifFinder)
	store, err := fsstorage.NewStore(ctx, opts)
	if err != nil {
		return nil, err
	}
	return store, nil
}
