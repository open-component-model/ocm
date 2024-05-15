package finalized

import (
	"runtime"

	"github.com/open-component-model/ocm/pkg/refmgmt"
	"github.com/open-component-model/ocm/pkg/runtimefinalizer"
)

type FinalizedRef struct {
	allocatable refmgmt.Allocatable
	id          runtimefinalizer.ObjectIdentity
	recorder    *runtimefinalizer.RuntimeFinalizationRecoder
}

func NewPlainFinalizedView(allocatable refmgmt.Allocatable) (*FinalizedRef, error) {
	return NewFinalizedView(allocatable, "", nil)
}

func NewFinalizedView(allocatable refmgmt.Allocatable, id runtimefinalizer.ObjectIdentity, rec *runtimefinalizer.RuntimeFinalizationRecoder) (*FinalizedRef, error) {
	err := allocatable.Ref()
	if err != nil {
		return nil, err
	}
	v := &FinalizedRef{allocatable, id, rec}

	runtime.SetFinalizer(v, cleanup)
	return v, nil
}

func (v *FinalizedRef) GetAllocatable() refmgmt.Allocatable {
	return v.allocatable
}

func (v *FinalizedRef) GetRefId() runtimefinalizer.ObjectIdentity {
	return v.id
}

func cleanup(v *FinalizedRef) {
	v.allocatable.Unref()
	if v.recorder != nil {
		v.recorder.Record(v.id)
	}
}
