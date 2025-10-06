//go:build integration && azurite

package verifyconsistency

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/forestrie/go-merklelog/massifs/storage"
	"github.com/forestrie/go-merklelog/mmr"
	"github.com/datatrails/veracity"
	"github.com/datatrails/veracity/tests/testcontext"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
	"github.com/forestrie/go-merklelog-provider-testing/mmrtesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReplicateMassifUpdate ensures that an extension to a previously replicated
// massif is handled correctly
func (s *ReplicateLogsCmdSuite) TestReplicateMassifUpdate() {
	logger.New("TestReplicateMassifUpdate")
	defer logger.OnExit()

	tc := testcontext.NewDefaultTestContext(
		s.T(),
		mmrtesting.WithTestLabelPrefix("TestReplicateMassifUpdate"),
	)

	// getter, err := tc.NewNativeObjectReader(massifs.StorageOptions{MassifHeight: integrationsupport.TestMassifHeight})
	// require.Nil(t, err)

	ctx := context.TODO()

	h8MassifLeaves := mmr.HeightIndexLeafCount(uint64(8 - 1)) // = ((2 << massifHeight) - 1 + 1) >> 1

	tests := []struct {
		name                   string
		massifHeight           uint8
		firstUpdateMassifs     uint64
		firstUpdateExtraLeaves uint64
		// if zero, the last massif will be completed. If the last is full and secondUpdateMassifs is zero the test is invalid
		secondUpdateMassifs     uint64
		secondUpdateExtraLeaves uint64
	}{
		// extend second massif
		{name: "complete first massif", massifHeight: 8, firstUpdateMassifs: 1, firstUpdateExtraLeaves: h8MassifLeaves - 3, secondUpdateMassifs: 0, secondUpdateExtraLeaves: 3},

		// make sure we cover the obvious edge cases
		{name: "complete first massif", massifHeight: 8, firstUpdateMassifs: 0, firstUpdateExtraLeaves: h8MassifLeaves - 3, secondUpdateMassifs: 0, secondUpdateExtraLeaves: 3},

		// make sure we cover update from partial blob to new massif
		{name: "partial first massif", massifHeight: 8, firstUpdateMassifs: 0, firstUpdateExtraLeaves: h8MassifLeaves - 6, secondUpdateMassifs: 2, secondUpdateExtraLeaves: 0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Populate the log with the content for the first update

			var err error

			require.True(s.T(), tt.firstUpdateMassifs > 0 || tt.firstUpdateExtraLeaves > 0, uint32(0), "invalid test")
			require.True(s.T(), tt.secondUpdateMassifs > 0 || tt.secondUpdateExtraLeaves > 0, uint32(0), "invalid test")
			replicaDir := s.T().TempDir()

			// tc.GenerateTenantLog(10, massifHeight, 0 /* leaf type plain */)
			logId0 := tc.G.NewLogID()
			tenantId0 := datatrails.Log2TenantID(logId0)

			builder, _ := testcontext.CreateLogForContext(tc, logId0, tt.massifHeight, uint32(tt.firstUpdateMassifs))

			leavesPerMassif := mmr.HeightIndexLeafCount(uint64(tt.massifHeight) - 1) // = ((2 << massifHeight) - 1 + 1) >> 1

			if tt.firstUpdateExtraLeaves > 0 {
				tc.AddLeaves(
					ctx, builder, logId0, tt.massifHeight, leavesPerMassif*tt.firstUpdateMassifs, tt.firstUpdateExtraLeaves,
				)
			}

			// Replicate the log
			// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
			app := veracity.NewApp("tests", true)
			veracity.AddCommands(app, true)

			err = app.Run([]string{
				"veracity",
				"--envauth", // uses the emulator
				"--container", tc.Cfg.Container,
				"--data-url", s.Env.AzuriteVerifiableDataURL,
				"--tenant", tenantId0,
				"--height", fmt.Sprintf("%d", tt.massifHeight),
				"replicate-logs",
				// "--ancestors", fmt.Sprintf("%d", tt.ancestors),
				"--replicadir", replicaDir,
				// note: firstUpdateMassifs is the count of *full* massifs we add before adding the leaves, so the count is also the "index" of the last massif.
				"--massif", fmt.Sprintf("%d", tt.firstUpdateMassifs),
			})
			s.NoError(err)

			headIndex, err := builder.ObjectReader.HeadIndex(ctx, storage.ObjectMassifData)
			s.NoError(err)

			firstMassifFilename := mustMassifFilename(s.T(), replicaDir, logId0, headIndex)
			firstHash := mustHashFile(s.T(), firstMassifFilename)

			// Add the content for the second update

			if tt.secondUpdateMassifs > 0 {
				massifLeaves := mmr.HeightIndexLeafCount(uint64(tt.massifHeight - 1)) // = ((2 << massifHeight) - 1 + 1) >> 1
				// CreateLog always deleted blobs, so we can only use AddLeavesToLog here
				for i := range tt.secondUpdateMassifs {
					tc.AddLeaves(
						ctx, builder, logId0, tt.massifHeight, leavesPerMassif*tt.firstUpdateMassifs+i, massifLeaves,
					)
				}
			}

			if tt.secondUpdateExtraLeaves > 0 {
				tc.AddLeaves(
					ctx, builder, logId0, tt.massifHeight, leavesPerMassif*(tt.firstUpdateMassifs+tt.secondUpdateMassifs), tt.secondUpdateExtraLeaves,
				)
			}

			// Replicate the content
			err = app.Run([]string{
				"veracity",
				"--envauth", // uses the emulator
				"--container", tc.Cfg.Container,
				"--data-url", s.Env.AzuriteVerifiableDataURL,
				"--tenant", tenantId0,
				"--height", fmt.Sprintf("%d", tt.massifHeight),
				"replicate-logs",
				"--replicadir", replicaDir,
				"--massif", fmt.Sprintf("%d", tt.firstUpdateMassifs+tt.secondUpdateMassifs),
			})
			s.NoError(err)

			headIndex, err = builder.ObjectReader.HeadIndex(ctx, storage.ObjectMassifData)
			s.NoError(err)

			// note: secondMassifFilename *may* be same as first depending on test config
			secondMassifFilename := mustMassifFilename(s.T(), replicaDir, logId0, headIndex)
			secondHash := mustHashFile(s.T(), secondMassifFilename)

			assert.NotEqual(s.T(), firstHash, secondHash, "the massif should have changed")

			// Attempt to replicate again, this will verify the local state and then do nothing
			err = app.Run([]string{
				"veracity",
				"--envauth", // uses the emulator
				"--container", tc.Cfg.Container,
				"--data-url", s.Env.AzuriteVerifiableDataURL,
				"--tenant", tenantId0,
				"--height", fmt.Sprintf("%d", tt.massifHeight),
				"replicate-logs",
				"--replicadir", replicaDir,
				"--massif", fmt.Sprintf("%d", tt.firstUpdateMassifs+tt.secondUpdateMassifs),
			})

			s.NoError(err)
		})
	}
}

