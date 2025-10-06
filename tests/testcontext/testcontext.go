//go:build integration && azurite

package testcontext

import (
	"testing"
	"time"

	"github.com/forestrie/go-merklelog/massifs"
	"github.com/forestrie/go-merklelog/massifs/storage"
	azstoragetesting "github.com/forestrie/go-merklelog-azure/tests/storage"
	"github.com/forestrie/go-merklelog-provider-testing/mmrtesting"
	"github.com/stretchr/testify/require"
)

type TestContext struct {
	azstoragetesting.TestContext
	LastTime           time.Time
	numEventsGenerated int
}

func NewDefaultTestContext(t *testing.T, opts ...massifs.Option) *TestContext {
	tc := azstoragetesting.NewDefaultTestContext(t, opts...)
	return &TestContext{
		TestContext: *tc,
		LastTime:    time.Now(),
	}
}

func NewLogBuilderFactory(tc *TestContext) mmrtesting.LogBuilder {
	builder := azstoragetesting.NewLogBuilder(&tc.TestContext)
	builder.LeafGenerator = mmrtesting.NewDataTrailsLeafGenerator(tc.GetG())
	return builder
}

func CreateLogBuilderContext(t *testing.T, massifHeight uint8, massifCount uint32, opts ...massifs.Option) (*TestContext, storage.LogID, mmrtesting.LogBuilder, mmrtesting.GeneratedLeaves) {

	tc := NewDefaultTestContext(t, opts...)
	logID := tc.G.NewLogID()
	builder, generated := CreateLogForContext(tc, logID, massifHeight, massifCount)
	return tc, logID, builder, generated
}

func CreateLogContext(t *testing.T, massifHeight uint8, massifCount uint32, opts ...massifs.Option) (*TestContext, storage.LogID) {

	tc, logID, _, _ := CreateLogBuilderContext(t, massifHeight, massifCount, opts...)
	return tc, logID
}

func CreateLogForContext(tc *TestContext, logID storage.LogID, massifHeight uint8, massifCount uint32) (mmrtesting.LogBuilder, mmrtesting.GeneratedLeaves) {

	builder := NewLogBuilderFactory(tc)
	tc.DeleteLog(logID)
	err := builder.SelectLog(tc.T.Context(), logID)
	require.NoError(tc.T, err)
	generated, err := tc.CreateLog(tc.T.Context(), builder, logID, massifHeight, massifCount)
	if err != nil {
		tc.T.Fatalf("CreateLog failed: %v", err)
	}
	return builder, generated
}

func CreateLogsForContext(tc *TestContext, massifHeight uint8, massifCount uint32, logIDs ...storage.LogID) ([]mmrtesting.GeneratedLeaves, []mmrtesting.LogBuilder) {

	var generated []mmrtesting.GeneratedLeaves
	var builders []mmrtesting.LogBuilder

	for _, logID := range logIDs {
		builder := NewLogBuilderFactory(tc)
		tc.DeleteLog(logID)
		err := builder.SelectLog(tc.T.Context(), logID)
		if err != nil {
			tc.T.Fatalf("SelectLog failed: %v", err)
		}
		builders = append(builders, builder)
		gen, err := tc.CreateLog(tc.T.Context(), builder, logID, massifHeight, massifCount)
		if err != nil {
			tc.T.Fatalf("CreateLog failed: %v", err)
		}
		generated = append(generated, gen)
	}
	return generated, builders
}
