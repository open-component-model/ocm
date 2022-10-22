//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package clisupport

import (
	"fmt"
	"sync"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

type ConfigOptionType interface {
	Name() string
	Description() string

	Create() Option

	Equal(optionType ConfigOptionType) bool
}

type ConfigOptionTypeSet interface {
	Name() string

	OptionTypes() []ConfigOptionType
	SharedOptionTypes() []ConfigOptionType

	HasOptionType(name string) bool
	HasSharedOptionType(name string) bool

	GetSharedOptionType(name string) ConfigOptionType
	GetOptionType(name string) ConfigOptionType
	GetTypeSet(name string) ConfigOptionTypeSet

	AddOptionType(ConfigOptionType) error
	AddTypeSet(ConfigOptionTypeSet) error

	Align(parent ConfigOptionTypeSet) error

	CreateOptions() ConfigOptions
}

type configOptionTypeSet struct {
	lock    sync.RWMutex
	name    string
	options map[string]ConfigOptionType
	sets    map[string]ConfigOptionTypeSet
	shared  map[string]struct{}

	parent ConfigOptionTypeSet
}

func NewConfigOptionSet(name string, types ...ConfigOptionType) ConfigOptionTypeSet {
	set := &configOptionTypeSet{
		name:    name,
		options: map[string]ConfigOptionType{},
		sets:    map[string]ConfigOptionTypeSet{},
		shared:  map[string]struct{}{},
	}
	for _, t := range types {
		set.AddOptionType(t)
	}
	return set
}

func (s *configOptionTypeSet) Align(parent ConfigOptionTypeSet) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.parent != nil {
		return errors.ErrClosed("config option set")
	}
	for _, o := range parent.OptionTypes() {
		old := s.options[o.Name()]
		if old != nil {
			if !old.Equal(o) {
				return fmt.Errorf("alignment option type %s doesn not match (%T<->%T)", o.Name(), o, old)
			}
		}
	}
	for _, o := range parent.OptionTypes() {
		old := s.options[o.Name()]
		if old != nil {
			s.options[o.Name()] = nil
		}
	}
	s.parent = parent
	return nil
}

func (s *configOptionTypeSet) Name() string {
	return s.name
}

func (s *configOptionTypeSet) AddOptionType(optionType ConfigOptionType) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.parent != nil {
		return errors.ErrClosed("config option set")
	}
	name := optionType.Name()
	s.options[name] = optionType
	return nil
}

func (s *configOptionTypeSet) OptionTypes() []ConfigOptionType {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var list []ConfigOptionType
	for n, o := range s.options {
		if o == nil {
			o = s.parent.GetOptionType(n)
		}
		list = append(list, o)
	}
	return list
}

func (s *configOptionTypeSet) SharedOptionTypes() []ConfigOptionType {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var list []ConfigOptionType
	for n, o := range s.options {
		if o == nil {
			if _, ok := s.shared[n]; ok {
				o = s.parent.GetOptionType(n)
			}
		}
		list = append(list, o)
	}
	return list
}

func (s *configOptionTypeSet) HasOptionType(name string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.options[name]
	return ok
}

func (s *configOptionTypeSet) HasSharedOptionType(name string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.shared[name]
	return ok
}

func (s *configOptionTypeSet) GetOptionType(name string) ConfigOptionType {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.getOptionType(name)
}

func (s *configOptionTypeSet) GetSharedOptionType(name string) ConfigOptionType {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if _, ok := s.shared[name]; ok {
		return s.getOptionType(name)
	}
	return nil
}

func (s *configOptionTypeSet) getOptionType(name string) ConfigOptionType {
	if t, ok := s.options[name]; ok {
		if t == nil {
			return s.parent.GetOptionType(name)
		}
		return t
	}
	return nil
}

func (s *configOptionTypeSet) AddTypeSet(set ConfigOptionTypeSet) error {
	if set == nil {
		return nil
	}
	var finalize utils.Finalizer
	defer finalize.Lock(&s.lock).Finalize()

	name := set.Name()
	if nested, ok := s.sets[name]; ok {
		if nested == set {
			return nil
		}
		return fmt.Errorf("%s: config type set with name %q already added", s.Name(), name)
	}
	list := set.OptionTypes()

	for _, o := range list {
		old := s.options[o.Name()]
		if old == nil {
			s.options[o.Name()] = o
		}
		s.shared[o.Name()] = struct{}{}
	}
	finalize.Finalize()
	if err := set.Align(s); err != nil {
		return err
	}
	s.sets[name] = set
	return nil
}

func (s *configOptionTypeSet) GetTypeSet(name string) ConfigOptionTypeSet {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.sets[name]
}

func (s *configOptionTypeSet) CreateOptions() ConfigOptions {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var opts []Option

	for n := range s.options {
		opts = append(opts, s.getOptionType(n).Create())
	}
	return NewOptions(opts)
}
