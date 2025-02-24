package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"sigs.k8s.io/yaml"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugindirattr"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/extraid"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/filelock"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/semverutils"
)

type PluginInfo struct {
	Size       int64                  `json:"size,omitempty"`
	ModTime    time.Time              `json:"modtime,omitempty"`
	Descriptor *descriptor.Descriptor `json:"descriptor,omitempty"`
}

type PluginSourceInfo struct {
	Repository *cpi.GenericRepositorySpec `json:"repository,omitempty"`
	Component  string                     `json:"component,omitempty"`
	Version    string                     `json:"version,omitempty"`
	Resource   string                     `json:"resource,omitempty"`
}

func (p *PluginSourceInfo) HasSourceInfo() bool {
	return p != nil && p.Repository != nil && p.Component != "" && p.Version != "" && p.Resource != ""
}

func (p *PluginSourceInfo) GetDescription() string {
	if p != nil && p.HasSourceInfo() {
		return p.Component + ":" + p.Version
	}
	return "local"
}

type PluginInstallationInfo struct {
	PluginSourceInfo `json:",inline"`
	PluginInfo       *PluginInfo `json:"info,omitempty"`
}

func (p *PluginInstallationInfo) HasSourceInfo() bool {
	return p != nil && p.PluginSourceInfo.HasSourceInfo()
}

func (p *PluginInstallationInfo) IsValidPluginInfo(execpath string) bool {
	if !p.HasPluginInfo() {
		return false
	}
	fi, err := os.Stat(execpath)
	if err != nil {
		return false
	}
	return fi.Size() == p.PluginInfo.Size && fi.ModTime() == p.PluginInfo.ModTime
}

func (p *PluginInstallationInfo) UpdatePluginInfo(execpath string) (bool, error) {
	desc, err := GetPluginInfo(execpath)
	if err != nil {
		return false, err
	}
	fi, err := os.Stat(execpath)
	if err != nil {
		return false, err
	}
	n := &PluginInfo{
		Size:       fi.Size(),
		ModTime:    fi.ModTime(),
		Descriptor: desc,
	}
	mod := !reflect.DeepEqual(n, p.PluginInfo)
	p.PluginInfo = n
	return mod, nil
}

func (p *PluginInstallationInfo) HasPluginInfo() bool {
	return p.PluginInfo != nil
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
	src, err := readPluginInstalltionInfo(dir, name)
	if err != nil {
		return err
	}
	if !src.HasSourceInfo() {
		return fmt.Errorf("no source information available for plugin %s", name)
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
		o.Printer.Printf("plugin %s already up-to-date\n", name)
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
			o.Printer.Printf("plugin %s already up-to-date\n", name)
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
		o.Printer.Printf("plugin %s already up-to-date\n", name)
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
		if cv.GetVersion() == o.Current {
			o.Printer.Printf("version %s already installed\n", o.Current)
			if !o.Force {
				return nil
			}
		}

		dir := plugindirattr.Get(o.Context)
		if dir == "" {
			home, err := os.UserHomeDir() // use home if provided
			if err != nil {
				return fmt.Errorf("failed to determine home directory to determine default plugin directory: %w", err)
			}
			dir = filepath.Join(home, plugindirattr.DEFAULT_PLUGIN_DIR)
			if err := os.Mkdir(dir, os.ModePerm|os.ModeDir); err != nil {
				return fmt.Errorf("failed to create default plugin directory: %w", err)
			}
			if err := plugindirattr.Set(o.Context, dir); err != nil {
				return fmt.Errorf("failed to set plugin dir after defaulting: %w", err)
			}
		}

		lock, err := filelock.LockDir(dir)
		if err != nil {
			return err
		}
		defer lock.Close()

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
		src, err := fs.OpenFile(file.Name(), vfs.O_RDONLY, 0)
		if err != nil {
			dst.Close()
			return errors.Wrapf(err, "cannot open plugin executable %s", file.Name())
		}
		_, err = io.Copy(dst, src)
		dst.Close()
		utils.IgnoreError(src.Close())
		utils.IgnoreError(os.Remove(file.Name()))
		utils.IgnoreError(SetPluginSourceInfo(dir, cv, found.Meta().Name, desc.PluginName))
		if err != nil {
			return errors.Wrapf(err, "cannot copy plugin file %s", target)
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

func SetPluginSourceInfo(dir string, cv ocm.ComponentVersionAccess, rsc, name string) error {
	src, err := readPluginInstalltionInfo(dir, name)
	if err != nil {
		return err
	}
	spec, err := cpi.ToGenericRepositorySpec(cv.Repository().GetSpecification())
	if err != nil {
		return err
	}
	src.PluginSourceInfo = PluginSourceInfo{
		Repository: spec,
		Component:  cv.GetName(),
		Version:    cv.GetVersion(),
		Resource:   rsc,
	}

	_, err = src.UpdatePluginInfo(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	return writePluginInstallationInfo(src, dir, name)
}

func writePluginInstallationInfo(src *PluginInstallationInfo, dir string, name string) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	//nolint:gosec // yes
	return os.WriteFile(filepath.Join(dir, "."+name+".info"), data, 0o644)
}

func readPluginInstalltionInfo(dir string, name string) (*PluginInstallationInfo, error) {
	data, err := os.ReadFile(filepath.Join(dir, "."+name+".info"))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, errors.Wrapf(err, "cannot read plugin info for %s", name)
		}
		return &PluginInstallationInfo{}, nil
	}

	var src PluginInstallationInfo
	if err := json.Unmarshal(data, &src); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal source information")
	}
	return &src, nil
}
