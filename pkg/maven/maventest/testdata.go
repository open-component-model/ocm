package maventest

import (
	"github.com/open-component-model/ocm/pkg/env"
)

func TestData(dest ...string) env.Option {
	return env.ProjectTestDataForCaller(dest...)
}
