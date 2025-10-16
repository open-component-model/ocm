package install

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/mandelsoft/goutils/errors"
	. "github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/xeipuuv/gojsonschema"
	globalconfig "ocm.software/ocm/api/config/extensions/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/attrs/logforward"
	logcfg "ocm.software/ocm/api/datacontext/config/logging"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/ocm/tools/toi"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/spiff"
)

func ValidateByScheme(src []byte, schemedata []byte) error {
	data, err := yaml.YAMLToJSON(src)
	if err != nil {
		return errors.Wrapf(err, "converting data to json")
	}
	schemedata, err = yaml.YAMLToJSON(schemedata)
	if err != nil {
		return errors.Wrapf(err, "converting scheme to json")
	}
	documentLoader := gojsonschema.NewBytesLoader(data)

	scheme, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemedata))
	if err != nil {
		return errors.Wrapf(err, "invalid scheme")
	}
	res, err := scheme.Validate(documentLoader)
	if err != nil {
		return err
	}

	if !res.Valid() {
		errs := res.Errors()
		errMsg := errs[0].String()
		for i := 1; i < len(errs); i++ {
			errMsg = fmt.Sprintf("%s;%s", errMsg, errs[i].String())
		}
		return errors.New(errMsg)
	}

	return nil
}

type ExecutorContext struct {
	Spec  toi.ExecutorSpecification
	Image *toi.Image
	CV    ocm.ComponentVersionAccess
}

func GetResource(res ocm.ResourceAccess, target interface{}) error {
	m, err := res.AccessMethod()
	if err != nil {
		return errors.Wrapf(err, "failed to instantiate access")
	}
	data, err := m.Get()
	m.Close()
	if err != nil {
		return errors.Wrapf(err, "cannot get resource content")
	}
	return runtime.DefaultYAMLEncoding.Unmarshal(data, target)
}

