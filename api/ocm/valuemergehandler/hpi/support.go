package hpi

import (
	"ocm.software/ocm/api/ocm/valuemergehandler/internal"
)

type EmptyConfig struct{}

var _ Config = (*EmptyConfig)(nil)

func (c *EmptyConfig) Complete(ctx Context) error {
	return nil
}

type Merger[C, T any] func(ctx Context, cfg C, local T, target *T) (bool, error)

func New[C any, L any, P internal.ConfigPointer[C]](algo string, desc string, merger internal.Merger[P, L]) Handler {
	return internal.New[C, L, P](algo, desc, merger)
}
