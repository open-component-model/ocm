package finalized

import (
	"runtime"

	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

type FinalizedRef struct {
	allocatable refmgmt.Allocatable
	id          finalizer.ObjectIdentity
	recorder    *finalizer.RuntimeFinalizationRecoder
}

func NewPlainFinalizedView(allocatable refmgmt.Allocatable) (*FinalizedRef, error) {
	return NewFinalizedView(allocatable, "", nil)
}

func NewFinalizedView(allocatable refmgmt.Allocatable, id finalizer.ObjectIdentity, rec *finalizer.RuntimeFinalizationRecoder) (*FinalizedRef, error) {
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

func (v *FinalizedRef) GetRefId() finalizer.ObjectIdentity {
	return v.id
}

func cleanup(v *FinalizedRef) {
	v.allocatable.Unref()
	if v.recorder != nil {
		v.recorder.Record(v.id)
	}
}