func DetermineExecutor(executor *toi.Executor, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (*ExecutorContext, error) {
	espec := ExecutorContext{Image: executor.Image}

	if espec.Image == nil {
		if cv == nil {
			return nil, errors.Newf("resource access not possible without component version")
		}
		if executor.ResourceRef == nil {
			return nil, errors.Newf("executor resource reference required for toi package executor")
		}
		res, eff, err := resourcerefs.ResolveResourceReference(cv, *executor.ResourceRef, resolver)
		if err != nil {
			return nil, errors.ErrNotFoundWrap(err, "executor resource", executor.ResourceRef.String())
		}
		defer func() {
			if eff != nil {
				eff.Close()
			}
		}()
		switch res.Meta().Type {
		case resourcetypes.OCI_IMAGE:
		case toi.TypeTOIExecutor:
			err := GetResource(res, &espec.Spec)
			if err != nil {
				return nil, errors.ErrInvalidWrap(err, "toi executor")
			}
			espec.Image = espec.Spec.Image
			if espec.Image == nil {
				if cv == nil {
					return nil, errors.Newf("resource access not possible without component version")
				}
				if espec.Spec.ImageRef == nil {
					return nil, errors.Newf("executor image reference required for toi executor")
				}
				var eff2 ocm.ComponentVersionAccess
				res, eff2, err = resourcerefs.ResolveResourceReference(eff, *espec.Spec.ImageRef, resolver)
				if err != nil {
					return nil, errors.ErrNotFoundWrap(err, "executor resource", executor.ResourceRef.String())
				}
				defer eff2.Close()
			}
		default:
			return nil, errors.ErrInvalid("executor resource type", res.Meta().Type, executor.ResourceRef.String())
		}

		if res.Meta().Type != resourcetypes.OCI_IMAGE {
			return nil, errors.ErrInvalid("executor resource type", res.Meta().Type)
		}
		ref, err := ocmutils.GetOCIArtifactRef(octx, res)
		if err != nil {
			return nil, errors.Wrapf(err, "image ref for executor resource %s not found", executor.ResourceRef.String())
		}
		espec.Image = &toi.Image{
			Ref: ref,
		}
		espec.CV, eff = eff, nil
	}
	return &espec, nil
}

func mappingKeyFor(value string, m map[string]string) string {
	if m == nil {
		return value
	}
	for k, v := range m {
		if v == value {
			return k
		}
	}
	return ""
}

// CheckCredentialRequests determine required credentials for executor.
func CheckCredentialRequests(executor *toi.Executor, spec *toi.PackageSpecification, espec *toi.ExecutorSpecification) (map[string]CredentialsRequestSpec, map[string]string, error) {
	credsspec := spec.Credentials
	credmapping := map[string]string{}

	if len(espec.Credentials) > 0 {
		if len(spec.CredentialsRequest.Credentials) > 0 {
			// first, determine mapping and subset of defined spec required for executor
			for k := range spec.Credentials {
				ke := k
				if executor.CredentialMapping != nil {
					if m, ok := executor.CredentialMapping[k]; ok {
						ke = m
					}
					if _, ok := espec.Credentials[ke]; ok {
						credmapping[k] = ke
					} else {
						delete(credsspec, k)
					}
				}
			}

			// second, check for spec errors and complete package spec
			for ke, e := range espec.Credentials {
				ko := mappingKeyFor(ke, executor.CredentialMapping)
				if ko == "" {
					return nil, nil, errors.Newf("credential mapping missing for executor credential key %q", ke)
				}
				if o, ok := credsspec[ko]; !ok {
					if !e.Optional {
						// implicit inheritance of executor spec setting
						credsspec[ko] = e
						credmapping[ko] = ke
					}
				} else {
					if err := o.Match(&e); err != nil {
						return nil, nil, errors.Wrapf(err, "credential %q does not match executor setting %q", ko, ke)
					}
				}
			}
		} else {
			// no credential requests specified for package, use the one from the executor
			credsspec = espec.Credentials
		}
	} else {
		if len(executor.CredentialMapping) > 0 {
			// determine subset of credentials required for executor
			credmapping = executor.CredentialMapping
			for k := range credsspec {
				if _, ok := credmapping[k]; !ok {
					delete(credsspec, k)
				} else {
					credmapping[k] = k
				}
			}
		} else {
			// assume to require all as defined
			for k := range credsspec {
				credmapping[k] = k
			}
		}
	}
	return credsspec, credmapping, nil
}

func ProcessConfig(name string, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver, template []byte, config []byte, libraries []metav1.ResourceReference, schemedata []byte) ([]byte, error) {
	var err error

	if len(config) == 0 {
		if len(schemedata) > 0 {
			err = ValidateByScheme([]byte("{}"), schemedata)
			if err != nil {
				return nil, errors.Wrapf(err, name+" validation failed")
			}
		}
		if len(template) == 0 {
			return nil, nil
		}
	}

	stubs := spiff.Options{}
	for i, lib := range libraries {
		res, eff, err := resourcerefs.ResolveResourceReference(cv, lib, resolver)
		if err != nil {
			return nil, errors.ErrNotFound("library resource %s not found", lib.String())
		}
		defer eff.Close()
		m, err := res.AccessMethod()
		if err != nil {
			return nil, errors.ErrNotFound("cannot access library resource", lib.String())
		}
		data, err := m.Get()
		m.Close()
		if err != nil {
			return nil, errors.ErrNotFound("cannot access library resource", lib.String())
		}
		stubs.Add(spiff.StubData(fmt.Sprintf("spiff lib%d", i), data))
	}

	if len(schemedata) > 0 || len(template) == 0 {
		// process input without template first to have final version without bfore using template
		// to be verified by json scheme
		if config != nil {
			config, err = spiff.CascadeWith(spiff.Context(octx), spiff.TemplateData(name, config), stubs)
			if err != nil {
				return nil, errors.Wrapf(err, "error processing "+name)
			}
		}
	}
	if len(schemedata) > 0 {
		toi.Log.Info("validating by scheme", name)
		err = ValidateByScheme(config, schemedata)
		if err != nil {
			return nil, errors.Wrapf(err, name+" validation failed")
		}
	}
	if len(template) > 0 {
		config, err = spiff.CascadeWith(spiff.Context(octx), spiff.TemplateData(name+" template", template), spiff.StubData(name, config), stubs)
		if err != nil {
			return nil, errors.Wrapf(err, "error processing "+name+" template")
		}
	}
	return config, err
}

// ExecuteAction prepared the execution options and executes the action.
func ExecuteAction(p common.Printer, d Driver, name string, spec *toi.PackageSpecification, creds *Credentials, params []byte, octxp ocm.ContextProvider, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (*OperationResult, error) {
	var err error

	var finalize Finalizer
	defer finalize.Finalize()

	octx := octxp.OCMContext()

	var executor *toi.Executor
	for idx, e := range spec.Executors {
		if e.Actions == nil {
			executor = &spec.Executors[idx]
			break
		}
		for _, a := range e.Actions {
			if a == name {
				executor = &spec.Executors[idx]
				break
			}
		}
	}
	if executor == nil {
		return nil, errors.Newf("no executor found for action %s", name)
	}

	// validate executor config
	espec, err := DetermineExecutor(executor, octx, cv, resolver)
	if err != nil {
		return nil, err
	}

	if espec.Spec.Actions != nil {
		found := false
		for _, a := range espec.Spec.Actions {
			if a == name {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.ErrNotSupported("action", name, "toi executor "+executor.ResourceRef.String())
		}
	}

	if espec.Spec.Outputs != nil {
		list := errors.ErrListf("invalid outputs")
		for o := range executor.Outputs {
			if _, ok := espec.Spec.Outputs[o]; !ok {
				list.Add(fmt.Errorf("output %s not available from executor", o))
			}
		}
		if list.Len() > 0 {
			return nil, list.Result()
		}
	}
	// prepare executor config
	econfig, err := ProcessConfig("executor config", octx, espec.CV, resolver, espec.Spec.Template, executor.Config, espec.Spec.Libraries, espec.Spec.Scheme)
	if err != nil {
		return nil, errors.Wrapf(err, "error executor config")
	}

	if econfig == nil {
		p.Printf("no executor config found\n")
	} else {
		p.Printf("using executor config:\n%s\n", utils.IndentLines(string(econfig), "  "))
	}
	// handle credentials
	credreqs, credmapping, err := CheckCredentialRequests(executor, spec, &espec.Spec)
	if err != nil {
		return nil, err
	}

	// prepare ocm config with credential settings and logging config forwarding
	if len(credreqs) > 0 {
		if creds == nil {
			return nil, errors.Newf("credential settings required")
		}
	}

	if p, ok := cv.Repository().(credentials.ConsumerIdentityProvider); ok {
		cid := p.GetConsumerId()
		if cid.IsSet() {
			if creds == nil {
				creds = &Credentials{}
			}
			creds.Forwarded = append(creds.Forwarded, toi.ForwardSpec{
				ConsumerId:   cid,
				ConsumerType: p.GetIdentityMatcher(),
			})
		}
	}
	ccfg, credvals, err := GetCredentials(octx.CredentialsContext(), creds, credreqs, credmapping)
	if err != nil {
		return nil, errors.Wrapf(err, "credential evaluation failed")
	}

	if lc := logforward.Get(octx); lc != nil {
		if err := ccfg.AddConfig(logcfg.NewWithConfig("default", lc)); err != nil {
			return nil, errors.Wrapf(err, "cannot create logging config forwarding")
		}
	}
	{
		data, _ := yaml.Marshal(ccfg)
		p.Printf("using ocm config:\n%s\n", utils.IndentLines(string(data), "  "))
	}

	// prepare user config
	params, err = ProcessConfig("parameter data", octx, cv, resolver, spec.Template, params, spec.Libraries, spec.Scheme)
	if err != nil {
		return nil, errors.Wrapf(err, "error processing parameters")
	}
	if params == nil {
		p.Printf("no parameter config found\n")
	} else {
		p.Printf("using package parameters:\n:%s\n", utils.IndentLines(string(params), "  "))
	}

	if executor.ParameterMapping != nil {
		orig := params
		params, err = spiff.CascadeWith(
			spiff.Functions(NewFunctions(octx, credvals)),
			spiff.TemplateData("executor parameter mapping", executor.ParameterMapping),
			spiff.StubData("package config", params))

		if err == nil {
			if !bytes.Equal(orig, params) {
				p.Printf("using executor parameters:\n%s\n", utils.IndentLines(string(params), "  "))
			}
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error mapping parameters to executor")
	}

	names := []string{}
	for n := range credreqs {
		m := credmapping[n]
		names = append(names, n+"->"+m)
	}
	sort.Strings(names)

	src := ""
	if executor.ResourceRef != nil {
		src = fmt.Sprintf("[%s]", executor.ResourceRef.String())
	}
	p.Printf("using executor image %s[%s] with credentials %v\n", espec.Image.Ref, src, names)

	// setup executor operation
	op := &Operation{
		Action:      name,
		Image:       *espec.Image,
		Environment: nil,
		Files:       nil,
		Outputs:     nil,
		Out:         nil,
		Err:         nil,
	}

	defer op.Close()

	// prepare file content to be passed to executor
	err = setupFiles(octx, &finalize, op, ccfg, params, econfig, cv, resolver)
	if err != nil {
		return nil, errors.Wrapf(err, "error setting up executor file set")
	}

	op.Outputs = executor.Outputs

	op.ComponentVersion = common.VersionedElementKey(cv).String()
	return d.Exec(op)
}

func (op *Operation) Close() error {
	list := errors.ErrListf("closing operation")
	for _, b := range op.Files {
		list.Add(b.Close())
	}
	return list.Result()
}

func setupFiles(octx ocm.Context, finalize *Finalizer, op *Operation, ccfg *globalconfig.Config, params []byte, econfig []byte, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) error {
	// prepare file content to be passed to executor
	op.Files = map[string]blobaccess.BlobAccess{}
	if ccfg != nil {
		data, err := runtime.DefaultYAMLEncoding.Marshal(ccfg)
		if err != nil {
			return errors.Wrapf(err, "marshalling ocm config failed")
		}
		op.Files[InputOCMConfig] = blobaccess.ForData(mime.MIME_OCTET, data)
	}
	if params != nil {
		op.Files[InputParameters] = blobaccess.ForData(mime.MIME_OCTET, params)
	}
	if econfig != nil {
		op.Files[InputConfig] = blobaccess.ForData(mime.MIME_OCTET, econfig)
	}
	if cv != nil {
		fs, err := osfs.NewTempFileSystem()
		if err != nil {
			return errors.Wrapf(err, "cannot create temp file system")
		}
		finalize.With(func() error { return vfs.Cleanup(fs) })
		repo, err := ctf.Create(octx, accessobj.ACC_CREATE, "arch", 0o600, accessio.FormatTGZ, accessio.PathFileSystem(fs))
		if err != nil {
			return errors.Wrapf(err, "cannot create repo for component version")
		}
		defer repo.Close()
		handler, err := standard.New(standard.Recursive(), standard.KeepGlobalAccess(), standard.Resolver(resolver))
		if err != nil {
			return errors.Wrapf(err, "cannot create transfer handler")
		}
		err = transfer.TransferVersion(nil, nil, cv, repo, handler)
		if err != nil {
			return errors.Wrapf(err, "component version transport failed")
		}
		op.Files[InputOCMRepo] = blobaccess.ForFile(mime.MIME_OCTET, "arch", fs)
	}
	return nil
}
