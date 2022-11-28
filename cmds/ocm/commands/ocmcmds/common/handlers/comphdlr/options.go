// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comphdlr

import (
	"github.com/Masterminds/semver/v3"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Option interface {
	ApplyToCompHandler(handler *TypeHandler)
}

type Options []Option

func (o Options) ApplyToCompHandler(handler *TypeHandler) {
	for _, e := range o {
		e.ApplyToCompHandler(handler)
	}
}

func OptionsFor(o options.OptionSetProvider) Options {
	var hopts []Option
	if constr := versionconstraintsoption.From(o); constr != nil {
		if len(constr.Constraints) > 0 {
			hopts = append(hopts, WithVersionConstraints(constr.Constraints))
		}
		if constr.Latest {
			hopts = append(hopts, LatestOnly())
		}
	}
	return hopts
}

////////////////////////////////////////////////////////////////////////////////

type constraints struct {
	constraints []*semver.Constraints
}

func (o constraints) ApplyToCompHandler(handler *TypeHandler) {
	handler.constraints = o.constraints
}

func WithVersionConstraints(c []*semver.Constraints) Option {
	return constraints{c}
}

////////////////////////////////////////////////////////////////////////////////

type latestonly struct {
	flag bool
}

func (o latestonly) ApplyToCompHandler(handler *TypeHandler) {
	handler.latest = o.flag
}

func LatestOnly(b ...bool) Option {
	return latestonly{utils.OptionalDefaultedBool(true, b...)}
}