// TestReplicatingMassifLogsForOneTenant test that by default af full replica is made
func (s *ReplicateLogsCmdSuite) TestReplicatingMassifLogsForOneTenant() {
	logger.New("Test4AzuriteMassifsForOneTenant")
	defer logger.OnExit()

	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("Test4AzuriteMassifsForOneTenant"))
	massifHeight := uint8(8)

	tests := []struct {
		massifCount uint32
	}{
		// make sure we cover the obvious edge cases
		{massifCount: 2},
		{massifCount: 5},
		{massifCount: 1},
	}

	for _, tt := range tests {

		massifCount := tt.massifCount

		s.Run(fmt.Sprintf("massifCount:%d", massifCount), func() {
			logId0 := tc.G.NewLogID()
			tenantId0 := datatrails.Log2TenantID(logId0)
			testcontext.CreateLogForContext(tc, logId0, massifHeight, massifCount)

			replicaDir := s.T().TempDir()

			// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
			app := veracity.NewApp("tests", true)
			veracity.AddCommands(app, true)

			err := app.Run([]string{
				"veracity",
				"--envauth", // uses the emulator
				"--container", tc.Cfg.Container,
				"--data-url", s.Env.AzuriteVerifiableDataURL,
				"--tenant", tenantId0,
				"--height", fmt.Sprintf("%d", massifHeight),
				"replicate-logs",
				"--replicadir", replicaDir,
				"--massif", fmt.Sprintf("%d", massifCount-1),
			})
			s.NoError(err)

			for i := range massifCount {
				expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
				s.FileExistsf(expectMassifFile, "the replicated massif should exist")
				expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
				s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
			}
		})
	}
}

