package signingtest

import (
	"ocm.software/ocm/api/helper/env"
)

func TestData(dest ...string) env.Option {
	return env.ProjectTestDataForCaller("testdata", dest...)
}

func ModifiableTestData(dest ...string) env.Option {
	return env.ModifiableProjectTestDataForCaller("testdata", dest...)
}
