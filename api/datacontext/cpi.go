package datacontext

import (
	"context"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/logging"
	"github.com/modern-go/reflect2"
	"ocm.software/ocm/api/datacontext/action/handlers"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/refmgmt/finalized"
	"ocm.software/ocm/api/utils/runtimefinalizer"
)

// NewContextBase creates a context base implementation supporting
// context attributes and the binding to a context.Context.
func NewContextBase(eff Context, typ string, key interface{}, parentAttrs Attributes, delegates Delegates) InternalContext {
	updater, _ := eff.(Updater)
	return newContextBase(eff, typ, key, parentAttrs, &updater,
		ComposeDelegates(logging.NewWithBase(delegates.LoggingContext()), handlers.NewRegistry(nil, delegates.GetActions())),
	)
}

// GCWrapper is the embeddable base type
// for a context wrapper handling garbage collection.
// It also handles the BindTo interface for a context.
type GCWrapper struct {
	ref  *finalized.FinalizedRef
	self Context // reference to wrapper
	ctx  Context // reference to internal context
	key  interface{}
}

var _ provider = (*GCWrapper)(nil) // wrapper provides access to internal context ref

// setSelf is not public to enforce
// the usage of this GCWrapper type in context
// specific garbage collection wrappers.
// It is enforced by the
// finalizableContextWrapper interface.
func (w *GCWrapper) setSelf(a refmgmt.Allocatable, self Context, ictx Context, key interface{}) {
	if a != nil {
		w.ref, _ = finalized.NewPlainFinalizedView(a)
	}
	w.self = self
	w.ctx = ictx
	w.key = key
}

func (w *GCWrapper) IsPersistent() bool {
	return true
}

func (w *GCWrapper) GetInternalContext() Context {
	return w.ctx
}

func init() { // linter complains about unused method.
	(&GCWrapper{}).setSelf(nil, nil, nil, nil)
}

// BindTo makes the Context reachable via the resulting context.Context.
// Go requires not to use a pointer receiver, here ??????
func (b GCWrapper) BindTo(ctx context.Context) context.Context {
	return context.WithValue(ctx, b.key, b.self)
}

func (w GCWrapper) getAllocatable() refmgmt.Allocatable {
	return w.ref.GetAllocatable()
}

type view interface {
	getAllocatable() refmgmt.Allocatable
}

type viewI interface {
	GetAllocatable() refmgmt.Allocatable
}

func GetContextRefCount(ctx Context) int {
	switch a := ctx.(type) {
	case view:
		if m, ok := a.getAllocatable().(refmgmt.RefMgmt); ok {
			return m.RefCount()
		}
	case viewI:
		if m, ok := a.GetAllocatable().(refmgmt.RefMgmt); ok {
			return m.RefCount()
		}
	}
	return -1
}

type persistent interface {
	IsPersistent() bool
}

type provider interface {
	GetInternalContext() Context
}

type ViewCreator[C Context] interface {
	CreateView() C
}

func IsPersistentContextRef(ctx Context) bool {
	if p, ok := ctx.(persistent); ok {
		return p.IsPersistent()
	}
	return false
}

// PersistentContextRef ensures a persistent context ref to the given
// context to avoid an automatic cleanup of the context, which is
// executed if all persistent refs are gone.
// If you want to keep context related objects longer than your used
// context reference, you should keep a persistent ref. This
// could be the one provided by context creation, or by retrieving
// an explicit one using this function.
func PersistentContextRef[C Context](ctx C) C {
	if IsPersistentContextRef(ctx) {
		return ctx
	}
	var c interface{} = ctx
	for {
		if p, ok := c.(provider); ok {
			c = p.GetInternalContext()
		} else {
			break
		}
	}
	return c.(ViewCreator[C]).CreateView()
}

func InternalContextRef[C Context](ctx C) C {
	if IsPersistentContextRef(ctx) {
		var c interface{} = ctx
		for {
			if p, ok := c.(provider); ok {
				c = p.GetInternalContext()
			} else {
				break
			}
		}
		return c.(C)
	}
	return ctx
}