// TestAncestorMassifsForOneTenant tests that the --ancestors option
// limits the number of historical massifs that are replicated Note that
// --ancestors=0 still requires consistency against local replica of the remote
func (s *ReplicateLogsCmdSuite) TestAncestorMassifLogsForOneTenant() {
	logger.New("Test4AzuriteMassifsForOneTenant")
	defer logger.OnExit()

	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("TestAncestorMassifLogsForOneTenant"))

	massifHeight := uint8(8)

	tests := []struct {
		massifCount uint32
		ancestors   uint32
	}{
		// make sure we cover the obvious edge cases
		{massifCount: 1, ancestors: 1},
		{massifCount: 2, ancestors: 1},
		{massifCount: 5, ancestors: 1},
		{massifCount: 1, ancestors: 2},
		{massifCount: 2, ancestors: 2},
		{massifCount: 5, ancestors: 2},
		{massifCount: 2, ancestors: 3},
		{massifCount: 5, ancestors: 3},

		{massifCount: 5},
		{massifCount: 1},
		{massifCount: 2},
	}

	for _, tt := range tests {

		massifCount := tt.massifCount

		s.Run(fmt.Sprintf("massifCount:%d", massifCount), func() {
			logId0 := tc.G.NewLogID()
			tenantId0 := datatrails.Log2TenantID(logId0)

			testcontext.CreateLogForContext(tc, logId0, massifHeight, massifCount)

			replicaDir := s.T().TempDir()

			// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
			app := veracity.NewApp("tests", true)
			veracity.AddCommands(app, true)

			err := app.Run([]string{
				"veracity",
				"--envauth", // uses the emulator
				"--container", tc.Cfg.Container,
				"--data-url", s.Env.AzuriteVerifiableDataURL,
				"--tenant", tenantId0,
				"--height", fmt.Sprintf("%d", massifHeight),
				"replicate-logs",
				"--ancestors", fmt.Sprintf("%d", tt.ancestors),
				"--replicadir", replicaDir,
				"--massif", fmt.Sprintf("%d", massifCount-1),
			})
			s.NoError(err)

			if tt.ancestors >= massifCount-1 {
				// then all massifs should be replicated
				for i := range massifCount {
					expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
					s.FileExistsf(expectMassifFile, "the replicated massif should exist")
					expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
					s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
				}
				return
			}

			// To allow testing cases where the ancestors are greater than the count, we need to guard against underflow here.
			end := max(2, massifCount) - 2 - tt.ancestors

			for i := range end {
				expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
				s.NoFileExistsf(expectMassifFile, "the replicated massif should exist")
				expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
				s.NoFileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
			}

			for i := massifCount - 1 - tt.ancestors; i < massifCount; i++ {

				expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
				s.FileExistsf(expectMassifFile, "the replicated massif should exist")
				expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
				s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
			}
		})
	}
}

