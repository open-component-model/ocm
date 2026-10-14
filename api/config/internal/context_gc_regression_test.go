package internal

import (
	"runtime"
	"testing"
	"time"
	"weak"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtimefinalizer"
)

// Inspect collection of the backend, not just whether its Close ran. A
// finalized ref in a backend cycle can retain it even after refcount reaches 0.
func orphanConfigBackend(t *testing.T, closeContext bool) weak.Pointer[coreContext] {
	t.Helper()
	ctx := Builder{}.New()
	backend := ctx.(*gcWrapper).coreContext
	ref := weak.Make(backend)
	if closeContext {
		if err := datacontext.Close(ctx); err != nil {
			t.Fatal(err)
		}
	}
	return ref
}

func TestConfigViewKeepsRootAttributesAlive(t *testing.T) {
	for _, secondView := range []bool{false, true} {
		name := "original_view"
		if secondView {
			name = "second_view"
		}
		t.Run(name, func(t *testing.T) {
			ctx := Builder{}.New()
			root := ctx.AttributesContext()
			recorder := runtimefinalizer.GetRuntimeFinalizationRecorder(root)
			root = nil
			if secondView {
				// Drop the original view without closing the remaining view's
				// backend. The second view must own the root's lifetime too.
				ctx = datacontext.PersistentContextRef(ctx.ConfigContext())
			}
			for i := 0; i < 20; i++ {
				runtime.GC()
				time.Sleep(10 * time.Millisecond)
			}
			if ids := recorder.Get(); len(ids) != 0 {
				t.Fatalf("root attributes finalized while config view is alive: %v", ids)
			}
			runtime.KeepAlive(ctx)
		})
	}
}

func TestConfigBackendIsCollected(t *testing.T) {
	for _, closeContext := range []bool{false, true} {
		name := "automatic"
		if closeContext {
			name = "explicit_close"
		}
		t.Run(name, func(t *testing.T) {
			ref := orphanConfigBackend(t, closeContext)
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				runtime.GC()
				if ref.Value() == nil {
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
			t.Fatal("configuration backend remains reachable after repeated GC")
		})
	}
}
