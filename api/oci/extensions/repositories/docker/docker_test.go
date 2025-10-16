//go:build docker_test

package docker_test

import (
	"github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/docker"
)

var _ = Describe("Local Docker Daemon", func() {
	It("validated access", func() {
		octx := oci.DefaultContext()
		spec := docker.NewRepositorySpec()
		testutils.MustBeSuccessful(spec.Validate(octx, nil))
	})
})
