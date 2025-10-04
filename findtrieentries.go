package veracity

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	"github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

/**
 * find trie Entries finds the trie entry associated with a given trie key
 */

const (
	logTenantFlagName = "log-tenant"

	logIDFlagName = "log-id"

	domainFlagName = "domain"

	appIDFlagName = "app-id"

	asLeafIndexesFlagName = "as-leafindexes"

	massifRangeStartFlagName = "massif-start"
	massifRangeEndFlagName   = "massif-end"
)

var (
	ErrNoLogTenant = fmt.Errorf("error, cannot find log tenant, please provide either %v or %v", logIDFlagName, logTenantFlagName)
)

// logIDToLogTenant converts the log id to the log tenant
func logIDToLogTenant(logID string) (string, error) {

	// first get the byte representation of the hex
	logIdBytes, err := hex.DecodeString(logID)
	if err != nil {
		return "", err
	}

	// we don't know if its a log version 0 log id or a log version 1 log id

	// attempt log version 1 first

	// log version 1 is the byte representation of the uuid part of the log tenant
	logIdUUid, err := uuid.ParseBytes(logIdBytes)
	if err != nil {

		// assume its log version 0 if it can't be parsed as bytes
		// which is just utf-8 bytes of the log tenant string
		return string(logIdBytes), nil
	}

	// if we get here we know its log version 1, so make the log tenant from the uuid
	logTenant := fmt.Sprintf("tenant/%s", logIdUUid.String())

	return logTenant, nil
}

// findTrieKeys searchs the log of the given log tenant for matches to the given triekeys
// and returns the leaf indexes (trie indexes) of all the matches as well as the number of trie entries considered
func findTrieKeys(
	ctx context.Context,
	log logger.Logger,
	massifReader readerSelector,
	tenantLogPath string,
	massifStartIndex int64,
	massifEndIndex int64,
	massifHeight uint8,
	trieKeys ...[]byte,
) ([]uint64, uint64, error) {

	// find the starting leaf index by finding the number of leaf nodes in a full massif, of the given massif height,
	//  then multiplying that by the number of massifs we are skipping over
	leafIndex := uint64(massifStartIndex * int64(mmr.HeightIndexLeafCount(uint64(massifHeight-1))))

	matchingLeafIndexes := []uint64{}
	entriesConsidered := uint64(0)

	massifIndex := massifStartIndex

	// search all massifs from the starting index to the end index
	for {

		// check if we have reached the last massif we want to consider checking
		// -1 means check until the last massif, so never break here if massifEndIndex == -1
		if massifIndex > massifEndIndex && massifEndIndex != -1 {
			break
		}

		massifContext, err := massifs.GetMassifContext(ctx, massifReader, uint32(massifIndex))

		// check if we have reached the last massif
		if errors.Is(err, storage.ErrDoesNotExist) {
			break
		}

		if err != nil {

			// check if we get an azblob error of blob not found
			// this is also an indication we have reached the last massif
			if strings.Contains(err.Error(), "BlobNotFound") {
				break
			}

			return nil, 0, err
		}

		// get the trieEntry count based on the size
		// NOTE: the leaf count and trie entry count are the same
		// NOTE: the leaf index and trie index are equivilent.
		trieEntries := massifContext.MassifLeafCount()

		log.Debugf("checking %v trie entries in massif %v for matches", trieEntries, massifIndex)

		// check each trie entry for matching trieKeys
		for range trieEntries {

			mmrIndex := mmr.MMRIndex(leafIndex)

			logTrieKey, err := massifContext.GetTrieKey(mmrIndex)
			if err != nil {
				return nil, 0, err
			}

			for _, trieKey := range trieKeys {

				// if a triekey matches add it to the matching leaf indexes
				if bytes.Equal(trieKey, logTrieKey) {
					matchingLeafIndexes = append(matchingLeafIndexes, leafIndex)
					break // only one trieKey will ever match, so if we found the matching trie key, don't keep looking
				}

			}

			leafIndex++
			entriesConsidered++

		}

		massifIndex++
	}

	return matchingLeafIndexes, entriesConsidered, nil

}

