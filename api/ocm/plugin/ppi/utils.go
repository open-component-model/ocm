package ppi

import (
	"encoding/json"
	"reflect"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"golang.org/x/exp/slices"

	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
)

type decoder runtime.TypedObjectDecoder[runtime.TypedObject]

type blobProviderBase struct {
	decoder
	nameDescription

	format string
}

func MustNewBlobProviderBase(name string, proto AccessSpec, desc string, format string) blobProviderBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}

	return blobProviderBase{
		decoder: decoder,
		nameDescription: nameDescription{
			name: name,
			desc: desc,
		},
		format: format,
	}
}

func (b *AccessMethodBase) BlobProviderBase() string {
	return b.format
}

////////////////////////////////////////////////////////////////////////////////

type InputTypeBase struct {
	blobProviderBase
}

func MustNewInputTypeBase(name string, proto InputSpec, desc string, format string) InputTypeBase {
	return InputTypeBase{MustNewBlobProviderBase(name, proto, desc, format)}
}

func (b *InputTypeBase) Format() string {
	return b.format
}

////////////////////////////////////////////////////////////////////////////////

type AccessMethodBase struct {
	blobProviderBase
	version string
}

func MustNewAccessMethodBase(name, version string, proto AccessSpec, desc string, format string) AccessMethodBase {
	return AccessMethodBase{
		blobProviderBase: MustNewBlobProviderBase(name, proto, desc, format),
		version:          version,
	}
}

func (b *AccessMethodBase) Version() string {
	return b.version
}

func (b *AccessMethodBase) Format() string {
	return b.format
}

////////////////////////////////////////////////////////////////////////////////

type UploaderBase = nameDescription

func MustNewUploaderBase(name, desc string) UploaderBase {
	return UploaderBase{
		name: name,
		desc: desc,
	}
}

////////////////////////////////////////////////////////////////////////////////

type ValueSetBase struct {
	decoder
	nameDescription

	version string
	format  string

	purposes []string
}

func MustNewValueSetBase(name, version string, proto runtime.TypedObject, purposes []string, desc string, format string) ValueSetBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}
	return ValueSetBase{
		decoder: decoder,
		nameDescription: nameDescription{
			name: name,
			desc: desc,
		},
		version:  version,
		format:   format,
		purposes: slices.Clone(purposes),
	}
}

func (b *ValueSetBase) Version() string {
	return b.version
}

func (b *ValueSetBase) Format() string {
	return b.format
}

func (b *ValueSetBase) Purposes() []string {
	return b.purposes
}

////////////////////////////////////////////////////////////////////////////////

type nameDescription struct {
	name string
	desc string
}

func (b *nameDescription) Name() string {
	return b.name
}

func (b *nameDescription) Description() string {
	return b.desc
}

////////////////////////////////////////////////////////////////////////////////

type transferHandler struct {
	name        string
	description string
	questions   []DecisionHandler
}

func NewTransferHandler(name, desc string) *transferHandler {
	return &transferHandler{
		name:        name,
		description: desc,
		questions:   nil,
	}
}

func (t *transferHandler) GetName() string {
	return t.name
}

func (t *transferHandler) GetDescription() string {
	return t.description
}

func (t *transferHandler) GetQuestions() []DecisionHandler {
	return t.questions
}

func (t *transferHandler) RegisterDecision(h DecisionHandler) error {
	if TransferHandlerQuestions[h.GetQuestion()] == nil {
		return errors.ErrInvalid(descriptor.KIND_QUESTION, h.GetQuestion())
	}
	for _, e := range t.questions {
		if e.GetQuestion() == h.GetQuestion() {
			return errors.ErrAlreadyExists(descriptor.KIND_QUESTION, e.GetQuestion())
		}
	}
	t.questions = append(t.questions, h)
	return nil
}

// DecisionHandlerBase provides access to the
// non-functional attributes of a DecisionHandler.
// It can be created with NewDecisionHandlerBase and
// embedded into the final DecisionHandler implementation.
type DecisionHandlerBase struct {
	question    string
	description string
	labels      *[]string
}

func (d *DecisionHandlerBase) GetQuestion() string {
	return d.question
}

func (d *DecisionHandlerBase) GetDescription() string {
	return d.description
}

func (d *DecisionHandlerBase) GetLabels() *[]string {
	return d.labels
}

func NewDecisionHandlerBase(q, desc string, labels ...string) DecisionHandlerBase {
	return DecisionHandlerBase{q, desc, generics.Pointer(slices.Clone(labels))}
}

////////////////////////////////////////////////////////////////////////////////

type QuestionResultFunc func(p Plugin, question QuestionArguments) (bool, error)

func ComponentVersionQuestionFunc(f func(p Plugin, question *ComponentVersionQuestionArguments) (bool, error)) QuestionResultFunc {
	return func(p Plugin, question QuestionArguments) (bool, error) {
		return f(p, question.(*ComponentVersionQuestionArguments))
	}
}

