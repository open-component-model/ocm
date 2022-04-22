// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"github.com/open-component-model/ocm/cmds/ocm/clictx/core"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	ocicpi "github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	ocmcpi "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	OCMCmdConfigType   = "ocm.cmd.config" + common.TypeGroupSuffix
	OCMCmdConfigTypeV1 = OCMCmdConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterConfigType(OCMCmdConfigType, cpi.NewConfigType(OCMCmdConfigType, &ConfigSpec{}))
	cpi.RegisterConfigType(OCMCmdConfigTypeV1, cpi.NewConfigType(OCMCmdConfigTypeV1, &ConfigSpec{}))
}

// ConfigSpec describes a memory based repository interface.
type ConfigSpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	OCMRepositories             map[string]*ocmcpi.GenericRepositorySpec `json:"ocmRepositories,omitempty"`
	OCIRepositories             map[string]*ocicpi.GenericRepositorySpec `json:"ociRepositories,omitempty"`
}

// NewConfigSpec creates a new memory ConfigSpec
func NewConfigSpec() *ConfigSpec {
	return &ConfigSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(OCMCmdConfigType),
	}
}

func (a *ConfigSpec) GetType() string {
	return OCMCmdConfigType
}

func (a *ConfigSpec) AddOCIRepository(name string, spec ocicpi.RepositorySpec) error {
	g, err := ocicpi.ToGenericRepositorySpec(spec)
	if err != nil {
		return err
	}
	if a.OCIRepositories == nil {
		a.OCIRepositories = map[string]*ocicpi.GenericRepositorySpec{}
	}
	a.OCIRepositories[name] = g
	return nil
}

func (a *ConfigSpec) AddOCMRepository(name string, spec ocmcpi.RepositorySpec) error {
	g, err := ocmcpi.ToGenericRepositorySpec(spec)
	if err != nil {
		return err
	}
	if a.OCMRepositories == nil {
		a.OCMRepositories = map[string]*ocmcpi.GenericRepositorySpec{}
	}

	a.OCMRepositories[name] = g
	return nil
}

func (a *ConfigSpec) ApplyTo(ctx config.Context, target interface{}) error {
	t, ok := target.(core.Context)
	if !ok {
		return config.ErrNoContext(OCMCmdConfigType)
	}
	for n, s := range a.OCIRepositories {
		t.OCI().Context().SetAlias(n, s)
	}
	for n, s := range a.OCMRepositories {
		t.OCM().Context().SetAlias(n, s)
	}
	return nil
}
