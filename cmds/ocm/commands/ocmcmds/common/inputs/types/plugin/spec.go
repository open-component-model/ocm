package plugin

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm"
	cpi "ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

type Spec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
	handler                                  *PluginHandler
}

var _ inputs.InputSpec = &Spec{}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	_, err := s.handler.GetAccess(s, ctx.OCMContext())
	if err != nil {
		return field.ErrorList{field.Invalid(fldPath, nil, err.Error())}
	}
	return nil
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	acc, err := s.handler.GetAccess(s, ctx.OCMContext())
	if err != nil {
		return nil, "", err
	}
	dir, err := inputs.GetBaseDir(ctx.FileSystem(), info.InputFilePath)
	if err != nil {
		dir = "."
	}
	return acc.GetBlob(dir)
}

func (s *Spec) GetInputVersion(ctx inputs.Context) string {
	return ""
}

func (s *Spec) Describe(ctx cpi.Context) string {
	return s.handler.Describe(s, ctx)
}

func (s *Spec) Handler() *PluginHandler {
	return s.handler
}

////////////////////////////////////////////////////////////////////////////////

type access struct {
	ctx ocm.Context

	handler *PluginHandler
	spec    *Spec
	info    *ppi.InputSpecInfo
	creds   json.RawMessage
}

func newAccess(p *PluginHandler, spec *Spec, ctx ocm.Context, info *ppi.InputSpecInfo) *access {
	return &access{
		ctx:     ctx,
		handler: p,
		spec:    spec,
		info:    info,
	}
}

func (m *access) MimeType() string {
	return m.info.MediaType
}

func (m *access) GetBlob(dir string) (blobaccess.BlobAccess, string, error) {
	spec, err := json.Marshal(m.spec)
	if err != nil {
		return nil, "", errors.Wrapf(err, "cannot marshal input spec")
	}
	return accessobj.CachedBlobAccessForWriter(m.ctx, m.MimeType(), plugin.NewInputDataWriter(m.handler.plug, dir, m.creds, spec)), m.info.Hint, nil
}

func (m *access) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	if len(m.info.ConsumerId) == 0 {
		return nil
	}
	return m.info.ConsumerId
}

func (m *access) GetIdentityMatcher() string {
	return hostpath.IDENTITY_TYPE
}
