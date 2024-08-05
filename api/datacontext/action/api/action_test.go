package api_test

import (
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/datacontext/action/api"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

const NAME = "testAction"

const CONSUMER_TYPE = "TestAction"

const ID_HOSTNAME = hostpath.ID_HOSTNAME

func RegisterAction(registry api.ActionTypeRegistry) {
	registry.RegisterAction(NAME, "test action", "nothing special", []string{ID_HOSTNAME})

	registry.RegisterActionType(api.NewActionType[*ActionSpec, *ActionResult](NAME, "v1"))
	registry.RegisterActionType(api.NewActionTypeByConverter[*ActionSpec, *ActionSpecV2, *ActionResult, *ActionResultV2](NAME, "v2", convertSpecV2{}, convertResultV2{}))
}

func NewActionSpec(field string) api.ActionSpec {
	return &ActionSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(runtime.TypeName(NAME, "v1")),
		Field:               field,
	}
}

func NewActionResult(msg string) api.ActionResult {
	return &ActionResult{
		ObjectVersionedType: runtime.NewVersionedObjectType(runtime.TypeName(NAME, "v1")),
		Message:             msg,
	}
}

////////////////////////////////////////////////////////////////////////////////
// internal version

type ActionSpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	Field                       string `json:"field"`
}

func (a *ActionSpec) Selector() api.Selector {
	return api.Selector(a.Field)
}

func (a *ActionSpec) GetConsumerAttributes() common.Properties {
	return common.Properties(credentials.NewConsumerIdentity(CONSUMER_TYPE,
		ID_HOSTNAME, a.Field,
	))
}

type ActionResult struct {
	runtime.ObjectVersionedType `json:",inline"`
	Message                     string `json:"message"`
}

func (r ActionResult) GetMessage() string {
	return r.Message
}

////////////////////////////////////////////////////////////////////////////////
// external version

type ActionSpecV2 struct {
	runtime.ObjectVersionedType `json:",inline"`
	Data                        string `json:"data"`
}

type ActionResultV2 struct {
	runtime.ObjectVersionedType `json:",inline"`
	Data                        string `json:"data"`
}

type convertSpecV2 struct{}

func (c convertSpecV2) ConvertFrom(in *ActionSpec) (*ActionSpecV2, error) {
	return &ActionSpecV2{
		ObjectVersionedType: runtime.NewVersionedObjectType(runtime.TypeName(NAME, "v2")),
		Data:                in.Field,
	}, nil
}

func (c convertSpecV2) ConvertTo(in *ActionSpecV2) (*ActionSpec, error) {
	return &ActionSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(runtime.TypeName(NAME, "v2")),
		Field:               in.Data,
	}, nil
}

type convertResultV2 struct{}

func (c convertResultV2) ConvertFrom(in *ActionResult) (*ActionResultV2, error) {
	return &ActionResultV2{
		ObjectVersionedType: runtime.NewVersionedObjectType(runtime.TypeName(NAME, "v2")),
		Data:                in.Message,
	}, nil
}

func (c convertResultV2) ConvertTo(in *ActionResultV2) (*ActionResult, error) {
	return &ActionResult{
		ObjectVersionedType: runtime.NewVersionedObjectType(runtime.TypeName(NAME, "v2")),
		Message:             in.Data,
	}, nil
}
