//go:build integration && azurite

package veracity

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/datatrails/go-datatrails-common/logger"
	"github.com/datatrails/veracity/tests/testcontext"
	"github.com/forestrie/go-merklelog-datatrails/datatrails"
	"github.com/forestrie/go-merklelog-provider-testing/mmrtesting"
	"github.com/stretchr/testify/assert"
)

const (
	massifHeight = 14
)

// TestConfigReaderForEmulator tests that we can run the veracity tool against
// emulator urls.
func TestNodeCmd(t *testing.T) {
	logger.New("TestVerifyList")
	defer logger.OnExit()
	url := os.Getenv("TEST_INTEGRATION_FORESTRIE_BLOBSTORE_URL")
	logger.Sugar.Infof("url: '%s'", url)

	// Create a single massif in the emulator
	tc, logID := testcontext.CreateLogContext(
		t, 8, 1,
		mmrtesting.WithTestLabelPrefix("TestNodeCmd"),
	)

	tenantID := datatrails.Log2TenantID(logID)

	tests := []struct {
		testArgs []string
	}{
		// get node 1
		{testArgs: []string{"<progname>", "-s", "devstoreaccount1", "-c", tc.Cfg.Container, "-t", tenantID, "node", fmt.Sprintf("%d", 1)}},
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
