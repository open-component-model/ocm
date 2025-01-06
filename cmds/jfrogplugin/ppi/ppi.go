package ppi

import (
	"fmt"
	"strconv"

	ocmconfig "ocm.software/ocm/api/config"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/blobhandler"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/jfrogplugin/config"
	"ocm.software/ocm/cmds/jfrogplugin/uploaders/helm"
)

const NAME = "jfrog"

func Plugin() (ppi.Plugin, error) {
	p := ppi.NewPlugin(NAME, version.Get().String())

	p.SetShort(NAME + " plugin")
	p.SetLong(`ALPHA GRADE plugin providing custom functions related to interacting with JFrog Repositories (e.g. Artifactory).

This plugin is solely for interacting with JFrog Servers and cannot be used for generic repository types.
Thus, you should only consider this plugin if
- You need to use a JFrog specific API
- You cannot use any of the generic (non-jfrog) implementations.

If given an OCI Artifact Set (for example by using it on a resource with a Helm Chart backed by an OCI registry),
it will do a best effort conversion to a normal helm chart and upload that in its stead. Note that this conversion
is not perfect however, since the Upload will inevitably strip provenance information from the chart.
This can lead to unintended side effects such as
- Having the wrong digest in the resource access
- Losing the ability to convert back to an OCI artifact set without changing digests and losing provenance information.
This means that effectively you should try to migrate to pure OCI registries instead of JFrog HELM repositories as soon
as possible (this uploader is just a stop gap).

Examples:

You can configure the JFrog plugin as an Uploader in an ocm config file with:

- type: ` + fmt.Sprintf("%s.ocm.%s", plugin.KIND_UPLOADER, ocmconfig.OCM_CONFIG_TYPE_SUFFIX) + `
  registrations:
  - name: ` + fmt.Sprintf("%s/%s/%s", plugin.KIND_PLUGIN, NAME, helm.NAME) + `
    artifactType: ` + artifacttypes.HELM_CHART + `
    priority: 200 # must be > ` + strconv.Itoa(blobhandler.DEFAULT_BLOBHANDLER_PRIO) + ` to be used over the default handler
    config:
      type: ` + fmt.Sprintf("%s/%s", helm.NAME, helm.VERSION) + `
      # this is only a sample JFrog Server URL, do NOT append /artifactory
      url: int.repositories.ocm.software 
      repository: ocm-helm-test
`)
	p.SetConfigParser(config.GetConfig)
	p.ForwardLogging(true)

	u := helm.New()
	if err := p.RegisterUploader(artifacttypes.HELM_CHART, "", u); err != nil {
		return nil, err
	}

	return p, nil
}
