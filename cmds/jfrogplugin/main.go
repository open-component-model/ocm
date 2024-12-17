package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/blobhandler"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/version"
	"ocm.software/ocm/cmds/jfrogplugin/uploaders/helm"
)

const NAME = "jfrog"

func main() {
	p := ppi.NewPlugin(NAME, version.Get().String())

	p.SetShort(NAME + " plugin")
	p.SetLong(`ALPHA GRADE plugin providing custom functions related to interacting with JFrog Repositories (e.g. Artifactory).

This plugin is solely for interacting with JFrog Servers and cannot be used for generic repository types.
Thus, you should only consider this plugin if
- You need to use a JFrog specific API
- You cannot use any of the generic (non-jfrog) implementations.

Examples:

You can configure the JFrog plugin as an Uploader in an ocm config file with:

- type: ` + fmt.Sprintf("%s.ocm.%s", plugin.KIND_UPLOADER, config.OCM_CONFIG_TYPE_SUFFIX) + `
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
	p.SetConfigParser(GetConfig)

	u := helm.New()
	if err := p.RegisterUploader(artifacttypes.HELM_CHART, "", u); err != nil {
		panic(err)
	}
	err := cmds.NewPluginCommand(p).Execute(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while running plugin: %v\n", err)
		os.Exit(1)
	}
}

type Config struct {
}

func GetConfig(raw json.RawMessage) (interface{}, error) {
	var cfg Config

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("could not get config: %w", err)
	}
	return &cfg, nil
}
