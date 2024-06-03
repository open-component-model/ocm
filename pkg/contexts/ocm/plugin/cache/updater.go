package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/extraid"
	"github.com/open-component-model/ocm/pkg/semverutils"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

type PluginSource struct {
	Repository *cpi.GenericRepositorySpec `json:"repository"`
	Component  string                     `json:"component"`
	Version    string                     `json:"version"`
	Resource   string                     `json:"resource"`
}

type PluginUpdater struct {
	Context     ocm.Context
	Force       bool
	RemoveMode  bool
	UpdateMode  bool
	Describe    bool
	Constraints []*semver.Constraints

	Current string
	Printer common.Printer
}

func NewPluginUpdater(ctx ocm.ContextProvider, printer common.Printer) *PluginUpdater {
	return &PluginUpdater{
		Context: ctx.OCMContext(),
		Printer: common.AssurePrinter(printer),
	}
}

func (o *PluginUpdater) SetupCurrent(name string) error {
	dir := plugindirattr.Get(o.Context)
	if dir == "" {
		return fmt.Errorf("no plugin dir configured")
	}
	src, err := ReadPluginSource(dir, name)
	if err != nil {
		return nil
	}
	o.Current = src.Version
	return nil
}

func (o *PluginUpdater) Remove(session ocm.Session, name string) error {
	dir := plugindirattr.Get(o.Context)
	if dir == "" {
		return fmt.Errorf("no plugin dir configured")
	}
	if err := RemovePluginSource(dir, name); err != nil {
		return errors.Wrapf(err, "cannot remove source info for plugin %q", name)
	}
	file := filepath.Join(dir, name)
	if ok, err := vfs.FileExists(osfs.New(), file); !ok && err == nil {
		return fmt.Errorf("plugin %s not found", name)
	}
	if err := RemoveFile(file); err != nil {
		return errors.Wrapf(err, "cannot remove plugin %q", name)
	}
	o.Printer.Printf("plugin %s removed\n", name)
	return nil
}

func (o *PluginUpdater) Update(session ocm.Session, name string) error {
	dir := plugindirattr.Get(o.Context)
	if dir == "" {
		return fmt.Errorf("no plugin dir configured")
	}
	src, err := ReadPluginSource(dir, name)
	if err != nil {
		return err
	}
	o.Current = src.Version
	repo, err := session.LookupRepository(o.Context, src.Repository)
	if err != nil {
		return err
	}
	comp, err := session.LookupComponent(repo, src.Component)
	if err != nil {
		return err
	}
	return o.downloadLatest(session, comp, src.Resource)
}

func (o *PluginUpdater) DownloadRef(session ocm.Session, ref string, name string) error {
	result, err := session.EvaluateVersionRef(o.Context, ref)
	if err != nil {
		return err
	}
	if result.Component == nil {
		return fmt.Errorf("component required")
	}
	if result.Version == nil {
		return o.downloadLatest(session, result.Component, name)
	}
	if result.Version.GetVersion() == o.Current && !o.Force {
		o.Printer.Printf("plugin %s already uptodate\n", name)
		return nil
	}
	return o.download(session, result.Version, name)
}

func (o *PluginUpdater) DownloadFromRepo(session ocm.Session, repo ocm.Repository, ref, name string) error {
	cr, err := ocm.ParseComp(ref)
	if err != nil {
		return err
	}

	var cv ocm.ComponentVersionAccess
	comp, err := session.LookupComponent(repo, cr.Component)
	if err != nil {
		return err
	}
	if cr.IsVersion() {
		if *cr.Version == o.Current && !o.Force {
			o.Printer.Printf("plugin %s already uptodate\n", name)
			return nil
		}
		cv, err = session.GetComponentVersion(comp, *cr.Version)
		if err != nil {
			return err
		}
		return o.download(session, cv, name)
	}
	return o.downloadLatest(session, comp, name)
}

func (o *PluginUpdater) downloadLatest(session ocm.Session, comp ocm.ComponentAccess, name string) error {
	var vers []string

	vers, err := comp.ListVersions()
	if err != nil {
		return errors.Wrapf(err, "cannot list versions for component %s", comp.GetName())
	}
	if len(vers) == 0 {
		return errors.Wrapf(err, "no versions found for component %s", comp.GetName())
	}

	versions, err := semverutils.MatchVersionStrings(vers, o.Constraints...)
	if err != nil {
		return fmt.Errorf("failed to match version strings for component %s: %w", comp.GetName(), err)
	}
	if len(versions) == 0 {
		return fmt.Errorf("no versions for component %s match the constraints", comp.GetName())
	}
	if len(versions) > 1 {
		versions = versions[len(versions)-1:]
	}
	if versions[0].Original() == o.Current && !o.Force {
		o.Printer.Printf("plugin %s already uptodate\n", name)
		return nil
	}
	cv, err := session.GetComponentVersion(comp, versions[0].Original())
	if err != nil {
		return err
	}
	return o.download(session, cv, name)
}