// NewFindTrieEntriesCmd finds the trie entries associated with a given trie key in the tenants Merkle Log.
//
//nolint:gocognit
func NewFindTrieEntriesCmd() *cli.Command {
	return &cli.Command{
		Name: "find-trie-entries",
		Usage: `finds the matching trie entries for the given trie key.

		By default returns all mmr Indexes of matching trie entries.

		The trieKey is HASH(DOMAIN | LOGID | APPID)

		NOTE: ignores the global --tenant option, please use --log-tenant command option.
`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  logTenantFlagName,
				Usage: fmt.Sprintf("the tenant of the log to search in. Required or can be derived from %v.", logIDFlagName),
			},
			&cli.StringFlag{
				Name:  logIDFlagName,
				Usage: fmt.Sprintf("the hexadecimal representation of the log ID. Required or can be derived from %v.", logTenantFlagName),
			},
			&cli.StringFlag{
				Name:     appIDFlagName,
				Usage:    "the app ID. For eventsv1 or assetsv2 this is the event id.",
				Required: true,
			},
			&cli.Uint64Flag{
				Name:  domainFlagName,
				Usage: "the domain used to derive the trie key.",
				Value: uint64(massifs.KeyTypeApplicationContent),
			},
			&cli.BoolFlag{
				Name:  asLeafIndexesFlagName,
				Usage: "if true, returns a list of matching leaf indexes instead of mmr indexes.",
				Value: false,
			},
			&cli.Int64Flag{
				Name:  massifRangeStartFlagName,
				Usage: "if set, start the search for matching trie entries at the massif at this given massif index. if omitted will start search at massif 0.",
				Value: 0,
			},
			&cli.Int64Flag{
				Name:  massifRangeEndFlagName,
				Usage: "if set, end the search for matching trie entries at the massif at this given massif index. if omitted will end search at the last massif.",
				Value: -1,
			},
		},
		Action: func(cCtx *cli.Context) error {
			cmd := &CmdCtx{}

			var err error
			var reader readerSelector

			// This command uses the structured logger for all optional output.
			if err := cfgLogging(cmd, cCtx); err != nil {
				return err
			}

			// get all flags
			logTenant := cCtx.String(logTenantFlagName)
			logID := cCtx.String(logIDFlagName)

			appID := cCtx.String(appIDFlagName)

			domain := cCtx.Uint64(domainFlagName)

			asLeafIndexes := cCtx.Bool(asLeafIndexesFlagName)

			massifStartIndex := cCtx.Int64(massifRangeStartFlagName)
			massifEndIndex := cCtx.Int64(massifRangeEndFlagName)

			// check we only have at least 1 of log tenant or logID
			if logTenant == "" && logID == "" {
				return ErrNoLogTenant
			}

			// if we don't have the log tenant derive it from the
			//  log id
			if logTenant == "" {
				var err error
				logTenant, err = logIDToLogTenant(logID)
				if err != nil {
					return err
				}
			}

			// If we are reading the massif log locally, the log path is the
			// data-local path. The reader does the right thing regardless of
			// whether the option is a directory or a file.
			// verifyEvent defaults it to tenantIdentity for the benefit of the remote reader implementation
			tenantLogPath := cCtx.String("data-local")

			if tenantLogPath == "" {
				tenantLogPath = logTenant
			}

			// configure the cmd massif reader
			if err := cCtx.Set("tenant", logTenant); err != nil {
				return err
			}

			if reader, err = cfgMassifReader(cmd, cCtx); err != nil {
				return err
			}

			leafIndexMatches := []uint64{}
			entriesConsidered := uint64(0)

			// if we have the logID use it to find the trie key.
			if logID != "" {

				logIDBytes, err := hex.DecodeString(logID)
				if err != nil {
					return err
				}

				trieKey := massifs.NewTrieKey(
					massifs.KeyType(domain),
					logIDBytes,
					[]byte(appID),
				)

				cmd.Log.Debugf("trieKey: %x", trieKey)

				reader.SelectLog(context.Background(), logIDBytes)

				leafIndexMatches, entriesConsidered, err = findTrieKeys(
					context.Background(),
					cmd.Log,
					reader,
					tenantLogPath,
					massifStartIndex,
					massifEndIndex,
					cmd.MassifFmt.MassifHeight,
					trieKey,
				)
				if err != nil {
					return err
				}

			}

			// if we don't have the trieKey we can derive it from the log tenant, but
			// we don't know if its log version 0 or log version 1, so do both.
			if logID == "" {

				logIDVersion0 := []byte(logTenant)

				trieKeyVersion0 := massifs.NewTrieKey(
					massifs.KeyType(domain),
					logIDVersion0,
					[]byte(appID),
				)

				cmd.Log.Debugf("trieKey version 0: %x", trieKeyVersion0)

				logTenantUUIDStr := strings.TrimPrefix(logTenant, "tenant/")
				logTenantUUID, err := uuid.Parse(logTenantUUIDStr)
				if err != nil {

					// we could continue with just version 0 here
					// but if we error here we know the log tenant isn't a valid
					// tenant identity, so there is no point searching for the trie key
					// as we know its invalid.
					return err
				}

				logIDVersion1, err := logTenantUUID.MarshalBinary()
				if err != nil {
					return err
				}

				trieKeyVersion1 := massifs.NewTrieKey(
					massifs.KeyType(domain),
					logIDVersion1,
					[]byte(appID),
				)

				reader.SelectLog(context.Background(), logIDVersion1)

				cmd.Log.Debugf("trieKey version 1: %x", trieKeyVersion1)

				leafIndexMatches, entriesConsidered, err = findTrieKeys(
					context.Background(),
					cmd.Log,
					reader,
					tenantLogPath,
					massifStartIndex,
					massifEndIndex,
					cmd.MassifFmt.MassifHeight,
					trieKeyVersion0,
					trieKeyVersion1,
				)
				if err != nil {
					return err
				}

			}

			cmd.Log.Debugf("entries considered: %v", entriesConsidered)

			// if we want the leaf index matches log them and return
			if asLeafIndexes {
				fmt.Printf("matches: %v\n", leafIndexMatches)
				return nil
			}

			// otherwise we want to log the mmr index matches
			mmrIndexMatches := []uint64{}
			for _, leafIndex := range leafIndexMatches {

				mmrIndex := mmr.MMRIndex(leafIndex)
				mmrIndexMatches = append(mmrIndexMatches, mmrIndex)
			}

			fmt.Printf("matches: %v\n", mmrIndexMatches)

			return nil

		},
	}
}
