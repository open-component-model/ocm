// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	ocmutils "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/localize"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
	"sigs.k8s.io/yaml"
)

type DeployDescriptor struct {
	ChartResource v1.ResourceReference   `json:"chartResource"`
	Values        json.RawMessage        `json:"values"`
	ImageMappings localize.ImageMappings `json:"imageMappings"`
}

func EvaluateDeployDescriptor(cv ocm.ComponentVersionAccess, res ocm.ResourceAccess, resolver ocm.ComponentVersionResolver, path string, fss ...vfs.FileSystem) ([]byte, string, error) {
	id := res.Meta().GetIdentity(cv.GetDescriptor().Resources)

	// first, get deploy descriptor
	data, err := ocmutils.GetResourceData(res)
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot get deploy descriptor from resource %s", id)
	}

	var desc DeployDescriptor
	err = yaml.Unmarshal(data, &desc)
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot unmarshal deploy descriptor from resource %s", id)
	}

	// second, determine substitutions
	mappings, err := localize.LocalizeMappings(desc.ImageMappings, cv, resolver)
	if err != nil {
		return nil, "", errors.Wrapf(err, "image localization failed")
	}

	values, err := localize.SubstituteMappingsForData(mappings, desc.Values)
	if err != nil {
		return nil, "", errors.Wrapf(err, "applying substitutions")
	}

	// third, get the helm chart
	cres, ccv, err := ocmutils.ResolveResourceReference(cv, desc.ChartResource, resolver)
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot resolve chart resource %s", desc.ChartResource)
	}
	effpath, err := download.DownloadResource(cv.GetContext(), cres, path, download.WithFileSystem(utils.FileSystem(fss...)))
	ccv.Close()
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot download chart resource %s", desc.ChartResource)
	}
	return values, effpath, nil
}

func Localize(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()
	err := ReadConfiguration(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot read ocm configuration")
	}

	// use the generic form here to enable the specification of any
	// supported repository type as target.
	fmt.Printf("local repository is %s\n", string(cfg.Target))
	repo, err := ctx.RepositoryForConfig(cfg.Target, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot open local repository")
	}
	defer repo.Close()

	// lookup component in local repo
	cv, err := repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION)
	if err != nil {
		return errors.Wrapf(err, "cannot get component version from %s", cfg.Target)
	}
	defer cv.Close()

	res, err := cv.GetResourcesByResourceSelectors(compdesc.ByResourceType(DEPLOY_SCRIPT_TYPE))
	if err != nil {
		return errors.Wrapf(err, "no deploy descriptor found")
	}

	fs := memoryfs.New()
	values, path, err := EvaluateDeployDescriptor(cv, res[0], nil, "chart", fs)
	if err != nil {
		return errors.Wrapf(err, "invalid deployment")
	}
	yamlvalues, err := runtime.ToYAML(values)
	if err != nil {
		return errors.Wrapf(err, "invalid values")
	}
	fmt.Printf("localized values:\n%s\n", utils.IndentLines(string(yamlvalues), "    ", false))
	files, err := ListFiles(path, fs)
	if err != nil {
		return errors.Wrapf(err, "cannot list chart files")
	}
	fmt.Printf("chart files:\n")
	for _, f := range files {
		fmt.Printf("  - %s\n", f)
	}
	return nil
}
