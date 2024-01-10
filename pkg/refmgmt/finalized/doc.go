// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Package finalized provided a view management for a backend object,
// which is based on Go Garbage Collection and runtime finalizers.
// Finalization is not possible in Go, if an object is involved in
// a reference cycle. In such a case the complete cycle not not
// garbage collected at all.
//
// If some kind of finalization is required together with cyclic
// object dependencies, the cleanup of the object can therefore not
// be done with runtime finalizers.
// We separate a garbage collectable view object, which holds a
// reference to a backend object featuring cycles.
// The view object uses reference counting for its backend
// together with runtime finalization. Therefore, it does not require
// a Close method. If the view is garbage collected it releases its
// reference to the backend object. If the last view vanished
// the cleanup method for the backend object is called.
//
// The object functionality is exposed via an interface, only, which
// is also implemented by the vies by embedding a pointer to the backend
// object.
//
// If the backend object requires a cycle by holding local objects
// requiring a reference to the object, this can be done
// by NOT using view objects for this cycle, but the backend object
// itself. If the local object wants to pass the backend object to some
// outgoing call, it MUST wrap the backend object again into a new view.
// Therefore, objects involved in the cycle MUST be prepared to handle
// such outgoing references.
package finalized
