package resource_test

import (
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/refmgmt/resource"
)

// Resource is the intended resource interface.
// It must incorporate the resource.ResourceView interface
// providing the view related part of the interface.
type Resource interface {
	resource.ResourceView[Resource]

	Name() string
	Operation(err error) error
}

type (
	// ResourceImpl implements io.Closer to finally release allocated resources
	// and the additional non-view-related part of the resource interface.
	ResourceImpl struct {
		resource.ResourceImplBase[Resource]
		name string
	}
	_resourceImpl = *ResourceImpl
)

var _ resource.ResourceImplementation[Resource] = (*ResourceImpl)(nil)

func (r *ResourceImpl) Name() string {
	return r.name
}

func (r *ResourceImpl) Operation(err error) error {
	return err
}

// Close is called for the last closed view and
// may handle the release of allocated sub resources.
func (r *ResourceImpl) Close() error {
	if r.name == "" {
		return fmt.Errorf("oops")
	}
	r.name = ""
	return nil
}

type _resourceView = resource.ResourceViewInt[Resource]

// resourceView implementation the mapping of a ResourceImpl
// to a fully-fledged Resource implementation including
// the view-related part.
type resourceView struct {
	_resourceView
	_resourceImpl
}

var _ Resource = (*resourceView)(nil)

// Close must be implemented to resolve the two provided Close
// methods to the one of the view-related part.
func (r *resourceView) Close() error {
	return r._resourceView.Close()
}

func (r *resourceView) Operation(err error) error {
	return r.Execute(func() error { return r._resourceImpl.Operation(err) })
}

func resourceViewCreator(impl *ResourceImpl, v resource.CloserView, d resource.ViewManager[Resource]) Resource {
	return &resourceView{
		_resourceView: resource.NewView(v, d),
		_resourceImpl: impl,
	}
}

// New create a new Resource by creating a ResourceImpl
// and adding the reference management by calling resource.NewResource,
// which internally will call the resourceViewCreator function to create
// the first view.
func New(name string) Resource {
	i := &ResourceImpl{name: name}
	return resource.NewResource(i, resourceViewCreator, name)
}

var _ = Describe("ref test", func() {
	It("handles main ref", func() {
		r := New("alice")

		Expect(r.IsClosed()).To(BeFalse())

		MustBeSuccessful(r.Close())
		Expect(r.IsClosed()).To(BeTrue())
		Expect(r.Close()).To(Equal(resource.ErrClosed))
		Expect(r.Name()).To(Equal(""))
	})

	It("handle last closed view", func() {
		r := New("alice")
		Expect(r.IsClosed()).To(BeFalse())
		v := Must(r.Dup())
		Expect(v.IsClosed()).To(BeFalse())

		MustBeSuccessful(r.Close())
		Expect(r.IsClosed()).To(BeTrue())
		Expect(v.IsClosed()).To(BeFalse())

		Expect(r.Close()).To(Equal(resource.ErrClosed))
		Expect(r.Name()).To(Equal("alice"))

		MustBeSuccessful(v.Close())
		Expect(v.IsClosed()).To(BeTrue())
		Expect(v.Close()).To(Equal(resource.ErrClosed))
		Expect(v.Name()).To(Equal(""))
	})

	It("executes operation", func() {
		r := New("alice")
		Expect(r.IsClosed()).To(BeFalse())

		Expect(r.Operation(nil)).To(Succeed())
		Expect(r.Operation(fmt.Errorf("fail"))).To(MatchError("fail"))
		MustBeSuccessful(r.Close())
		Expect(r.Operation(nil)).To(Equal(resource.ErrClosed))
	})
})
