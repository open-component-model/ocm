package accessio

import (
	"math/rand"
	"time"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/refmgmt"
)

var (
	ErrClosed   = refmgmt.ErrClosed
	ErrReadOnly = errors.ErrReadOnly()
)

////////////////////////////////////////////////////////////////////////////////

type NopCloser = iotools.NopCloser

////////////////////////////////////////////////////////////////////////////////

type retriableError struct {
	wrapped error
}

func IsRetriableError(err error) bool {
	if err == nil {
		return false
	}
	return errors.IsA(err, &retriableError{})
}

func RetriableError(err error) error {
	if err == nil {
		return nil
	}
	return &retriableError{err}
}

func RetriableError1[T any](r T, err error) (T, error) {
	if err == nil {
		return r, nil
	}
	return r, &retriableError{err}
}

func RetriableError2[S, T any](s S, r T, err error) (S, T, error) {
	if err == nil {
		return s, r, nil
	}
	return s, r, &retriableError{err}
}

func (e *retriableError) Error() string {
	return e.wrapped.Error()
}

func (e *retriableError) Unwrap() error {
	return e.wrapped
}

func Retry(cnt int, d time.Duration, f func() error) error {
	for {
		err := f()
		if err == nil || cnt <= 0 || !IsRetriableError(err) {
			return err
		}
		jitter := time.Duration(rand.Int63n(int64(d))) //nolint:gosec // just an random number
		d = 2*d + (d/2-jitter)/10
		cnt--
	}
}

func Retry1[T any](cnt int, d time.Duration, f func() (T, error)) (T, error) {
	for {
		r, err := f()
		if err == nil || cnt <= 0 || !IsRetriableError(err) {
			return r, err
		}
		jitter := time.Duration(rand.Int63n(int64(d))) //nolint:gosec // just an random number
		d = 2*d + (d/2-jitter)/10
		cnt--
	}
}

func Retry2[S, T any](cnt int, d time.Duration, f func() (S, T, error)) (S, T, error) {
	for {
		s, t, err := f()
		if err == nil || cnt <= 0 || !IsRetriableError(err) {
			return s, t, err
		}
		jitter := time.Duration(rand.Int63n(int64(d))) //nolint:gosec // just an random number
		d = 2*d + (d/2-jitter)/10
		cnt--
	}
}