type ComponentReferenceQuestionFunc = func(p Plugin, question *ComponentReferenceQuestionArguments) (bool, error)

func ForComponentReferenceQuestion(f func(p Plugin, question *ComponentReferenceQuestionArguments) (bool, error)) QuestionResultFunc {
	return func(p Plugin, question QuestionArguments) (bool, error) {
		return f(p, question.(*ComponentReferenceQuestionArguments))
	}
}

type ArtifactQuestionFunc = func(p Plugin, question *ArtifactQuestionArguments) (bool, error)

func ForArtifactQuestion(f ArtifactQuestionFunc) QuestionResultFunc {
	return func(p Plugin, question QuestionArguments) (bool, error) {
		return f(p, question.(*ArtifactQuestionArguments))
	}
}

type defaultDecisionHandler struct {
	DecisionHandlerBase
	handler func(p Plugin, question QuestionArguments) (bool, error)
}

// NewDecisionHandler provides a default decision handler based on its standard
// fields and a handler function.
func NewDecisionHandler(q, desc string, h QuestionResultFunc, labels ...string) DecisionHandler {
	return &defaultDecisionHandler{
		DecisionHandlerBase: NewDecisionHandlerBase(q, desc, labels...),
		handler:             h,
	}
}

func (d *defaultDecisionHandler) DecideOn(p Plugin, question QuestionArguments) (bool, error) {
	return d.handler(p, question)
}

////////////////////////////////////////////////////////////////////////////////
// specialized handler creation

func NewTransferResourceDecision(desc string, h ArtifactQuestionFunc, labels ...string) DecisionHandler {
	return NewDecisionHandler(Q_TRANSFER_RESOURCE, desc, ForArtifactQuestion(h))
}

func NewTransferSourceDecision(desc string, h ArtifactQuestionFunc, labels ...string) DecisionHandler {
	return NewDecisionHandler(Q_TRANSFER_SOURCE, desc, ForArtifactQuestion(h))
}

func NewEnforceTransportDesision(desc string, h ComponentReferenceQuestionFunc, labels ...string) DecisionHandler {
	return NewDecisionHandler(Q_ENFORCE_TRANSPORT, desc, ForComponentReferenceQuestion(h))
}

func NewTransferVersionDecision(desc string, h ComponentReferenceQuestionFunc, labels ...string) DecisionHandler {
	return NewDecisionHandler(Q_TRANSFER_VERSION, desc, ForComponentReferenceQuestion(h))
}

func NewOverwriteVersionDecision(desc string, h ComponentReferenceQuestionFunc, labels ...string) DecisionHandler {
	return NewDecisionHandler(Q_OVERWRITE_VERSION, desc, ForComponentReferenceQuestion(h))
}

func NewUpdateVersionDecision(desc string, h ComponentReferenceQuestionFunc, labels ...string) DecisionHandler {
	return NewDecisionHandler(Q_UPDATE_VERSION, desc, ForComponentReferenceQuestion(h))
}

////////////////////////////////////////////////////////////////////////////////

type signingHandler struct {
	name        string
	description string
	consumer    ConsumerProvider
	signer      signing.Signer
	verifier    signing.Verifier
}

func NewSigningHandler(name, desc string, signer signing.Signer) *signingHandler {
	return &signingHandler{
		name:        name,
		description: desc,
		signer:      signer,
	}
}

func (s *signingHandler) WithVerifier(verifier signing.Verifier) *signingHandler {
	s.verifier = verifier
	return s
}

func (s *signingHandler) WithCredentials(p ConsumerProvider) *signingHandler {
	s.consumer = p
	return s
}

func (t *signingHandler) GetName() string {
	return t.name
}

func (t *signingHandler) GetDescription() string {
	return t.description
}

func (t *signingHandler) GetSigner() signing.Signer {
	return t.signer
}

func (t *signingHandler) GetVerifier() signing.Verifier {
	return t.verifier
}

func (t *signingHandler) GetConsumerProvider() ConsumerProvider {
	return t.consumer
}

////////////////////////////////////////////////////////////////////////////////

// Config is a generic structured config stored in a string map.
type Config map[string]interface{}

func (c Config) GetValue(name string) (interface{}, bool) {
	v, ok := c[name]
	return v, ok
}

func (c Config) ConvertFor(list ...flagsets.ConfigOptionType) error {
	for _, o := range list {
		if v, ok := c[o.GetName()]; ok {
			t := reflect.TypeOf(o.Create().Value())
			if t != reflect.TypeOf(v) {
				data, err := json.Marshal(v)
				if err != nil {
					return errors.Wrapf(err, "cannot marshal option value for %q", o.GetName())
				}
				value := reflect.New(t)
				err = json.Unmarshal(data, value.Interface())
				if err != nil {
					return errors.Wrapf(err, "cannot unmarshal option value for %q[%s]", o.GetName(), o.ValueType())
				}
				c[o.GetName()] = value.Elem().Interface()
			}
		}
	}
	return nil
}
