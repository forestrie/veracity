//go:build integration

package verifyconsistency

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	"github.com/google/uuid"
	fsstorage "github.com/robinbryce/go-merklelog-fs/storage"
	"github.com/stretchr/testify/require"
)

func fileSHA256(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}

func mustHashFile(t *testing.T, filename string) []byte {
	t.Helper()
	hash, err := fileSHA256(filename)
	require.NoError(t, err)
	return hash
}

func mustMassifFilename(t *testing.T, replicaDir string, logID storage.LogID, massifIndex uint32) string {
	t.Helper()
	prefix := filepath.Join(
		replicaDir, fsstorage.LogIDPrefix, uuid.UUID(logID).String(), fsstorage.MassifsDirName) + "/"

	return storage.FmtMassifPath(prefix, uint32(massifIndex))
}

func mustCheckpointFilename(t *testing.T, replicaDir string, logID storage.LogID, massifIndex uint32) string {
	t.Helper()

	prefix := filepath.Join(
		replicaDir, fsstorage.LogIDPrefix, uuid.UUID(logID).String(), fsstorage.CheckpointsDirName) + "/"
	return storage.FmtCheckpointPath(prefix, uint32(massifIndex))
}
