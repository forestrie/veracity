//go:build integration && azurite

package verifyevents

import (
	"fmt"

	"testing"

	"github.com/datatrails/go-datatrails-common-api-gen/assets/v2/assets"
	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/datatrails/go-datatrails-simplehash/simplehash"
	"github.com/datatrails/veracity"
	"github.com/datatrails/veracity/tests/testcontext"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
	"github.com/robinbryce/go-merklelog-provider-testing/mmrtesting"
	"github.com/stretchr/testify/require"
)

// TestVerifyIncludedMultiMassif tests that the veracity sub command verify-included
// works for massifs beyond the first one and covers some obvious edge cases.
func (s *VerifyEventsSuite) TestVerifyIncludedMultiMassif() {
	logger.New("TestVerifyIncludedMultiMassif")
	defer logger.OnExit()

	massifHeight := uint8(8)
	leavesPerMassif := mmr.HeightIndexLeafCount(uint64(massifHeight) - 1)

	tests := []struct {
		name        string
		massifCount uint32
		// leaf indices to verify the inclusion of.
		leaves []uint64
	}{
		// make sure we cover the obvious edge cases
		{name: "2 massifs, last of first and first of last", massifCount: 2, leaves: []uint64{leavesPerMassif - 1, leavesPerMassif}},
		{name: "single massif first few and last few", massifCount: 1, leaves: []uint64{0, 1, 2, leavesPerMassif - 2, leavesPerMassif - 1}},
		{name: "5 massifs, first and last of each", massifCount: 5, leaves: []uint64{
			0, leavesPerMassif - 1,
			1 * leavesPerMassif, 2*leavesPerMassif - 1,
			2 * leavesPerMassif, 3*leavesPerMassif - 1,
			3 * leavesPerMassif, 4*leavesPerMassif - 1,
		}},
	}

	// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
	app := veracity.NewApp("tests", true)
	veracity.AddCommands(app, true)

	for _, tt := range tests {

		massifCount := tt.massifCount
		s.Run(fmt.Sprintf("massifCount:%d", massifCount), func() {

			tc, logId, _, generated := testcontext.CreateLogBuilderContext(
				s.T(),
				massifHeight, tt.massifCount,
				mmrtesting.WithTestLabelPrefix("TestVerifyIncludedMultiMassif"))

			for _, iLeaf := range tt.leaves {

				event := datatrailsAssetEvent(
					s.T(), generated.Encoded[iLeaf], generated.Args[iLeaf],
					generated.MMRIndices[iLeaf], uint8(massifs.Epoch2038),
				)
				marshaler := simplehash.NewEventMarshaler()
				eventJson, err := marshaler.Marshal(event)
				require.NoError(s.T(), err)
				s.StdinWriteAndClose(eventJson)
				tenantId0 := datatrails.Log2TenantID(logId)

				err = app.Run([]string{
					"veracity",
					"--envauth", // uses the emulator
					"--container", tc.Cfg.Container,
					"--data-url", s.Env.AzuriteVerifiableDataURL,
					"--tenant", tenantId0,
					"--height", fmt.Sprintf("%d", massifHeight),
					"verify-included",
				})
				s.NoError(err)
				s.ReplaceStdin() // reset stdin for write & close
			}
		})
	}
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
