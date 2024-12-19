package app

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/mandelsoft/goutils/exception"
	. "github.com/mandelsoft/goutils/finalizer"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/download"
	utils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/ocm/tools/toi/support"
	"ocm.software/ocm/api/tech/helm/loader"
	"ocm.software/ocm/api/utils/compression"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/tarutils"
	"ocm.software/ocm/cmds/helminstaller/app/driver"
)

func Merge(values ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	for _, val := range values {
		for k, v := range val {
			result[k] = v
		}
	}
	return result
}

type Execution struct {
	driver driver.Driver
	*support.ExecutorOptions
	path string
	fs   vfs.FileSystem
}

func (e *Execution) outf(msg string, args ...interface{}) {
	out.Outf(e.OutputContext, msg, args...)
}

func (e *Execution) unpackChart(dir string) {
	e.outf("Unpacking chart archive to %s...\n", dir)
	e.Logger.Debug("unpacking chart archive", "directory", dir)
	r := Must1f(R1(e.fs.Open(e.path)), "cannot read downloaded chart archive %q", e.path)
	defer r.Close()

	e.Logger.Debug("auto decompress downloaded chart")
	reader, _ := Must2f(R2(compression.AutoDecompress(r)), "cannot uncompress downloaded chart archive %q", e.path)
	e.Logger.Debug("preparing chart filesystem")
	chartfs := Must1f(R1(projectionfs.New(e.fs, dir)), "cannot create projection %q", e.path)
	e.Logger.Debug("extracting chart archive", "archive", e.path)
	Mustf(tarutils.ExtractTarToFs(chartfs, reader), "cannot extract downloaded chart archive %q", e.path)
	e.Logger.Debug("lookup chart folder")
	entries := Must1f(R1(vfs.ReadDir(e.fs, dir)), "cannot find chart folder in %q", dir)
	if len(entries) != 1 {
		Throw(fmt.Errorf("expected single chart folder in archive, but found %d folders", len(entries)))
	}
	e.Logger.Debug("found chart folder", "folder", entries[0].Name())
	e.path = filepath.Join(dir, entries[0].Name())
}

func (e *Execution) addSubCharts(finalize *Finalizer, subCharts map[string]v1.ResourceReference) {
	dir := e.path + ".dir"
	Mustf(e.fs.Mkdir(dir, 0o700), "cannot mkdir %q", dir)
	finalize.With(Calling1(e.fs.RemoveAll, dir))

	e.unpackChart(dir)

	// prepare dependencies
	e.outf("Preparing dependencies...\n")
	chartFile := filepath.Join(e.path, "Chart.yaml")
	chartData := Must1f(R1(vfs.ReadFile(e.fs, chartFile)), "cannot read Chart.yaml")

	var chart map[string]interface{}
	Mustf(runtime.DefaultYAMLEncoding.Unmarshal(chartData, &chart), "cannot parse Chart.yaml")
	deps := []interface{}{}
	if d := chart["dependencies"]; d != nil {
		if a, ok := d.([]interface{}); ok {
			deps = a
		}
	}

	var loop Finalizer
	defer loop.Finalize()

	charts := filepath.Join(e.path, "charts")
	Mustf(e.fs.Mkdir(charts, 0o700), "cannot mkdir %q", charts)
	e.outf("Loading %d sub charts into %s...\n", len(subCharts), charts)
	for n, r := range subCharts {
		e.outf("  Loading sub chart %q from resource %s@%s\n", n, r, common.VersionedElementKey(e.ComponentVersion))
		acc, rcv := Must2f(R2(resourcerefs.ResolveResourceReference(e.ComponentVersion, r, nil)), "chart reference", r.String())
		loop.Close(rcv)

		if acc.Meta().Type != resourcetypes.HELM_CHART {
			Throw(errors.Newf("%s: resource type %q required, but found %q", r, resourcetypes.HELM_CHART, acc.Meta().Type))
		}

		_, subpath := Must2f(R2(download.For(e.Context).Download(common.NewPrinter(e.OutputContext.StdOut()), acc, filepath.Join(charts, n), e.fs)), "downloading helm chart %s", r)

		chartObj := Must1f(R1(loader.Load(subpath, e.fs)), "cannot load subchart %q", subpath)
		found := false
		for _, dep := range deps {
			m, ok := dep.(map[string]interface{})
			if ok {
				if m["alias"] == n {
					e.outf("    found dependency %q for subchart %q\n", n, chartObj.Name())
					m["name"] = chartObj.Name()
					found = true
					break
				}
				if m["name"] == chartObj.Name() {
					if m["alias"] == nil {
						e.outf("    setting alias %q for dependency for subchart %q\n", n, chartObj.Name())
						if n != chartObj.Name() {
							m["alias"] = n
						}
						found = true
					}
				}
			}
		}
		if !found {
			e.outf("    adding dependency %q for subchart %q\n", n, chartObj.Name())
			m := map[string]interface{}{}
			m["name"] = chartObj.Name()
			if n != chartObj.Name() {
				m["alias"] = n
			}
			deps = append(deps, m)
		}
		loop.Finalize()
	}

	chart["dependencies"] = deps
	chartData = Must1f(R1(runtime.DefaultYAMLEncoding.Marshal(chart)), "cannot marshal Chart.yaml")
	Mustf(vfs.WriteFile(e.fs, chartFile, chartData, 0o600), "cannot write Chart.yaml")
}

