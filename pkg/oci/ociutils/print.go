// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ociutils

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
)

type InfoHandler interface {
	Info(m cpi.ManifestAccess) string
}

var lock sync.Mutex
var handlers = map[string]InfoHandler{}

func RegisterInfoHandler(mime string, h InfoHandler) {
	lock.Lock()
	defer lock.Unlock()
	handlers[mime] = h
}

func getHandler(mime string) InfoHandler {
	lock.Lock()
	defer lock.Unlock()
	return handlers[mime]
}

func PrintArtefact(art cpi.ArtefactAccess) string {
	if art.IsManifest() {
		return fmt.Sprintf("type: %s\n", artdesc.MediaTypeImageManifest) + PrintManifest(art.ManifestAccess())
	}
	if art.IsIndex() {
		return fmt.Sprintf("type: %s\n", artdesc.MediaTypeImageIndex+PrintIndex(art.IndexAccess()))
	}
	return "unspecific"
}

func indent(orig string, gap string) string {
	s := ""
	for _, l := range strings.Split(orig, "\n") {
		s += gap + l + "\n"
	}
	return s
}

func PrintManifest(m cpi.ManifestAccess) string {
	man := m.GetDescriptor()
	s := "config:\n"
	s += fmt.Sprintf("  type:   %s\n", man.Config.MediaType)
	s += fmt.Sprintf("  digest: %s\n", man.Config.Digest)

	h := getHandler(man.Config.MediaType)
	if h != nil {
		s += indent(h.Info(m), "  ")
	}
	s += "layers:\n"
	for _, l := range man.Layers {
		s += fmt.Sprintf("- type:   %s\n", l.MediaType)
		s += fmt.Sprintf("  digest: %s\n", l.Digest)
	}
	return s
}

func PrintIndex(i cpi.IndexAccess) string {
	s := "manifests:\n"
	for _, l := range i.GetDescriptor().Manifests {
		s += fmt.Sprintf("- type:   %s\n", l.MediaType)
		s += fmt.Sprintf("  digest: %s\n", l.Digest)
		a, err := i.GetArtefact(l.Digest)
		if err != nil {
			s += fmt.Sprintf("  error: %s\n", err)
		} else {
			s += fmt.Sprintf("  resolved artefact:\n")
			s += indent(PrintArtefact(a), "    ")
		}
	}
	return s
}