// TestSparseReplicaCreatedAfterExtendedOffline tests that the --ancestors
// option limits the number of historical massifs that are replicated, and in
// the case where the verifier has been off line for a long, the resulting
// replica is sparse. --ancestors is set what the user wants to have a bound on
// the work done in any one run
func (s *ReplicateLogsCmdSuite) TestSparseReplicaCreatedAfterExtendedOffline() {
	logger.New("TestSparseReplicaCreatedAfterExtendedOffline")
	defer logger.OnExit()

	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("TestSparseReplicaCreatedAfterExtendedOffline"))
	ctx := context.TODO()
	massifCount := uint32(4)
	massifHeight := uint8(8)

	logId0 := tc.G.NewLogID()

	_, builders := testcontext.CreateLogsForContext(tc, massifHeight, 1, logId0)
	builder := builders[0]

	// This test requires two invocations. For the first invocation, we make ony one massif available.
	// Then after that is successfully replicated, we add the rest of the massifs.
	tenantId0 := datatrails.Log2TenantID(logId0)

	leavesPerMassif := mmr.HeightIndexLeafCount(uint64(massifHeight) - 1) // = ((2 << massifHeight) - 1 + 1) >> 1

	replicaDir := s.T().TempDir()

	// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
	app := veracity.NewApp("tests", true)
	veracity.AddCommands(app, true)

	err := app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--tenant", tenantId0,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		//  --ancestors defaults to 0 which means "all", but only massif is available
		"--replicadir", replicaDir,
		"--massif", "0",
	})
	s.NoError(err)

	// add the rest of the massifs
	for i := uint32(1); i < massifCount; i++ {
		tc.AddLeaves(
			ctx, builder, logId0, massifHeight, leavesPerMassif*uint64(i), leavesPerMassif,
		)
	}

	// This call, due to the --ancestors=1, should only replicate the last
	// massif, and this will leave a gap in the local replica. Importantly, this
	// means the remote log has not been checked as being consistent with the
	// local. The supported way to fill the gaps is to run with --ancestors=0 (which is the default)
	// this will fill the gaps and ensure remote/local consistency
	err = app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--tenant", tenantId0,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		"--ancestors", "1", // this will replicate massif 2 & 3
		"--replicadir", replicaDir,
		"--massif", fmt.Sprintf("%d", massifCount-1),
	})
	s.NoError(err)

	// check the 0'th massifs and seals was replicated (by the first run of veractity)
	expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(0))
	s.FileExistsf(expectMassifFile, "the replicated massif should exist")
	expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(0))
	s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")

	// check the gap was not mistakenly filled
	for i := uint32(1); i < massifCount-2; i++ {
		expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, i)
		s.NoFileExistsf(expectMassifFile, "the replicated massif should NOT exist")
		expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, i)
		s.NoFileExistsf(expectCheckpointFile, "the replicated checkpoint should NOT exist")
	}

	// check the massifs from the second veracity run were replicated
	for i := massifCount - 2; i < massifCount; i++ {
		expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, i)
		s.FileExistsf(expectMassifFile, "the replicated massif should exist")
		expectSealFile := mustCheckpointFilename(s.T(), replicaDir, logId0, i)
		s.FileExistsf(expectSealFile, "the replicated seal should exist")
	}
}

// TestFullReplicaByDefault tests that we get a full replica when
// updating a previous replica after further massifs have been added
func (s *ReplicateLogsCmdSuite) TestFullReplicaByDefault() {
	logger.New("TestFullReplicaByDefault")
	defer logger.OnExit()

	ctx := context.TODO()

	massifCount := uint32(4)
	massifHeight := uint8(8)

	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("TestFullReplicaByDefault"))
	logId0 := tc.G.NewLogID()
	_, builders := testcontext.CreateLogsForContext(
		tc, massifHeight, massifCount, logId0,
	)
	builder := builders[0]

	// This test requires two invocations. For the first invocation, we make ony one massif available.
	// Then after that is successfully replicated, we add the rest of the massifs.
	tenantId0 := datatrails.Log2TenantID(logId0)

	leavesPerMassif := mmr.HeightIndexLeafCount(uint64(massifHeight) - 1) // = ((2 << massifHeight) - 1 + 1) >> 1

	replicaDir := s.T().TempDir()

	// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
	app := veracity.NewApp("tests", true)
	veracity.AddCommands(app, true)

	err := app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--tenant", tenantId0,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		//  --ancestors defaults to 0 which means "all", but only massif is available
		"--replicadir", replicaDir,
		"--massif", "0",
	})
	s.NoError(err)

	// add the rest of the massifs
	for i := uint32(1); i < massifCount; i++ {
		tc.AddLeaves(
			ctx, builder, logId0, massifHeight, leavesPerMassif*uint64(i), leavesPerMassif,
		)
	}

	// This call, due to the --ancestors=0 default, should replicate all the new massifs.
	// The previously replicated massifs should not be re-verified.
	// The first new replicated massif should be verified as consistent with the
	// last local massif. This last point isn't assured by this test, but if
	// debugging it, that behavior can be observed.
	err = app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--tenant", tenantId0,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		//  --ancestors defaults to 0 which means "all", but only massif is available
		"--replicadir", replicaDir,
		"--massif", fmt.Sprintf("%d", massifCount-1),
	})
	s.NoError(err)

	// check the 0'th massifs and seals was replicated (by the first run of veractity)
	expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, 0)
	s.FileExistsf(expectMassifFile, "the replicated massif should exist")
	expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, 0)
	s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")

	// check the massifs from the second veracity run were replicated
	for i := uint32(1); i < massifCount; i++ {

		expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
		s.FileExistsf(expectMassifFile, "the replicated massif should exist")
		expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
		s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
	}
}

