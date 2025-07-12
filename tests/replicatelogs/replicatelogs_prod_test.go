//go:build integration

package verifyconsistency

import (
	"github.com/datatrails/veracity"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
)

// Test that the watch command returns no error or that the error is "no changes"
func (s *ReplicateLogsCmdSuite) TestReplicateFirstPublicMassif() {

	replicaDir := s.T().TempDir()

	app := veracity.NewApp("tests", false)
	veracity.AddCommands(app, false)

	err := app.Run([]string{
		"veracity",
		"--data-url", s.Env.VerifiableDataURL,
		"--tenant", s.Env.PublicTenantId,
		"replicate-logs",
		"--replicadir", replicaDir,
		"--progress",
		"--massif", "1",
	})
	s.NoError(err)
	logID := datatrails.TenantID2LogID(s.Env.PublicTenantId)

	expectMassifFile := mustMassifFilename(s.T(), replicaDir, logID, 0)
	s.FileExistsf(expectMassifFile, "the replicated massif should exist")

	expectCheckpointFile := mustCheckpointFilename(s.T(), replicaDir, logID, 0)
	s.FileExistsf(expectCheckpointFile, "the replicated checkpoint should exist")
}
