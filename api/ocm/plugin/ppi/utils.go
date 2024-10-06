package ppi

import (
	"encoding/json"
	"reflect"

	errors2 "github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"

	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/tech/signing"
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
		return errors2.ErrInvalid(descriptor.KIND_QUESTION, h.GetQuestion())
	}
	for _, e := range t.questions {
		if e.GetQuestion() == h.GetQuestion() {
			return errors2.ErrAlreadyExists(descriptor.KIND_QUESTION, e.GetQuestion())
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

func ComponentVersionQuestionFunv(f func(p Plugin, question *ComponentVersionQuestion) (bool, error)) func(p Plugin, question Question) (bool, error) {
	return func(p Plugin, question Question) (bool, error) {
		return f(p, question.(*ComponentVersionQuestion))
	}
}

func ComponentReferenceQuestionFunc(f func(p Plugin, question *ComponentReferenceQuestion) (bool, error)) func(p Plugin, question Question) (bool, error) {
	return func(p Plugin, question Question) (bool, error) {
		return f(p, question.(*ComponentReferenceQuestion))
	}
}

func ArtifactQuestionFunc(f func(p Plugin, question *ArtifactQuestion) (bool, error)) func(p Plugin, question Question) (bool, error) {
	return func(p Plugin, question Question) (bool, error) {
		return f(p, question.(*ArtifactQuestion))
	}
}

type defaultDecisionHandler struct {
	DecisionHandlerBase
	handler func(p Plugin, question Question) (bool, error)
}

// NewDecisionHandler provides a default decision handler based on its standard
// fields and a handler function.
func NewDecisionHandler(q, desc string, h func(p Plugin, question Question) (bool, error), labels ...string) DecisionHandler {
	return &defaultDecisionHandler{
		DecisionHandlerBase: NewDecisionHandlerBase(q, desc, labels...),
		handler:             h,
	}
}

func (d *defaultDecisionHandler) DecideOn(p Plugin, question Question) (bool, error) {
	return d.handler(p, question)
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

func (c Config) ConvertFor(list ...options.OptionType) error {
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