// finalizableContextWrapper is the interface for
// a context wrapper used to establish a garbage collectable
// runtime finalizer.
// It is a helper interface for Go generics to enforce a
// struct pointer.
type finalizableContextWrapper[C InternalContext, P any] interface {
	InternalContext

	SetContext(C)
	provider
	setSelf(refmgmt.Allocatable, Context, Context, interface{})
	*P
}

// FinalizedContext wraps a context implementation C into a separate wrapper
// object of type *W and returns this wrapper.
// It should have the type
//
//	struct {
//	   C
//	}
//
// The wrapper is created and a runtime finalizer is
// defined for this object, which calls the Finalize Method on the
// context implementation.
func FinalizedContext[W Context, C InternalContext, P finalizableContextWrapper[C, W]](c C) P {
	var v W
	p := (P)(&v)
	p.SetContext(c)
	p.setSelf(c.GetAllocatable(), p, c, c.GetKey()) // prepare for generic bind operation
	return p
}

type contextBase struct {
	ctxtype     string
	allocatable refmgmt.Allocatable
	id          ContextIdentity
	key         interface{}
	effective   Context
	attributes  *_attributes
	delegates

	finalizer *finalizer.Finalizer
	recorder  *runtimefinalizer.RuntimeFinalizationRecoder
}

var _ Context = (*contextBase)(nil)

func newContextBase(eff Context, typ string, key interface{}, parentAttrs Attributes, updater *Updater, delegates Delegates) *contextBase {
	recorder := &runtimefinalizer.RuntimeFinalizationRecoder{}
	id := ContextIdentity(fmt.Sprintf("%s/%d", typ, contextrange.NextId()))
	c := &contextBase{
		ctxtype:    typ,
		id:         id,
		key:        key,
		effective:  eff,
		finalizer:  &finalizer.Finalizer{},
		attributes: newAttributes(eff, parentAttrs, updater),
		delegates:  delegates,
		recorder:   recorder,
	}
	c.allocatable = refmgmt.NewAllocatable(c.cleanup, true)
	Debug(c, "create context", "id", c.GetId())
	return c
}

func (c *contextBase) IsIdenticalTo(ctx Context) bool {
	if reflect2.IsNil(ctx) {
		return false
	}
	return c.GetId() == ctx.GetId()
}

func (c *contextBase) GetAllocatable() refmgmt.Allocatable {
	return c.allocatable
}

func (c *contextBase) BindTo(ctx context.Context) context.Context {
	panic("should never be called")
}

func (c *contextBase) GetType() string {
	return c.ctxtype
}

func (c *contextBase) GetId() ContextIdentity {
	return c.id
}

func (c *contextBase) GetKey() interface{} {
	return c.key
}

func (c *contextBase) AttributesContext() AttributesContext {
	return c.effective.AttributesContext()
}

func (c *contextBase) GetAttributes() Attributes {
	return c.attributes
}

func (c *contextBase) GetRecorder() *runtimefinalizer.RuntimeFinalizationRecoder {
	return c.recorder
}

func (c *contextBase) cleanup() error {
	if c.recorder != nil {
		c.recorder.Record(c.id)
	}
	return c.Cleanup()
}

func (c *contextBase) Cleanup() error {
	list := errors.ErrListf("cleanup %s", c.id)
	list.Addf(nil, c.finalizer.Finalize(), "finalizers")
	list.Add(c.attributes.Finalize())
	return list.Result()
}

func (c *contextBase) Finalize() error {
	return c.finalizer.Finalize()
}

func (c *contextBase) Finalizer() *finalizer.Finalizer {
	return c.finalizer
}

// AssureUpdater is used to assure the existence of an updater in
// a root context if a config context is down the context hierarchy.
// This method SHOULD only be called by a config context.
func AssureUpdater(attrs AttributesContext, u Updater) {
	c, ok := attrs.(*gcWrapper)
	if !ok {
		return
	}
	if c.updater == nil {
		c.updater = u
	}
}
