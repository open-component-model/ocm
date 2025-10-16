package main

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"helm.sh/helm/v3/pkg/chart"
	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/ocmutils/localize"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
	"ocm.software/ocm/api/tech/helm/loader"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/tarutils"
	"ocm.software/ocm/api/utils/template"
	"ocm.software/ocm/examples/lib/helper"
	"sigs.k8s.io/yaml"
)

type DeployDescriptor struct {
	ChartResource v1.ResourceReference   `json:"chartResource"`
	Values        json.RawMessage        `json:"values"`
	Templater     string                 `json:"templater"`
	ImageMappings localize.ImageMappings `json:"imageMappings"`
}

func EvaluateDeployDescriptor(cv ocm.ComponentVersionAccess, res ocm.ResourceAccess, resolver ocm.ComponentVersionResolver,
	config template.Values, path string, fss ...vfs.FileSystem,
) ([]byte, *chart.Chart, string, error) {
	fs := utils.FileSystem(fss...)
	id := res.Meta().GetIdentity(cv.GetDescriptor().Resources)

	// first, get deploy descriptor
	data, err := ocmutils.GetResourceData(res)
	if err != nil {
		return evalErr(err, "cannot get deploy descriptor from resource %s", id)
	}

	var desc DeployDescriptor
	err = yaml.Unmarshal(data, &desc)
	if err != nil {
		return evalErr(err, "cannot unmarshal deploy descriptor from resource %s", id)
	}

	// second, determine substitutions
	mappings, err := localize.LocalizeMappings(desc.ImageMappings, cv, resolver)
	if err != nil {
		return evalErr(err, "image localization failed")
	}

	values, err := localize.SubstituteMappingsForData(mappings, desc.Values)
	if err != nil {
		return evalErr(err, "applying substitutions")
	}

	if desc.Templater == "" {
		desc.Templater = "merge"
	}
	if config == nil {
		config = map[string]interface{}{}
	}
	templater, err := template.DefaultRegistry().Create(desc.Templater, fs)
	if err != nil {
		if desc.Templater == "merge" {
			templater = template.NewMerge()
		} else {
			return evalErr(err, "")
		}
	}

	// third, process values with config
	effvalues, err := templater.Process(string(values), config)
	if err != nil {
		return evalErr(err, "error templating with %q", desc.Templater)
	}

	// fourth, get the helm chart
	cres, ccv, err := resourcerefs.ResolveResourceReference(cv, desc.ChartResource, resolver)
	if err != nil {
		return evalErr(err, "cannot resolve chart resource %s", desc.ChartResource)
	}
	effpath, err := download.DownloadResource(cv.GetContext(), cres, path, download.WithFileSystem(fs))
	ccv.Close()
	if err != nil {
		return evalErr(err, "cannot download chart resource %s", desc.ChartResource)
	}
	chart, err := loader.Load(effpath, fs)
	if err != nil {
		return evalErr(err, "cannot load helm chart")
	}
	return []byte(effvalues), chart, effpath, nil
}

func evalErr(err error, msg string, args ...interface{}) ([]byte, *chart.Chart, string, error) {
	return nil, nil, "", errors.Wrapf(err, msg, args...)
}

func Localize(cfg *helper.Config, config template.Values, release, namespace string) error {
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

	res, err := cv.SelectResources(rscsel.ArtifactType(DEPLOY_SCRIPT_TYPE))
	if err != nil {
		return errors.Wrapf(err, "no deploy descriptor found")
	}

	fs := memoryfs.New()
	values, chart, path, err := EvaluateDeployDescriptor(cv, res[0], nil, config, "chart", fs)
	if err != nil {
		return errors.Wrapf(err, "invalid deployment")
	}

	////////////////////////////////////////////////////////////////////////////
	// print chart info

	fmt.Printf("downloaded chart archive: %s\n", path)
	files, err := tarutils.ListArchiveContent(path, fs)
	if err != nil {
		return errors.Wrapf(err, "cannot list chart files")
	}
	fmt.Printf("chart files:\n")
	for _, f := range files {
		fmt.Printf("  - %s\n", f)
	}
	desc, _ := runtime.ToYAML(chart.Metadata)
	fmt.Printf("chart meta data:\n%s\n", utils.IndentLines(string(desc), "    ", false))

	yamlvalues, err := runtime.ToYAML(values)
	if err != nil {
		return errors.Wrapf(err, "invalid values")
	}
	fmt.Printf("localized and configured values:\n%s\n", utils.IndentLines(string(yamlvalues), "    ", false))

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// installation

	if release != "" {
		if namespace == "" {
			namespace = "default"
		}
		err := InstallChart(chart, release, namespace)
		if err != nil {
			return errors.Wrapf(err, "installation failed")
		}
	}
	return nil
}
