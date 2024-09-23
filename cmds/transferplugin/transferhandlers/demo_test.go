package transferhandlers_test

import (
	"bytes"
	"encoding/json"

	"github.com/mandelsoft/goutils/sliceutils"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	v2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/plugin/common"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/transferplugin/app"
)

var _ = Describe("Test Environment", func() {
	It("", func() {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		config := Must(vfs.ReadFile(osfs.OsFs, "testdata/config"))

		question := &ppi.ArtifactQuestion{
			Source: ppi.SourceComponentVersion{
				Name:       "",
				Version:    "",
				Provider:   metav1.Provider{},
				Repository: ocm.GenericRepositorySpec{},
				Labels:     nil,
			},
			Artifact: ppi.Artifact{
				Meta:   v2.ElementMeta{},
				Access: ocm.GenericAccessSpec{},
				AccessInfo: ppi.AccessInfo{
					Kind: ociartifact.Type,
					Host: "ghcr.io",
					Port: "",
					Path: "",
					Info: "",
				},
			},
			Options: ppi.TransferOptions{},
		}

		in := Must(json.Marshal(question))

		app.Run(sliceutils.AsSlice("--config", string(config), "transferhandler", "demo", ppi.Q_TRANSFER_RESOURCE), cmds.StdIn(bytes.NewBuffer(in)), cmds.StdOut(&stdout), cmds.StdErr(&stderr))
		Expect(stdout.Bytes()).To(YAMLEqual(`
decision: true
`))
	})

	It("handles empty list", func() {
		b := ppi.NewDecisionHandlerBase("x", "")
		Expect(b.GetLabels()).NotTo(BeNil())
	})

	It("describes plugin", func() {
		d := Must(app.New()).Descriptor()
		p, out := misc.NewBufferedPrinter()
		common.DescribePluginDescriptorCapabilities(nil, &d, p)
		Expect(out.String()).To(StringEqualTrimmedWithContext(utils.Crop(`
  Capabilities:     Transfer Handlers
  Description: 
        plugin providing a transfer handler to enable value transport for dedicated external repositories.

  Transfer Handlers:
  - Name: demo
      enable value transport for dedicated external repositories
    Questions:
    - Name: transferresource
        value transport only for dedicated access types and service hosts
      consumes no labels
    - Name: transfersource
        value transport only for dedicated access types and service hosts
      consumes no labels

`, 2)))
	})
})
