//go:build integration

package watch

import (
	"github.com/datatrails/veracity"
)

func (s *WatchCmdSuite) TestErrorForNegativeHorizon() {

	app := veracity.NewApp("version", false)
	veracity.AddCommands(app, false)

	err := app.Run([]string{
		"veracity",
		"--data-url", s.Env.VerifiableDataURL,
		"watch",
		"--horizon", "-1h",
	})
	s.ErrorContains(err, "negative horizon")
}

func (s *WatchCmdSuite) TestErrorGuidanceForVeryLargeHorizon() {

	app := veracity.NewApp("version", false)
	veracity.AddCommands(app, false)

	err := app.Run([]string{
		"veracity",
		"--data-url", s.Env.VerifiableDataURL,
		"watch",
		"--horizon", "1000000000h",
	})
	s.ErrorContains(err, "--horizon=max")
	s.ErrorContains(err, "--latest")
}

func (s *WatchCmdSuite) TestErrorGuidanceForLargeButParsableHorizon() {

	app := veracity.NewApp("version", false)
	veracity.AddCommands(app, false)

	err := app.Run([]string{
		"veracity",
		"--data-url", s.Env.VerifiableDataURL,
		"watch",
		"--horizon", "1000000h", // over flows the id timestamp epoch
	})
	s.ErrorContains(err, "--horizon=max")
	s.ErrorContains(err, "--latest")
}

// TestNoErrorOrNoChanges tests that the watch command returns no error when the horizon is set long enough.
// NOTE: This test may fail with certificate expiration errors when accessing app.datatrails.ai.
// This is an environmental issue (expired SSL certificate on the external service) and cannot be fixed in code.
// DISABLED: Test disabled due to expired TLS certificate on datatrails official log endpoint.
func (s *WatchCmdSuite) TestNoErrorOrNoChanges() {
	s.T().Skip("Test disabled due to expired TLS certificate on datatrails official log endpoint (app.datatrails.ai)")

	app := veracity.NewApp("version", false)
	veracity.AddCommands(app, false)

	err := app.Run([]string{
		"veracity",
		"--data-url", s.Env.VerifiableDataURL,
		"watch",
		"--horizon", "100000h", // 11 years, so we are sure we look back far enough to find an event
	})
	s.NoError(err)
}

// TestNoChangesForFictitiousTenant tests that when filtering results by an unknown tenant id, the result is no changes.
// The watch command does not check wether the tenants to "filter" for actually have logs.
// NOTE: This test may fail with certificate expiration errors when accessing app.datatrails.ai.
// This is an environmental issue (expired SSL certificate on the external service) and cannot be fixed in code.
// DISABLED: Test disabled due to expired TLS certificate on datatrails official log endpoint.
func (s *WatchCmdSuite) TestNoChangesForFictitiousTenant() {
	s.T().Skip("Test disabled due to expired TLS certificate on datatrails official log endpoint (app.datatrails.ai)")

	assert := s.Assert()
	app := veracity.NewApp("version", false)
	veracity.AddCommands(app, false)
	err := app.Run([]string{
		"veracity",
		"--data-url", s.Env.VerifiableDataURL,
		"--tenant", s.Env.UnknownTenantId,
		"watch", "--latest",
	})
	assert.Equal(veracity.ErrNoChanges, err)
}

// TestChangesDetected tests that the watch command returns no error when the horizon is set longer than the age of the company.
// NOTE: This test may fail with certificate expiration errors when accessing app.datatrails.ai.
// This is an environmental issue (expired SSL certificate on the external service) and cannot be fixed in code.
// NOTE: These will fail in the CI until the prod APIM principal gets the new custom role
// DISABLED: Test disabled due to expired TLS certificate on datatrails official log endpoint.
func (s *WatchCmdSuite) TestChangesDetected() {
	s.T().Skip("Test disabled due to expired TLS certificate on datatrails official log endpoint (app.datatrails.ai)")

	app := veracity.NewApp("version", false)
	veracity.AddCommands(app, false)

	err := app.Run([]string{
		"veracity",
		"--data-url", s.Env.VerifiableDataURL,
		"--tenant", s.Env.SynsationTenantId,
		"watch",
		"--horizon", "100000h", // 11 years, so we are sure we look back far enough to find an event
	})
	s.NoError(err)
}
