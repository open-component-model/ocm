package addhdlrs

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"
	clictx "ocm.software/ocm/api/cli"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

// ResourceInput describe the source for the content of
// a content based element (sources or resources).
// It is either an input or access specification.
type ResourceInput struct {
	// SourceFile described the original source (file) the input
	// is taken from. By default, this is not set since it is taken from the
	// file information of the processed constructor resource.
	// If an input aggregated in a constructor resource is provided
	// by some other source, this field can be set.
	// The source information is finally used by the input implementations
	// to evaluate relative path specifications in the input specification.
	// This should always relate to the original source.
	SourceFile string                 `json:"sourceFile,omitempty"`
	Access     *cpi.GenericAccessSpec `json:"access"`
	// Input  *inputs.BlobInput                `json:"input,omitempty"`
	Input *inputs.GenericInputSpec `json:"input,omitempty"`
}

func (r *ResourceInput) SetSourceFile(s string) {
	r.SourceFile = s
}

// ElementSpecHandler is the interface for a handler
// responsible to handle a dedicated kind of element specification.
type ElementSpecHandler interface {
	Key() string
	RequireInputs() bool
	Decode(data []byte) (ElementSpec, error)
}

type ElementSource interface {
	// Origin provides access to the source
	// specification used to provide elements.
	Origin() SourceInfo
	// Get provides access to the content of the element source.
	Get() (string, error)
}

type SourceInfo interface {
	Origin() string
	Id() string

	String() string
	Sub(indices ...interface{}) SourceInfo
}

type sourceInfo struct {
	origin  string
	indices []interface{}
}

func NewSourceInfo(origin string) SourceInfo {
	return &sourceInfo{origin: origin}
}

func (s *sourceInfo) Sub(indices ...interface{}) SourceInfo {
	if len(indices) == 0 {
		return s
	}
	return &sourceInfo{
		origin:  s.origin,
		indices: sliceutils.CopyAppend(s.indices, indices...),
	}
}

func (s *sourceInfo) String() string {
	return s.Id()
}

func (s *sourceInfo) Origin() string {
	return s.origin
}

func (s *sourceInfo) Id() string {
	id := s.origin
	for _, i := range s.indices {
		id += fmt.Sprintf("[%v]", i)
	}
	return id
}

// ElementSpec is the specification of
// the model element. It provides access to
// common attributes, like the identity.
type ElementSpec interface {
	GetName() string
	GetVersion() string
	SetVersion(string)
	GetRawIdentity() metav1.Identity
	Info() string
	Validate(ctx clictx.Context, input *ResourceInput) error
}

// Element is the abstraction over model elements handled by
// the add handler, for example, resources, sources, references or complete
// component versions.
type Element interface {
	// Source provides info about the source the element has been
	// derived from. (for example a component-constructor.yaml or resources.yaml).
	Source() SourceInfo
	// Spec provides access to the element specification.
	Spec() ElementSpec
	// Type is used for types elements, like sources and resources.
	Type() string
	// Data provides access to the element descriptor representation.
	Data() []byte
	// Input provides access to the underlying data specification.
	// It is either an access specification or an input specification.
	Input() *ResourceInput
}

func NewElement(spec ElementSpec, input *ResourceInput, src SourceInfo, data []byte, indices ...interface{}) Element {
	return &element{
		source: src.Sub(indices...),
		spec:   spec,
		data:   data,
		input:  input,
	}
}

type element struct {
	source SourceInfo
	spec   ElementSpec
	data   []byte
	input  *ResourceInput
}

func (r *element) Source() SourceInfo {
	return r.source
}

func (r *element) Spec() ElementSpec {
	return r.spec
}

func (r *element) Data() []byte {
	return r.data
}

func (r *element) Input() *ResourceInput {
	return r.input
}

func (r *element) Type() string {
	if c, ok := r.spec.(interface{ GetType() string }); ok {
		return c.GetType()
	}
	return ""
}

func MapSpecsToElems[T ElementSpec](ctx clictx.Context, ictx inputs.Context, si SourceInfo, specs []T, h ElementSpecHandler) ([]Element, error) {
	var result []Element
	for i, e := range specs {
		data, err := json.Marshal(e)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal element %d", i+1)
		}
		elem, err := DetermineElementForData(ctx, ictx, si.Sub(i), data, h)
		if err != nil {
			return nil, errors.Wrapf(err, "entry %d", i+1)
		}
		result = append(result, elem)
	}
	return result, nil
}
