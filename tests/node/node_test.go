//go:build integration && azurite

package node

import (
	"fmt"
	"strings"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/datatrails/veracity"
	"github.com/datatrails/veracity/tests/testcontext"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
	"github.com/robinbryce/go-merklelog-provider-testing/mmrtesting"
	"github.com/stretchr/testify/assert"
)

// TestNodeMultiMassif tests that the veracity sub command node
// works for massifs beyond the first one and covers some obvious edge cases.
// This really just tests that the correspondence between the massif index and the leaf index holds
// regardless of the massif count.
func (s *NodeSuite) TestVerifyIncludedMultiMassif() {
	logger.New("TestNodeMultiMassif")
	defer logger.OnExit()

	massifHeight := uint8(8)
	leavesPerMassif := mmr.HeightIndexLeafCount(uint64(massifHeight) - 1)

	tests := []struct {
		name        string
		massifCount uint32
		// leaf indices to check.
		leaves []uint64
	}{
		// make sure we cover the obvious edge cases
		{name: "leaf 0", massifCount: 1, leaves: []uint64{0}},
		{name: "single massif first few and last few", massifCount: 1, leaves: []uint64{0, 1, 2, leavesPerMassif - 2, leavesPerMassif - 1}},
		{name: "2 massifs, last of first and first of last", massifCount: 2, leaves: []uint64{leavesPerMassif - 1, leavesPerMassif}},
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

			tc, logID, _, generated := testcontext.CreateLogBuilderContext(
				s.T(),
				massifHeight,
				tt.massifCount,
				mmrtesting.WithTestLabelPrefix("TestNodeMultiMassif"),
			)

			tenantID := datatrails.Log2TenantID(logID)

			for _, iLeaf := range tt.leaves {

				s.ReplaceStdout()

				mmrIndex := mmr.MMRIndex(iLeaf)

				err := app.Run([]string{
					"veracity",
					"--envauth", // uses the emulator
					"--container", tc.Cfg.Container,
					"--data-url", s.Env.AzuriteVerifiableDataURL,
					"--tenant", tenantID,
					"--height", fmt.Sprintf("%d", massifHeight),
					"node",
					"--mmrindex", fmt.Sprintf("%d", mmrIndex),
				})
				s.NoError(err)

				stdout := s.CaptureAndCloseStdout()

				leafValue := fmt.Sprintf("%x", generated.Args[iLeaf].Value)
				assert.Equal(s.T(), leafValue, strings.TrimSpace(stdout))
			}
		})
	}
}
