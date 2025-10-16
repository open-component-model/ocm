package ppi

import (
	"context"
	"encoding/json"
	"io"

	"github.com/spf13/cobra"

	"ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/action"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/internal"
	"ocm.software/ocm/api/utils/runtime"
)

type (
	Descriptor             = descriptor.Descriptor
	UploaderKey            = descriptor.UploaderKey
	UploaderDescriptor     = descriptor.UploaderDescriptor
	DownloaderKey          = descriptor.DownloaderKey
	DownloaderDescriptor   = descriptor.DownloaderDescriptor
	AccessMethodDescriptor = descriptor.AccessMethodDescriptor
	CLIOption              = descriptor.CLIOption

	ActionSpecInfo       = internal.ActionSpecInfo
	AccessSpecInfo       = internal.AccessSpecInfo
	ValueSetInfo         = internal.ValueSetInfo
	UploadTargetSpecInfo = internal.UploadTargetSpecInfo
)

var REALM = descriptor.REALM

type Plugin interface {
	Name() string
	Version() string
	Descriptor() descriptor.Descriptor

	SetDescriptorTweaker(func(descriptor descriptor.Descriptor) descriptor.Descriptor)

	SetShort(s string)
	SetLong(s string)
	SetConfigParser(config func(raw json.RawMessage) (interface{}, error))
	ForwardLogging(b ...bool)

	RegisterDownloader(arttype, mediatype string, u Downloader) error
	GetDownloader(name string) Downloader
	GetDownloaderFor(arttype, mediatype string) Downloader

	RegisterUploader(arttype, mediatype string, u Uploader) error
	GetUploader(name string) Uploader
	GetUploaderFor(arttype, mediatype string) Uploader
	DecodeUploadTargetSpecification(data []byte) (UploadTargetSpec, error)

	RegisterAccessMethod(m AccessMethod) error
	DecodeAccessSpecification(data []byte) (AccessSpec, error)
	GetAccessMethod(name string, version string) AccessMethod

	RegisterAction(a Action) error
	DecodeAction(data []byte) (ActionSpec, error)
	GetAction(name string) Action

	RegisterValueMergeHandler(h ValueMergeHandler) error
	GetValueMergeHandler(name string) ValueMergeHandler

	RegisterValueSet(h ValueSet) error
	DecodeValueSet(purpose string, data []byte) (runtime.TypedObject, error)
	GetValueSet(purpose, name, version string) ValueSet

	RegisterCommand(c Command) error
	GetCommand(name string) Command
	Commands() []Command

	RegisterConfigType(c cpi.ConfigType) error
	GetConfigType(name string) *descriptor.ConfigTypeDescriptor
	ConfigTypes() []descriptor.ConfigTypeDescriptor

	GetOptions() *Options
	GetConfig() (interface{}, error)
}

type AccessMethod interface {
	runtime.TypedObjectDecoder[AccessSpec]

	Name() string
	Version() string

	// Options provides the list of CLI options supported to compose the access
	// specification.
	Options() []options.OptionType

	// Description provides a general description for the access method kind.
	Description() string
	// Format describes the attributes of the dedicated version.
	Format() string

	ValidateSpecification(p Plugin, spec AccessSpec) (info *AccessSpecInfo, err error)
	Reader(p Plugin, spec AccessSpec, creds credentials.Credentials) (io.ReadCloser, error)
	ComposeAccessSpecification(p Plugin, opts Config, config Config) error
}

type AccessSpec = runtime.TypedObject

type AccessSpecProvider func() AccessSpec

type UploadFormats runtime.KnownTypes[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]]

type Uploader interface {
	Decoders() UploadFormats

	Name() string
	Description() string

	ValidateSpecification(p Plugin, spec UploadTargetSpec) (info *UploadTargetSpecInfo, err error)
	Upload(ctx context.Context, p Plugin, arttype, mediatype, hint, digest string, spec UploadTargetSpec, creds credentials.Credentials, reader io.Reader) (AccessSpecProvider, error)
}

type UploadTargetSpec = runtime.TypedObject

type DownloadResultProvider func() (string, error)

type Downloader interface {
	Name() string
	Description() string
	ConfigSchema() []byte

	Writer(p Plugin, arttype, mediatype string, filepath string, config []byte) (io.WriteCloser, DownloadResultProvider, error)
}

type ActionSpec = action.ActionSpec

type ActionResult = action.ActionResult

type Action interface {
	Name() string
	Description() string
	DefaultSelectors() []string
	ConsumerType() string

	Execute(p Plugin, spec ActionSpec, creds credentials.DirectCredentials) (result ActionResult, err error)
}

type Value = runtime.RawValue

type ValueMergeResult struct {
	Modified bool   `json:"modified"`
	Value    Value  `json:"value"`
	Message  string `json:"message,omitempty"`
}

type ValueMergeData struct {
	Local   Value `json:"local"`
	Inbound Value `json:"inbound"`
}

type ValueMergeHandler interface {
	Name() string
	Description() string

	Execute(p Plugin, local Value, inbound Value, config json.RawMessage) (result ValueMergeResult, err error)
}

type ValueSet interface {
	runtime.TypedObjectDecoder[AccessSpec]

	Name() string
	Version() string

	// Purposes describes the purposes the set should be used for.
	// So far, only the purpose PURPOSE_ROUTINGSLIP is defined.
	Purposes() []string

	// Options provides the list of CLI options supported to compose the access
	// specification.
	Options() []options.OptionType

	// Description provides a general description for the access method kind.
	Description() string
	// Format describes the attributes of the dedicated version.
	Format() string

	ValidateSpecification(p Plugin, spec runtime.TypedObject) (info *ValueSetInfo, err error)
	ComposeSpecification(p Plugin, opts Config, config Config) error
}

// Command is the interface for a CLI command provided by a plugin.
type Command interface {
	// Name of command used in the plugin.
	// This is also the default object type and is used to
	// name top-level commands in the CLI.
	Name() string
	Description() string
	Usage() string
	Short() string
	Example() string
	// ObjectType is optional and can be used
	// together with a verb. It then is used as
	// sub command name for the object type.
	// By default, the command name is used.
	ObjectType() string
	// Verb is optional and can be set
	// to place the command in the verb hierarchy of
	// the OCM CLI. It is used together with the ObjectType.
	// (command will be *ocm <verb> <object type>*.
	Verb() string
	// Realm is optional and is used to place the command
	// in a realm. This requires a verb.
	Realm() string
	CLIConfigRequired() bool

	Command() *cobra.Command
}