// Test4MassifsForThreeTenants multiple massifs are replicated
// when the output of the watch command is provided on stdin
func (s *ReplicateLogsCmdSuite) Test4MassifsForThreeTenants() {
	logger.New("Test4AzuriteMassifsForThreeTenants")
	defer logger.OnExit()

	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("Test4AzuriteMassifsForThreeTenants"))

	massifCount := uint32(4)
	massifHeight := uint8(8)

	logId0 := tc.G.NewLogID()
	logId1 := tc.G.NewLogID()
	logId2 := tc.G.NewLogID()

	testcontext.CreateLogsForContext(tc, massifHeight, massifCount, logId0, logId1, logId2)

	changes := []struct {
		LogID       storage.LogID `json:"logid"`
		MassifIndex int           `json:"massifindex"`
	}{
		{logId0, int(massifCount - 1)},
		{logId1, int(massifCount - 1)},
		{logId2, int(massifCount - 1)},
	}

	data, err := json.Marshal(changes)
	s.NoError(err)
	// note: the suite does a before & after pipe for Stdin
	s.StdinWriteAndClose(data)

	replicaDir := s.T().TempDir()

	// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
	app := veracity.NewApp("tests", true)
	veracity.AddCommands(app, true)

	err = app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		"--replicadir", replicaDir,
	})
	s.NoError(err)

	for _, change := range changes {
		for i := range change.MassifIndex + 1 {
			expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(
				expectMassifFile, "the replicated massif should exist")
			expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
		}
	}
}

// TestThreeTenantsOneAtATime uses --concurrency to force the replication to go one tenant at a time
// The test just ensures the obvious boundary case works
func (s *ReplicateLogsCmdSuite) TestThreeTenantsOneAtATime() {
	logger.New("TestThreeTenantsOneAtATime")
	defer logger.OnExit()

	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("TestThreeTenantsOneAtATime"))

	massifCount := uint32(4)
	massifHeight := uint8(8)

	logId0 := tc.G.NewLogID()
	logId1 := tc.G.NewLogID()
	logId2 := tc.G.NewLogID()
	testcontext.CreateLogsForContext(tc, massifHeight, massifCount, logId0, logId1, logId2)

	changes := []struct {
		LogID       storage.LogID `json:"logid"`
		MassifIndex int           `json:"massifindex"`
	}{
		{logId0, int(massifCount - 1)},
		{logId1, int(massifCount - 1)},
		{logId2, int(massifCount - 1)},
	}

	data, err := json.Marshal(changes)
	s.NoError(err)
	// note: the suite does a before & after pipe for Stdin
	s.StdinWriteAndClose(data)

	replicaDir := s.T().TempDir()

	// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
	app := veracity.NewApp("tests", true)
	veracity.AddCommands(app, true)

	err = app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		"--replicadir", replicaDir,
		"--concurrency", "1",
	})
	s.NoError(err)

	for _, change := range changes {
		for i := range change.MassifIndex + 1 {
			expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(
				expectMassifFile, "the replicated massif should exist")
			expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
		}
	}
}

// TestConcurrencyZero uses --concurrency to force the replication to go one tenant at a time
// The test just ensures the obvious boundary case works
func (s *ReplicateLogsCmdSuite) TestConcurrencyZero() {
	logger.New("TestConcurrencyZero")
	defer logger.OnExit()

	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("TestConcurrencyZero"))

	massifCount := uint32(4)
	massifHeight := uint8(8)

	logId0 := tc.G.NewLogID()
	logId1 := tc.G.NewLogID()
	logId2 := tc.G.NewLogID()

	testcontext.CreateLogsForContext(tc, massifHeight, massifCount, logId0, logId1, logId2)

	changes := []struct {
		LogID       storage.LogID `json:"logid"`
		MassifIndex int           `json:"massifindex"`
	}{
		{logId0, int(massifCount - 1)},
		{logId1, int(massifCount - 1)},
		{logId2, int(massifCount - 1)},
	}

	data, err := json.Marshal(changes)
	s.NoError(err)
	// note: the suite does a before & after pipe for Stdin
	s.StdinWriteAndClose(data)

	replicaDir := s.T().TempDir()

	// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
	app := veracity.NewApp("tests", true)
	veracity.AddCommands(app, true)

	err = app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		"--replicadir", replicaDir,
		"--concurrency", "0",
	})
	s.NoError(err)

	for _, change := range changes {
		for i := range change.MassifIndex + 1 {
			expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(
				expectMassifFile, "the replicated massif should exist")

			expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
		}
	}
}

