package sliceutils

import (
	"cmp"
	"slices"

	"github.com/mandelsoft/goutils/general"
)

// Slice is a slice of an arbitrary type.
// It offers convenience methods.
type Slice[E any] []E

func (s Slice[E]) Get(i int) E {
	return s[i]
}

func (s Slice[E]) Set(i int, e E) {
	s[i] = e
}

func (s *Slice[E]) Clip() {
	*s = slices.Clip(*s)
}

func (s *Slice[E]) Grow(n int) {
	*s = slices.Grow(*s, n)
}

// Add appends element to the slice.
func (s *Slice[E]) Add(elems ...E) {
	*s = append(*s, elems...)
}

func (s Slice[E]) CopyAppend(elems ...E) Slice[E] {
	return CopyAppend(s, elems...)
}

func (s *Slice[E]) AppendUniqueFunc(cmp general.EqualsFunc[E], elems ...E) {
	*s = AppendUniqueFunc(*s, cmp, elems...)
}

func (s Slice[E]) CopyAppendUniqueFunc(cmp general.EqualsFunc[E], elems ...E) Slice[E] {
	return CopyAppendUniqueFunc(s, cmp, elems...)
}

func (s *Slice[E]) InsertIndex(i int, elems ...E) {
	*s = slices.Insert(*s, i, elems...)
}

func (s *Slice[E]) DeleteIndex(i int) {
	*s = append((*s)[:i], (*s)[i+1:]...)
}

// Delete removes the elements s[i:j] from s.
// Delete panics if j > len(s) or s[i:j] is not a valid slice of s.
// Delete is O(len(s)-i), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
// Delete zeroes the elements s[len(s)-(j-i):len(s)].
func (s *Slice[E]) DeleteRange(i, j int) {
	*s = slices.Delete(*s, i, j)
}

func (s *Slice[E]) Replace(i, j int, v ...E) {
	*s = slices.Replace(*s, i, j, v...)
}

func (s *Slice[E]) CompactFunc(eq general.EqualsFunc[E]) {
	*s = slices.CompactFunc(*s, eq)
}

// ContainsFunc reports whether at least one
// element e of s satisfies f(e).
func (s Slice[E]) ContainsFunc(f general.ContainsFunc[E]) bool {
	return slices.ContainsFunc(s, f)
}

// IndexFunc returns the first index i satisfying f(s[i]),
// or -1 if none do.
func (s Slice[E]) IndexFunc(f general.ContainsFunc[E]) int {
	return slices.IndexFunc(s, f)
}

func (s Slice[E]) SortFunc(cmp general.CompareFunc[E]) {
	slices.SortFunc(s, cmp)
}

func (s Slice[E]) SortStableFunc(cmp general.CompareFunc[E]) {
	slices.SortStableFunc(s, cmp)
}

func (s Slice[E]) MinFunc(cmp func(a, b E) int) {
	slices.MinFunc(s, cmp)
}

func (s Slice[E]) MaxFunc(cmp func(a, b E) int) {
	slices.MaxFunc(s, cmp)
}

////////////////////////////////////////////////////////////////////////////////

type ComparableSlice[E comparable] Slice[E]

func (s *ComparableSlice[E]) AppendUnique(elems ...E) {
	*s = AppendUnique(*s, elems...)
}

// Contains reports whether s contains at least one
// element e
func (s ComparableSlice[E]) Contains(e E) bool {
	return slices.Contains(s, e)
}

func (s ComparableSlice[E]) ContainsAll(elems ...E) bool {
	for _, e := range elems {
		if !s.Contains(e) {
			return false
		}
	}
	return true
}

////
// base functions from Slice[E]

func (s ComparableSlice[E]) Get(i int) E {
	return s[i]
}

func (s ComparableSlice[E]) Set(i int, e E) {
	s[i] = e
}

func (s *ComparableSlice[E]) Clip() {
	*s = slices.Clip(*s)
}

func (s *ComparableSlice[E]) Grow(n int) {
	*s = slices.Grow(*s, n)
}

// Add appends element to the slice.
func (s *ComparableSlice[E]) Add(elems ...E) {
	*s = append(*s, elems...)
}

func (s ComparableSlice[E]) CopyAppend(elems ...E) Slice[E] {
	return CopyAppend(s, elems...)
}

func (s *ComparableSlice[E]) AppendUniqueFunc(cmp general.EqualsFunc[E], elems ...E) {
	*s = AppendUniqueFunc(*s, cmp, elems...)
}

func (s ComparableSlice[E]) CopyAppendUniqueFunc(cmp general.EqualsFunc[E], elems ...E) ComparableSlice[E] {
	return CopyAppendUniqueFunc(s, cmp, elems...)
}

func (s *ComparableSlice[E]) InsertIndex(i int, elems ...E) {
	*s = slices.Insert(*s, i, elems...)
}

func (s *ComparableSlice[E]) DeleteIndex(i int) {
	*s = append((*s)[:i], (*s)[i+1:]...)
}

