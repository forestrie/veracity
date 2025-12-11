//go:build integration

package verifyconsistency

import (
	"github.com/datatrails/veracity"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
)

// TestReplicateFirstPublicMassif tests replication of the first public massif.
// NOTE: This test may fail with certificate expiration errors when accessing app.datatrails.ai.
// This is an environmental issue (expired SSL certificate on the external service) and cannot be fixed in code.
// DISABLED: Test disabled due to expired TLS certificate on datatrails official log endpoint.
func (s *ReplicateLogsCmdSuite) TestReplicateFirstPublicMassif() {
	s.T().Skip("Test disabled due to expired TLS certificate on datatrails official log endpoint (app.datatrails.ai)")

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