func (e *Execution) Execute(cfg *Config, values map[string]interface{}, kubeconfig []byte) (err error) {
	var finalize Finalizer
	defer finalize.CatchException().FinalizeWithErrorPropagation(&err)

	if e.Action != "install" && e.Action != "uninstall" {
		return errors.ErrNotSupported("action", e.Action)
	}

	values = Merge(Must1(cfg.GetValues()), values)

	e.outf("Loading helm chart from resource %s@%s\n", cfg.Chart, common.VersionedElementKey(e.ComponentVersion))
	acc, rcv := Must2f(R2(resourcerefs.ResolveResourceReference(e.ComponentVersion, cfg.Chart, nil)), "chart reference", cfg.Chart.String())
	finalize.Close(rcv)

	if acc.Meta().Type != resourcetypes.HELM_CHART {
		return errors.Newf("resource type %q required, but found %q", resourcetypes.HELM_CHART, acc.Meta().Type)
	}

	// have to use the OS filesystem here for using the helm library
	file := Must1(vfs.TempFile(e.fs, "", "helm-*"))
	path := file.Name()
	file.Close()
	e.fs.Remove(path)

	spec := Must1f(R1(acc.Access()), "getting access specification")
	data, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	e.Logger.Info("starting download", "path", path, "access", string(data))

	_, e.path = Must2f(R2(download.For(e.Context).Download(common.NewPrinter(e.OutputContext.StdOut()), acc, path, e.fs)), "downloading helm chart")

	finalize.With(Calling1(e.fs.Remove, e.path))

	if len(cfg.SubCharts) > 0 {
		e.addSubCharts(&finalize, cfg.SubCharts)
	}

	e.outf("Localizing helm chart...\n")
	e.Logger.Debug("Localizing helm chart")
	for i, v := range cfg.ImageMapping {
		acc, rcv := Must2f(R2(resourcerefs.ResolveResourceReference(e.ComponentVersion, v.ResourceReference, nil)), "mapping", fmt.Sprintf("%d (%s)", i+1, &v.ResourceReference))
		rcv.Close()
		ref := Must1f(R1(utils.GetOCIArtifactRef(e.Context, acc)), "mapping %d: cannot resolve resource %s to an OCI Reference", i+1, v)
		ix := strings.Index(ref, ":")
		if ix < 0 {
			ix = strings.Index(ref, "@")
			if ix < 0 {
				return errors.Wrapf(err, "mapping %d: image tag or digest missing (%s)", i+1, ref)
			}
		}
		repo := ref[:ix]
		tag := ref[ix+1:]
		if v.Repository != "" {
			e.Logger.Debug("substitute image repository", "ref", ref, "target", v.Repository)
			Mustf(Set(values, v.Repository, repo), "mapping %d: assigning repository to property %q", v.Repository)
		}
		if v.Tag != "" {
			e.Logger.Debug("substitute image tag", "ref", ref, "target", v.Tag)
			Mustf(Set(values, v.Tag, tag), "mapping %d: assigning tag to property %q", v.Tag)
		}
		if v.Image != "" {
			e.Logger.Debug("substitute image ref", "ref", ref, "target", v.Image)
			Mustf(Set(values, v.Image, ref), "mapping %d: assigning image to property %q", v.Image)
		}
	}

	e.outf("Installing helm chart [%s]...\n", e.Action)

	ns := "default"
	if cfg.Namespace != "" {
		ns = cfg.Namespace
	}
	if s, ok := values["namespace"].(string); ok && s != "" {
		ns = s
	}
	release := cfg.Release
	if s, ok := values["release"].(string); ok && s != "" {
		release = s
	}
	e.Logger.Debug("executing helm deployment", "action", e.Action, "namespace", ns, "release", release)
	valuesdata := Must1f(R1(runtime.DefaultYAMLEncoding.Marshal(values)), "marshal values")

	dcfg := &driver.Config{
		ChartPath:       e.path,
		Release:         release,
		Namespace:       ns,
		CreateNamespace: cfg.CreateNamespace,
		Values:          valuesdata,
		Kubeconfig:      kubeconfig,
		Output:          e.OutputContext.StdOut(),
		Debug:           e.Logger,
	}
	switch e.Action {
	case "install":
		return e.driver.Install(dcfg)
	case "uninstall":
		return e.driver.Uninstall(dcfg)
	default:
		return errors.ErrNotImplemented("action", e.Action)
	}
}