// Delete removes the elements s[i:j] from s.
// Delete panics if j > len(s) or s[i:j] is not a valid slice of s.
// Delete is O(len(s)-i), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
// Delete zeroes the elements s[len(s)-(j-i):len(s)].
func (s *ComparableSlice[E]) DeleteRange(i, j int) {
	*s = slices.Delete(*s, i, j)
}

func (s *ComparableSlice[E]) Replace(i, j int, v ...E) {
	*s = slices.Replace(*s, i, j, v...)
}

func (s *ComparableSlice[E]) CompactFunc(eq general.EqualsFunc[E]) {
	*s = slices.CompactFunc(*s, eq)
}

// ContainsFunc reports whether at least one
// element e of s satisfies f(e).
func (s ComparableSlice[E]) ContainsFunc(f general.ContainsFunc[E]) bool {
	return slices.ContainsFunc(s, f)
}

// IndexFunc returns the first index i satisfying f(s[i]),
// or -1 if none do.
func (s ComparableSlice[E]) IndexFunc(f general.ContainsFunc[E]) int {
	return slices.IndexFunc(s, f)
}

func (s ComparableSlice[E]) SortFunc(cmp general.CompareFunc[E]) {
	slices.SortFunc(s, cmp)
}

func (s ComparableSlice[E]) SortStableFunc(cmp general.CompareFunc[E]) {
	slices.SortStableFunc(s, cmp)
}

func (s ComparableSlice[E]) MinFunc(cmp func(a, b E) int) {
	slices.MinFunc(s, cmp)
}

func (s ComparableSlice[E]) MaxFunc(cmp func(a, b E) int) {
	slices.MaxFunc(s, cmp)
}

////////////////////////////////////////////////////////////////////////////////

// OrderedSlice if a slice with comparable elements.
// It offers additional convenience methods relying
// on the comparable feature.
type OrderedSlice[E cmp.Ordered] Slice[E]

func (s *OrderedSlice[E]) AppendUnique(elems ...E) {
	*s = AppendUnique(*s, elems...)
}

// Contains reports whether s contains at least one
// element e
func (s OrderedSlice[E]) Contains(e E) bool {
	return slices.Contains(s, e)
}

func (s OrderedSlice[E]) ContainsAll(elems ...E) bool {
	for _, e := range elems {
		if !s.Contains(e) {
			return false
		}
	}
	return true
}

////
// base functions from Slice[E]

func (s OrderedSlice[E]) Get(i int) E {
	return s[i]
}

func (s OrderedSlice[E]) Set(i int, e E) {
	s[i] = e
}

func (s *OrderedSlice[E]) Clip() {
	*s = slices.Clip(*s)
}

func (s *OrderedSlice[E]) Grow(n int) {
	*s = slices.Grow(*s, n)
}

func (s *OrderedSlice[E]) Add(elems ...E) {
	*s = append(*s, elems...)
}

func (s OrderedSlice[E]) CopyAppend(elems ...E) OrderedSlice[E] {
	return CopyAppend(s, elems...)
}

func (s OrderedSlice[E]) CopyAppendUnique(elems ...E) OrderedSlice[E] {
	return CopyAppendUniqueFunc(s, general.EqualsComparable[E], elems...)
}

func (s OrderedSlice[E]) CopyAppendUniqueFunc(cmp general.EqualsFunc[E], elems ...E) OrderedSlice[E] {
	return CopyAppendUniqueFunc(s, cmp, elems...)
}

func (s *OrderedSlice[E]) DeleteIndex(i int) {
	*s = append((*s)[:i], (*s)[i+1:]...)
}

func (s *OrderedSlice[E]) DeleteRange(i, j int) {
	*s = slices.Delete(*s, i, j)
}

func (s *OrderedSlice[E]) Replace(i, j int, v ...E) {
	*s = slices.Replace(*s, i, j, v...)
}

func (s *OrderedSlice[E]) Compact() {
	*s = slices.Compact(*s)
}

func (s *OrderedSlice[E]) CompactFunc(eq general.EqualsFunc[E]) {
	*s = slices.CompactFunc(*s, eq)
}

func (s OrderedSlice[E]) Sort() {
	slices.Sort(s)
}

func (s OrderedSlice[E]) SortFunc(cmp general.CompareFunc[E]) {
	slices.SortFunc(s, cmp)
}

func (s OrderedSlice[E]) SortStableFunc(cmp general.CompareFunc[E]) {
	slices.SortStableFunc(s, cmp)
}

func (s OrderedSlice[E]) Min() E {
	return slices.Min(s)
}

func (s OrderedSlice[E]) MinFunc(cmp general.CompareFunc[E]) E {
	return slices.MinFunc(s, cmp)
}

func (s OrderedSlice[E]) Max() E {
	return slices.Max(s)
}

func (s OrderedSlice[E]) MaxFunc(cmp general.CompareFunc[E]) E {
	return slices.MaxFunc(s, cmp)
}

func (s OrderedSlice[E]) ContainsFunc(f general.ContainsFunc[E]) bool {
	return slices.ContainsFunc(s, f)
}

func (s OrderedSlice[E]) IndexFunc(f general.ContainsFunc[E]) int {
	return slices.IndexFunc(s, f)
}

func (s OrderedSlice[E]) Index(e E) int {
	return slices.Index(s, e)
}
