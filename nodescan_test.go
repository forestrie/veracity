//go:build integration && azurite

package veracity

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/datatrails/go-datatrails-common-api-gen/assets/v2/assets"
	v2assets "github.com/datatrails/go-datatrails-common-api-gen/assets/v2/assets"
	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/datatrails/go-datatrails-simplehash/simplehash"
	"github.com/datatrails/veracity/tests/testcontext"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
	"github.com/robinbryce/go-merklelog-provider-testing/mmrtesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigReaderForEmulator tests that we can run the veracity tool against
// emulator urls.
func TestNodeScanCmd(t *testing.T) {
	logger.New("TestVerifyList")
	defer logger.OnExit()
	url := os.Getenv("TEST_INTEGRATION_FORESTRIE_BLOBSTORE_URL")
	logger.Sugar.Infof("url: '%s'", url)

	// Create a single massif in the emulator
	tc, logID, _, generated := testcontext.CreateLogBuilderContext(
		t, 8, 1,
		mmrtesting.WithTestLabelPrefix("TestNodeScanCmd"),
	)

	// tc.GenerateTenantLog(10)

	tenantID := datatrails.Log2TenantID(logID)

	marshaledEvents, eventsResponse, err := marshalEventsList(tc, generated)
	require.NoError(t, err)

	// Arbitrarily chose to look for leaf 7
	leafIndex := 7

	// We could try an capture the output from the tool and check this, but that
	// sort of thing is extensively checked elsewhere. Here we are just
	// confirming the derived leaf hash is matched in the blob. WHich is the
	// purpose of the diagnostic tool. But we log the expected mmrIndex for development ease.

	fmt.Printf("MMRIndex(%d) == %d\n", leafIndex, mmr.MMRIndex(uint64(leafIndex)))

	// Get the idtimestamp that ensures every entry is unique
	idTimestamp, _, err := massifs.SplitIDTimestampHex(
		eventsResponse[leafIndex].MerklelogEntry.Commit.Idtimestamp)
	require.NoError(t, err)

	// Produce the expected leaf hash for the tree entry for this event
	simplehashv3Hasher := simplehash.NewHasherV3()
	err = simplehashv3Hasher.HashEventFromJSON(
		marshaledEvents[leafIndex],
		simplehash.WithPrefix([]byte{LeafTypePlain}),
		simplehash.WithIDCommitted(idTimestamp))
	require.NoError(t, err)
	expectedLeafNodeValue := fmt.Sprintf("%x", simplehashv3Hasher.Sum(nil))

	tests := []struct {
		testArgs []string
	}{
		// match the expected leaf hash by scanning massif 0. typically this is
		// useful only as a diagnostic tool or triage tool. typically the
		// precise node is located by the mmrIndex and the leafIndex derived
		// from that.
		{testArgs: []string{
			"<progname>", "-s", "devstoreaccount1", "-c", tc.Cfg.Container, "-t", tenantID,
			"nodescan", "-m", "0", "-v", expectedLeafNodeValue}},
	}

	for _, tc := range tests {
		t.Run(strings.Join(tc.testArgs, " "), func(t *testing.T) {
			app := AddCommands(NewApp("version", true), true)
			ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			err := app.RunContext(ctx, tc.testArgs)
			cancel()
			assert.NoError(t, err)
		})
	}
}

func marshalEventsList(
	tc *testcontext.TestContext, generated mmrtesting.GeneratedLeaves) ([][]byte, []*assets.EventResponse, error) {

	marshaller := v2assets.NewFlatMarshalerForEvents()
	eventJsonList := make([][]byte, 0)
	events := make([]*assets.EventResponse, len(generated.MMRIndices))

	for iLeaf := 0; iLeaf < len(generated.MMRIndices); iLeaf++ {
		event := datatrailsAssetEvent(
			tc.T, generated.Encoded[iLeaf], generated.Args[iLeaf],
			generated.MMRIndices[iLeaf], uint8(massifs.Epoch2038),
		)
		events[iLeaf] = event
		eventJson, err := marshaller.Marshal(event)
		if err != nil {
			return nil, nil, err
		}

		eventJsonList = append(eventJsonList, eventJson)
	}
	return eventJsonList, events, nil
}

func datatrailsAssetEvent(t *testing.T, a any, args mmrtesting.AddLeafArgs, index uint64, epoch uint8) *assets.EventResponse {
	ae, ok := a.(*assets.EventResponse)
	require.True(t, ok, "expected *assets.EventResponse, got %T", a)

	ae.MerklelogEntry = &assets.MerkleLogEntry{
		Commit: &assets.MerkleLogCommit{
			Index:       index,
			Idtimestamp: massifs.IDTimestampToHex(args.ID, epoch),
		},
	}
	return ae
}
