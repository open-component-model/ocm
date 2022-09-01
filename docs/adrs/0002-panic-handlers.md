# 2. Externally provided panic handlers

Status: accepted
Date: 2022.09.01.
Authors: [@Skarlso]
Deciders: [@Skarlso @mandelsoft]

## Context

Right now, the OCM library part has explicit `panic` calls in it in certain cases. These lead to certain cases that we
would like to avoid, such as outlined in the [Consequences](#consequences) section.

## Decision

Use panic handlers defined as follows:

```go
// PanicHandler defines how a handler should look like. It returns true if the handler
// wants the panic to happen.
type PanicHandler func(interface{})bool

// PanicHandlers is a list of functions which will be invoked when a panic happens.
var PanicHandlers = []PanicHandler{}

// HandleCrash simply catches a crash and runs any user defined panic handlers that exist.
// The handlers can then decide to actually panic or not.
//
// E.g., you can provide one or more additional handlers for something like shutting down go routines gracefully.
func HandleCrash(additionalHandlers ...PanicHandler) {
	if r := recover(); r != nil {
        if len(PanicHandlers) == 0 && len(additionalHandlers) == 0 {
            // If there are no handlers, throw the panic.
            panic(r)
        }
		for _, fn := range PanicHandlers {
			if fn(r) {
                panic(r)
            }
		}
		for _, fn := range additionalHandlers {
			if fn(r) {
                panic(r)
            }
		}
	}
}
```

In code, on every path at the top of the path or at the panic's location, there should be a single defer call to this
function such as:

```go
func (this *DLL) Append(d *DLL) {
    // Add defer handling crashes
    defer panics.HandleCrash()
	if d.next != nil || d.prev != nil {
		panic("dll element already in use")
	}
	if this.lock != nil {
		this.lock.Lock()
		defer this.lock.Unlock()
		d.lock = this.lock
	}
	d.next = this.next
	d.prev = this
	if this.next != nil {
		this.next.prev = d
	}
	this.next = d
}
```

From a libraries perspective:

```go
func logPanic(r interface{}) {
	if v, ok := r.(string); ok {
		klog.Errorf("Encountered a panic: %s\n", v, stacktrace)
	} else {
		klog.Errorf("Encountered a panic: %#v (%v)\n", r, r)
	}
}

func main() {
    // TODO: Figure this out how a library would use OCM
    // For now, this is an illustration.
    ocm.RegisterPanicHandlers(logPanic)
    // do other ocm things. a crash now will be logged instead of aborting.
}
```

### Discussion

How do we add handlers? Two ways:

- add them during initialization of the type handlers
- add them as a new context such as `PanicHandlerContext` which any call then use

## Consequences

### Avoids interrupts

Consider the following scenario. An outside user, for example, the OCM controller, would require this library and use
it to parse OCM components. During the parsing, it also applies and verifies these components.

Let's assume that a panic happens inside OCM. Since the requiring library is unaware that a panic can happen inside the
library, it doesn't have recovery code in place. The controller crashes and can't further process any other OCM
artefacts that exist in the cluster.

### Avoids try-panic-crash-try-panic-crash loops

The above crash will lead to a try-panic-crash-try-panic-crash loop which might create some intermediary objects in the
process and leave the cluster in an unclean space. To avoid this, the panic handlers can make sure that the controller
has the opportunity to finish and clean up resources that it doesn't require.

### Makes it explicit

Panics that happen in library are usually explicit inside functions such as `MustParse` or `MustCompile`. Adding these
handlers makes it public and known that the OCM library can, and will panic in certain situations and that it's up to
the library to add decide how to proceed.