// TestConcurrencyCappedToTenantCount sets --concurrency greater than the number of tenants
// and shows all tenants are replicated
func (s *ReplicateLogsCmdSuite) TestConcurrencyCappedToTenantCount() {
	logger.New("TestConcurrencyCappedToTenantCount")
	defer logger.OnExit()

	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("TestConcurrencyCappedToTenantCount"))

	massifCount := uint32(4)
	massifHeight := uint8(8)

	logId0 := tc.G.NewLogID()
	logId1 := tc.G.NewLogID()
	logId2 := tc.G.NewLogID()

	testcontext.CreateLogsForContext(tc, massifHeight, massifCount, logId0, logId1, logId2)

	changes := []struct {
		LogID       storage.LogID `json:"logid"`
		MassifIndex int           `json:"massifindex"`
	}{
		{logId0, int(massifCount - 1)},
		{logId1, int(massifCount - 1)},
		{logId2, int(massifCount - 1)},
	}

	data, err := json.Marshal(changes)
	s.NoError(err)
	// note: the suite does a before & after pipe for Stdin
	s.StdinWriteAndClose(data)

	replicaDir := s.T().TempDir()

	// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
	app := veracity.NewApp("tests", true)
	veracity.AddCommands(app, true)

	err = app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		"--replicadir", replicaDir,
		"--concurrency", "30000",
	})
	s.NoError(err)

	for _, change := range changes {
		for i := range change.MassifIndex + 1 {

			expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(
				expectMassifFile, "the replicated massif should exist")
			expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
		}
	}
}

// Test4MassifsForThreeTenantsFromFile multiple massifs are replicated
// when the output of the watch command is provided in a file on disc
func (s *ReplicateLogsCmdSuite) Test4MassifsForThreeTenantsFromFile() {
	logger.New("Test4AzuriteMassifsForThreeTenantsFromFile")
	defer logger.OnExit()
	tc := testcontext.NewDefaultTestContext(s.T(), mmrtesting.WithTestLabelPrefix("Test4AzuriteMassifsForThreeTenantsFromFile"))

	massifCount := uint32(4)
	massifHeight := uint8(8)

	logId0 := tc.G.NewLogID()
	logId1 := tc.G.NewLogID()
	logId2 := tc.G.NewLogID()

	testcontext.CreateLogsForContext(tc, massifHeight, massifCount, logId0, logId1, logId2)

	changes := []struct {
		LogID       storage.LogID `json:"logid"`
		MassifIndex int           `json:"massifindex"`
	}{
		{logId0, int(massifCount - 1)},
		{logId1, int(massifCount - 1)},
		{logId2, int(massifCount - 1)},
	}

	data, err := json.Marshal(changes)
	s.NoError(err)
	// note: the suite does a before & after pipe for Stdin

	replicaDir := s.T().TempDir()

	inputFilename := filepath.Join(s.T().TempDir(), "input.json")
	createFileFromData(s.T(), data, inputFilename)

	// note: VERACITY_IKWID is set in main, we need it to enable --envauth so we force it here
	app := veracity.NewApp("tests", true)
	veracity.AddCommands(app, true)

	err = app.Run([]string{
		"veracity",
		"--envauth", // uses the emulator
		"--container", tc.Cfg.Container,
		"--data-url", s.Env.AzuriteVerifiableDataURL,
		"--height", fmt.Sprintf("%d", massifHeight),
		"replicate-logs",
		"--replicadir", replicaDir,
		"--changes", inputFilename,
	})
	s.NoError(err)

	for _, change := range changes {
		for i := range change.MassifIndex + 1 {
			expectMassifFile := mustMassifFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(
				expectMassifFile, "the replicated massif should exist")

			expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logId0, uint32(i))
			s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
		}
	}
}

func createFileFromData(t *testing.T, data []byte, filename string) {
	f, err := os.Create(filename)
	require.NoError(t, err)
	defer f.Close()
	n, err := f.Write(data)
	require.NoError(t, err)
	require.Equal(t, n, len(data))
}
