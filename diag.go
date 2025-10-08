package veracity

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/forestrie/go-merklelog/massifs"
	"github.com/forestrie/go-merklelog/massifs/snowflakeid"
	"github.com/forestrie/go-merklelog/mmr"
	"github.com/urfave/cli/v2"
)

// NewDiagCmd prints diagnostic information about the massif blob containg a
// specific mmrIndex
func NewDiagCmd() *cli.Command {
	return &cli.Command{Name: "diag",
		Usage: "print diagnostics about a blob identified by massif index or by an mmrindex",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name: "mmrindex", Aliases: []string{"i"},
			},
			&cli.Int64Flag{
				Name: "massif", Aliases: []string{"m"},
				Usage: "allow inspection of an arbitrary mmr index by explicitly specifying a massif index",
				Value: -1,
			},
		},
		Action: func(cCtx *cli.Context) error {
			var err error
			var massif massifs.MassifContext
			var reader massifs.ObjectReader

			ctx := context.Background()

			cmd := &CmdCtx{}
			if reader, err = cfgMassifReader(cmd, cCtx); err != nil {
				return err
			}
			if cmd.MassifFmt.MassifHeight == 0 {
				return fmt.Errorf("massif height can't be zero")
			}

			fmt.Printf("%8d trie-header-start\n", massifs.TrieHeaderStart())
			fmt.Printf("%8d trie-data-start\n", massifs.TrieDataStart())
			fmt.Printf("%8d peak-stack-start\n", massifs.PeakStackStart(cmd.MassifFmt.MassifHeight))

			// support identifying the massif implicitly via the index of a log
			// entry. note that mmrIndex 0 is just handled as though the caller
			// asked for massifIndex 0
			mmrIndex := cCtx.Uint64("mmrindex")
			signedMassifIndex := cCtx.Int64("massif")
			massifIndex := uint32(signedMassifIndex)
			if mmrIndex > uint64(0) && signedMassifIndex == -1 {
				massifIndex = uint32(massifs.MassifIndexFromMMRIndex(cmd.MassifFmt.MassifHeight, mmrIndex))
			}
			fmt.Printf("%8d peak-stack-len\n", massifs.PeakStackLen(uint64(massifIndex)))
			var logStart uint64
			switch massif.Start.Version {
			case 1:
				logStart = massifs.PeakStackEnd(cmd.MassifFmt.MassifHeight)
			case 0:
				logStart = massifs.PeakStackEndV0(uint64(massifIndex), cmd.MassifFmt.MassifHeight)
			}
			fmt.Printf("%8d tree-start\n", logStart)
			fmt.Printf("%8d massif\n", massifIndex)
			if mmrIndex > 0 {
				fmt.Printf("%8d mmrindex\n", mmrIndex)
			}
			if cCtx.Bool("noread") {
				return nil
			}
			tenant := cCtx.String("tenant")
			if tenant == "" && !cCtx.IsSet("data-local") {
				fmt.Println("a tenant is required to get diagnostics that require reading a blob")
				return nil
			}
			massif, err = massifs.GetMassifContext(ctx, reader, massifIndex)
			if err != nil {
				return err
			}
			fmt.Printf("%8d start:massif-height\n", massif.Start.MassifHeight)
			fmt.Printf("%8d start:data-epoch\n", massif.Start.DataEpoch)
			fmt.Printf("%8d start:commitment-epoch\n", massif.Start.CommitmentEpoch)
			fmt.Printf("%8d start:first-index\n", massif.Start.FirstIndex)
			fmt.Printf("%8d start:peak-stack-len\n", massif.Start.PeakStackLen)

			fmt.Printf("%8d count\n", massif.Count())
			fmt.Printf("%8d leaf-count\n", massif.MassifLeafCount())
			fmt.Printf("%8d last-leaf-mmrindex\n", massif.LastLeafMMRIndex())

			// trieIndex is equivilent to leafIndex, but we use the term trieIndex
			//  when dealing with trie data.
			trieIndex := mmr.LeafIndex(mmrIndex)
			fmt.Printf("%8d trie-index\n", trieIndex)

			// FirstIndex is the *size* of the mmr preceding the current massif
			expectTrieIndexMassif := trieIndex - mmr.LeafCount(massif.Start.FirstIndex)
			fmt.Printf("%8d trie-index - massif-first-index\n", expectTrieIndexMassif)

			logTrieKey, err := massif.GetTrieKey(mmrIndex)
			if err != nil {
				return fmt.Errorf("when expecting %d for %d: %v", expectTrieIndexMassif, mmrIndex, err)
			}
			logTrieEntry, err := massif.GetTrieEntry(mmrIndex)
			if err != nil {
				entryIndex := mmr.LeafIndex(mmrIndex)
				// FirstIndex is the *size* of the mmr preceding the current massif
				expectTrieIndexMassif := entryIndex - mmr.LeafCount(massif.Start.FirstIndex)
				return fmt.Errorf("when expecting %d for %d: %v", expectTrieIndexMassif, mmrIndex, err)
			}

			logNodeValue, err := massif.Get(mmrIndex)
			if err != nil {
				return err
			}
			fmt.Printf("%x log-value\n", logNodeValue)

			idBytes := logTrieKey[massifs.TrieEntryIDTimestampStart:massifs.TrieEntryIDTimestampEnd]
			id := binary.BigEndian.Uint64(idBytes)
			unixMS, err := snowflakeid.IDUnixMilli(id, uint8(massif.Start.CommitmentEpoch))
			if err != nil {
				return err
			}
			idTime := time.UnixMilli(unixMS)
			fmt.Printf("%x log-trie-key\n", logTrieKey[:32])
			fmt.Printf("%x %s\n", logTrieKey[32:], idTime.Format(time.DateTime))
			fmt.Printf("%x log-trie-entry\n", logTrieEntry[:64])

			return nil
		},
	}
}
