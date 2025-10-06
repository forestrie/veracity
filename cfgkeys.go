package veracity

import (
	"fmt"

	"github.com/forestrie/go-merklelog/massifs"
	"github.com/datatrails/veracity/keyio"
	"github.com/urfave/cli/v2"
)

func CfgKeys(cmd *CmdCtx, cCtx *cli.Context) error {

	var err error

	cmd.CBORCodec, err = massifs.NewCBORCodec()
	if err != nil {
		return fmt.Errorf("failed to create CBOR codec: %w", err)
	}

	if cCtx.IsSet("checkpoint-public") && cCtx.IsSet("checkpoint-public-pem") {
		return fmt.Errorf("cannot set both checkpoint-public and checkpoint-public-pem, use only one")
	}

	if cCtx.IsSet("checkpoint-public") {
		pub, err := keyio.ReadECDSAPublicCOSE(cCtx.String("checkpoint-public"))
		if err != nil {
			return fmt.Errorf("failed to read checkpoint public key: %w", err)
		}
		cmd.CheckpointPublic = pub
		return nil
	}
	if cCtx.IsSet("checkpoint-public-pem") {
		pub, err := keyio.ReadECDSAPublicPEM(cCtx.String("checkpoint-public-pem"))
		if err != nil {
			return fmt.Errorf("failed to read checkpoint public key: %w", err)
		}
		cmd.CheckpointPublic = pub
		return nil
	}

	if cCtx.IsSet("checkpoint-jwks") {
		pub, err := keyio.ReadECDSAPublicJOSE(cCtx.String("checkpoint-jwks"))
		if err != nil {
			return fmt.Errorf("failed to read checkpoint public key: %w", err)
		}
		cmd.CheckpointPublic = pub
		return nil
	}

	return nil
}