func (o *PluginUpdater) download(session ocm.Session, cv ocm.ComponentVersionAccess, name string) (err error) {
	defer errors.PropagateErrorf(&err, nil, "%s", common.VersionedElementKey(cv))

	var found ocm.ResourceAccess
	var wrong ocm.ResourceAccess
	for _, r := range cv.GetResources() {
		if name != "" && r.Meta().Name != name {
			continue
		}
		if r.Meta().Type == "ocmPlugin" {
			if r.Meta().ExtraIdentity.Get(extraid.ExecutableOperatingSystem) == runtime.GOOS &&
				r.Meta().ExtraIdentity.Get(extraid.ExecutableArchitecture) == runtime.GOARCH {
				found = r
				break
			}
			wrong = r
		} else {
			if name != "" {
				wrong = r
			}
		}
	}
	if found == nil {
		if wrong != nil {
			if wrong.Meta().Type != "ocmPlugin" {
				return fmt.Errorf("resource %q has wrong type: %s", wrong.Meta().Name, wrong.Meta().Type)
			}
			return fmt.Errorf("os %s architecture %s not found for resource %q", runtime.GOOS, runtime.GOARCH, wrong.Meta().Name)
		}
		if name != "" {
			return fmt.Errorf("resource %q not found", name)
		}
		return fmt.Errorf("no ocmPlugin found")
	}
	o.Printer.Printf("found resource %s[%s]\n", found.Meta().Name, found.Meta().ExtraIdentity.String())

	file, err := os.CreateTemp(os.TempDir(), "plugin-*")
	if err != nil {
		return errors.Wrapf(err, "cannot create temp file")
	}
	file.Close()
	fs := osfs.New()
	_, _, err = download.For(o.Context).Download(o.Printer, found, file.Name(), fs)
	if err != nil {
		return errors.Wrapf(err, "cannot download resource %s", found.Meta().Name)
	}

	desc, err := GetPluginInfo(file.Name())
	if err != nil {
		return err
	}
	if o.Describe {
		data, err := yaml.Marshal(desc)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal plugin descriptor")
		}
		o.Printer.Printf("%s", string(data))
	} else {
		err := o.SetupCurrent(desc.PluginName)
		if err != nil {
			return err
		}
		if cv.GetVersion() == o.Current {
			o.Printer.Printf("version %s already installed\n", o.Current)
			if !o.Force {
				return nil
			}
		}
		dir := plugindirattr.Get(o.Context)
		if dir != "" {
			target := filepath.Join(dir, desc.PluginName)

			verb := "installing"
			if ok, _ := vfs.FileExists(fs, target); ok {
				if !o.Force && (cv.GetVersion() == o.Current || !o.UpdateMode) {
					return fmt.Errorf("plugin %s already found in %s", desc.PluginName, dir)
				}
				if o.UpdateMode {
					verb = "updating"
				}
				fs.Remove(target)
			}
			o.Printer.Printf("%s plugin %s[%s] in %s...\n", verb, desc.PluginName, desc.PluginVersion, dir)
			dst, err := fs.OpenFile(target, vfs.O_CREATE|vfs.O_TRUNC|vfs.O_WRONLY, 0o755)
			if err != nil {
				return errors.Wrapf(err, "cannot create plugin file %s", target)
			}
			defer dst.Close()
			src, err := fs.OpenFile(file.Name(), vfs.O_RDONLY, 0)
			if err != nil {
				return errors.Wrapf(err, "cannot open plugin executable %s", file.Name())
			}
			_, err = io.Copy(dst, src)
			utils2.IgnoreError(src.Close())
			utils2.IgnoreError(os.Remove(file.Name()))
			utils2.IgnoreError(WritePluginSource(dir, cv, found.Meta().Name, desc.PluginName))
			if err != nil {
				return errors.Wrapf(err, "cannot copy plugin file %s", target)
			}
		}
	}
	return nil
}

func RemoveFile(file string) error {
	if ok, err := vfs.FileExists(osfs.New(), file); !ok || err != nil {
		return err
	}
	return os.Remove(file)
}

func RemovePluginSource(dir string, name string) error {
	return RemoveFile(filepath.Join(dir, "."+name+".info"))
}

func WritePluginSource(dir string, cv ocm.ComponentVersionAccess, rsc, name string) error {
	spec, err := cpi.ToGenericRepositorySpec(cv.Repository().GetSpecification())
	if err != nil {
		return err
	}
	cv.Repository().GetSpecification()
	src := &PluginSource{
		Repository: spec,
		Component:  cv.GetName(),
		Version:    cv.GetVersion(),
		Resource:   rsc,
	}

	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	//nolint: gosec // yes
	return os.WriteFile(filepath.Join(dir, "."+name+".info"), data, 0o644)
}

func ReadPluginSource(dir string, name string) (*PluginSource, error) {
	data, err := os.ReadFile(filepath.Join(dir, "."+name+".info"))
	if err != nil {
		return nil, fmt.Errorf("no source information available for plugin %s", name)
	}

	var src PluginSource
	if err := json.Unmarshal(data, &src); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal source information")
	}
	return &src, nil
}
