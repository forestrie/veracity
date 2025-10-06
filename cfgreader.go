package veracity

import (
	"fmt"

	"github.com/forestrie/go-merklelog-azure/blobs"
	"github.com/urfave/cli/v2"
)

// cfgReader establishes the blob read only data accessor
// only azure blob storage is supported. Both emulated and production.
func cfgReader(cmd *CmdCtx, cCtx *cli.Context, url string) (blobs.Reader, error) {
	var err error
	var reader blobs.Reader

	if cmd.Log == nil {
		if err = cfgLogging(cmd, cCtx); err != nil {
			return nil, err
		}
	}
	opts := blobs.Options{
		Container: cCtx.String("container"),
		Account:   cCtx.String("account"),
		EnvAuth:   cCtx.Bool("envauth"),
	}
	reader, cmd.RemoteURL, err = blobs.NewBlobReader(cmd.Log, url, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blob store: %v", err)
	}

	return reader, nil
}
