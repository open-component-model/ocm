// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dirtree

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/registrations"
)

func init() {
	download.RegisterHandlerRegistrationHandler("ocm/dirtree", &RegistrationHandler{})
}

type Config struct {
	AsArchive   bool     `json:"asArchive"`
	ConfigTypes []string `json:"configTypes"`
}

type RegistrationHandler struct{}

var _ download.HandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx download.Target, config download.HandlerConfig, olist ...download.HandlerOption) (bool, error) {
	var err error

	if handler != "" {
		return true, fmt.Errorf("invalid dirtree handler %q", handler)
	}

	attr, err := registrations.DecodeConfig[Config](config)
	if err != nil {
		return true, errors.Wrapf(err, "cannot unmarshal download handler configuration")
	}

	opts := download.NewHandlerOptions(olist...)
	if opts.MimeType != "" && !slices.Contains(supportedMimeTypes, opts.MimeType) {
		return true, errors.Wrapf(err, "mime type %s not supported", opts.MimeType)
	}
	if opts.ArtifactType != "" && slices.Contains(defaultArtifactTypes, opts.ArtifactType) && !attr.AsArchive {
		return true, nil
	}

	h := New(attr.ConfigTypes...).SetArchiveMode(attr.AsArchive)
	supported := generics.Conditional(len(attr.ConfigTypes) > 0, attr.ConfigTypes, supportedMimeTypes)
	if opts.MimeType == "" {
		for _, m := range supported {
			opts.MimeType = m
			download.For(ctx).Register(opts.ArtifactType, opts.MimeType, h)
		}
	} else {
		download.For(ctx).Register(opts.ArtifactType, opts.MimeType, h)
	}

	return true, nil
}
