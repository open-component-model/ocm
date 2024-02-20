package helm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	me "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/helm"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
)

var _ = Describe("Test Environment", func() {

	Context("https://github.com/open-component-model/ocm/issues/648", func() {
		var env *TestEnv

		BeforeEach(func() {
			env = NewTestEnv(TestData())
		})

		AfterEach(func() {
			env.Cleanup()
		})

		FIt("creates ctf and adds component", func() {
			/*
				# Create component archive
				ocm create ca --file ca --scheme ocm.software/v3alpha1 --provider test.com test.com/test 1.0.0

				# Pull Helm chart (normally this will be a Helm chart that is build in a pipeline
				helm pull oci://registry-1.docker.io/bitnamicharts/postgresql --version 14.0.5

				# Add resource (notice the full path here, because that is causing the issue)
				ocm add resources --file ca --name postgresql --type helmChart --inputType helm --inputPath E:\t\bugrepo\postgresql-14.0.5.tgz --inputVersion 14.0.5
			*/

			Expect(env.Execute("add", "resources", "--file", "ca", "--name", "postgresql", "--type", "helmChart", "--inputType", me.TYPE, "--inputPath", `E:\t\bugrepo\postgresql-14.0.5.tgz`, "--inputVersion", "14.0.5")).To(Succeed())
		})
	})
})
