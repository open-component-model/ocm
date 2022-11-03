// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"strings"
)

type Registry[H any] struct {
	mappings map[UploaderKey]H
}

func NewRegistry[H any]() *Registry[H] {
	return &Registry[H]{
		mappings: map[UploaderKey]H{},
	}
}

func (p *Registry[H]) lookupMedia(key UploaderKey) (H, bool) {
	var zero H
	for {
		if h, ok := p.mappings[key]; ok {
			return h, true
		}
		if i := strings.LastIndex(key.MediaType, "+"); i > 0 {
			key.MediaType = key.MediaType[:i]
		} else {
			break
		}
	}
	return zero, false
}

func (p *Registry[H]) GetHandler(key UploaderKey) H {
	return p.mappings[key]
}

func (p *Registry[H]) LookupHandler(arttype, mediatype string) (H, bool) {
	key := UploaderKey{
		ArtifactType: arttype,
		MediaType:    mediatype,
	}

	h, ok := p.lookupMedia(key)
	if ok {
		return h, ok
	}

	key.MediaType = ""
	if h, ok := p.mappings[key]; ok {
		return h, ok
	}

	key.MediaType = mediatype
	key.ArtifactType = ""
	return p.lookupMedia(key)
}

func (p *Registry[H]) Register(key UploaderKey, h H) {
	p.mappings[key] = h
}
