package veracity

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/datatrails/go-datatrails-merklelog/massifs/snowflakeid"
	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	"github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/datatrails/go-datatrails-simplehash/simplehash"
	appdata "github.com/forestrie/go-merklelog-datatrails/appdata"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
	"github.com/urfave/cli/v2"
)

// NewEventDiagCmd provides diagnostic support for event verification
//
//nolint:gocognit
func NewEventDiagCmd() *cli.Command {
	return &cli.Command{Name: "event-log-info",
		Aliases: []string{"ediag"},
		Usage:   "print diagnostics about an events entry in the log",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "bendump", Aliases: []string{"b"}, Value: false},
			&cli.StringFlag{
				Name: "json", Aliases: []string{"j"},
			},
		},
		Action: func(cCtx *cli.Context) error {

			var reader readerSelector
			var massif massifs.MassifContext

			tenantIdentity := cCtx.String("tenant")

			appData, err := appdata.ReadAppData(cCtx.Args().Len() == 0, cCtx.Args().Get(0))
			if err != nil {
				return err
			}

			appEntries, err := appdata.AppDataToVerifiableLogEntries(appData, tenantIdentity)
			if err != nil {
				return err
			}

			cmd := &CmdCtx{}

			if err = cfgMassifFmt(cmd, cCtx); err != nil {
				return err
			}

			reader, err = newReaderSelector(cmd, cCtx)
			if err != nil {
				return err
			}
			cmpPrint := func(fmtEq, fmtNe string, a, b any) bool {
				if reflect.DeepEqual(a, b) {
					fmt.Printf(fmtEq, a)
					return true
				}
				fmt.Printf(fmtNe, a, b)
				return false
			}

			for _, appEntry := range appEntries {

				if appEntry.MMRIndex() == 0 {
					continue
				}

				// Get the mmrIndex from the request and then compute the massif
				// it implies based on the massifHeight command line option.
				mmrIndex := appEntry.MMRIndex()
				massifIndex := uint32(massifs.MassifIndexFromMMRIndex(cmd.MassifFmt.MassifHeight, mmrIndex))
				var logID storage.LogID
				if cCtx.String("tenant") != "" {
					tenantIdentity := cCtx.String("tenant")
					logID = datatrails.TenantID2LogID(tenantIdentity)
				}
				if logID == nil {

					// The tenant identity on the event is the original tenant
					// that created the event. For public assets and shared
					// assets, this is true regardless of which tenancy the
					// record is fetched from.  Those same events will appear in
					// the logs of all tenants they were shared with.
					logID = appEntry.LogID()
				}
				reader.SelectLog(cCtx.Context, logID)
				// read the massif blob
				massif, err = massifs.GetMassifContext(context.Background(), reader, massifIndex)
				if err != nil {
					return err
				}

				// Get the human time from the idtimestamp committed on the event.
				idTimestamp, err := appEntry.IDTimestamp(&massif)
				if err != nil {
					return err
				}

				idTimestampWithEpoch := make([]byte, len(idTimestamp)+1)
				idTimestampWithEpoch[0] = byte(0) // 0 epoch
				copy(idTimestampWithEpoch[1:], idTimestamp)

				eventIDTimestamp, _, err := massifs.SplitIDTimestampBytes(idTimestampWithEpoch)
				if err != nil {
					return err
				}
				eventIDTimestampMS, err := snowflakeid.IDUnixMilli(eventIDTimestamp, uint8(massif.Start.CommitmentEpoch))
				if err != nil {
					return err
				}

				leafIndex := mmr.LeafIndex(mmrIndex)
				// Note that the banner info is all from the event response
				fmt.Printf("%d %s %s\n", leafIndex, time.UnixMilli(eventIDTimestampMS).Format(time.RFC3339Nano), appEntry.AppID())

				leafIndexMassif, err := massif.GetMassifLeafIndex(leafIndex)
				if err != nil {
					return fmt.Errorf("when expecting %d for %d: %v", leafIndexMassif, mmrIndex, err)
				}
				fmt.Printf(" |%8d leaf-index-massif\n", leafIndexMassif)

				// Read the trie entry from the log
				logTrieEntry := massifs.GetTrieEntry(massif.Data, massif.IndexStart(), leafIndexMassif)
				logNodeValue, err := massif.Get(mmrIndex)
				if err != nil {
					return err
				}

				logTrieKey := massifs.GetTrieKey(massif.Data, massif.IndexStart(), leafIndexMassif)

				logTrieIDTimestampBytes := logTrieEntry[massifs.TrieEntryIDTimestampStart:massifs.TrieEntryIDTimestampEnd]
				logTrieIDTimestamp := binary.BigEndian.Uint64(logTrieIDTimestampBytes)
				unixMS, err := snowflakeid.IDUnixMilli(logTrieIDTimestamp, uint8(massif.Start.CommitmentEpoch))
				if err != nil {
					return err
				}
				idTime := time.UnixMilli(unixMS)

				// TODO: log version 1 uses the uuid bytes of the log tenant
				// so we need to handle that here as well as log version 0.
				trieKey := massifs.NewTrieKey(
					massifs.KeyTypeApplicationContent,
					[]byte(tenantIdentity),
					[]byte(strings.TrimPrefix(appEntry.AppID(), "public")))
				if len(trieKey) != massifs.TrieKeyBytes {
					return massifs.ErrIndexEntryBadSize
				}
				cmpPrint(
					" |%x trie-key\n",
					" |%x != log-trie-key %x\n", trieKey[:32], logTrieKey[:32])
				fmt.Printf(" |%x %s log-idtimestamp\n", logTrieIDTimestampBytes, idTime.Format(time.DateTime))
				cmpPrint(
					" |%x idtimestamp\n",
					" |%x != log-idtimestamp %x\n", eventIDTimestamp, logTrieIDTimestamp)

				// Compute the event data hash, independent of domain and idtimestamp
				v3Event, err := simplehash.V3FromEventJSON(appEntry.SerializedBytes()) // NOTE for assetsv2 the serialized bytes is actually the event json API response
				if err != nil {
					return err
				}

				eventHasher := sha256.New()
				if err = simplehash.V3HashEvent(eventHasher, v3Event); err != nil {
					return err
				}
				eventHash := eventHasher.Sum(nil)
				fmt.Printf(" |%x v3hash (just the schema fields hashed)\n", eventHash)
				if cCtx.Bool("bendump") {
					bencode, err2 := bencodeEvent(v3Event)
					if err2 != nil {
						return err2
					}
					fmt.Printf(" |%s\n", string(bencode))
				}

				leafHasher := simplehash.NewHasherV3()
				err = leafHasher.HashEventFromV3(
					v3Event,
					simplehash.WithPrefix([]byte{LeafTypePlain}),
					simplehash.WithIDCommitted(eventIDTimestamp))
				if err != nil {
					return err
				}
				leafHash := leafHasher.Sum(nil)

				ok := cmpPrint(
					" |%x leaf\n",
					" |%x leaf != log-leaf %x\n", leafHash, logNodeValue)
				if !ok {
					// if the leaf doesn't match we definitely cant verify it
					continue
				}

				// Generate the proof for the mmrIndex and get the root. We use
				// the mmrSize from the end of the blob in which the leaf entry
				// was recorded. Any size > than the leaf index would work.
				mmrSize := massif.RangeCount()
				proof, err := mmr.InclusionProof(&massif, mmrSize, mmrIndex)
				if err != nil {
					return err
				}

				verified, err := mmr.VerifyInclusion(&massif, eventHasher, mmrSize, logNodeValue, mmrIndex, proof)
				if verified {
					fmt.Printf("OK|%d %d\n", mmrIndex, leafIndex)
					continue
				}
				if err != nil {
					fmt.Printf("XX|%d %d|%s\n", mmrIndex, leafIndex, err.Error())
					continue
				}
				fmt.Printf("XX|%d %d\n", mmrIndex, leafIndex)
			}

			return nil
		},
	}
}
