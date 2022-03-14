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

package oci

import (
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
)

func AsTags(tag string) []string {
	if tag != "" {
		return []string{tag}
	}
	return nil
}

func TransferArtefact(art cpi.ArtefactAccess, set cpi.ArtefactSink, tags ...string) error {
	if art.GetDescriptor().IsIndex() {
		return TransferIndex(art.IndexAccess(), set, tags...)
	} else {
		return TransferManifest(art.ManifestAccess(), set, tags...)
	}
}

func TransferIndex(art cpi.IndexAccess, set cpi.ArtefactSink, tags ...string) error {
	for _, l := range art.GetDescriptor().Manifests {
		art, err := art.GetArtefact(l.Digest)
		if err != nil {
			return errors.Wrapf(err, "getting indexed artefact %s", l.Digest)
		}
		err = TransferArtefact(art, set)
		if err != nil {
			return errors.Wrapf(err, "transferring indexed artefact %s", l.Digest)
		}
	}
	_, err := set.AddArtefact(art, tags...)
	if err != nil {
		return errors.Wrapf(err, "transferring index artefact")
	}
	return err
}

func TransferManifest(art cpi.ManifestAccess, set cpi.ArtefactSink, tags ...string) error {
	blob, err := art.GetConfigBlob()
	if err != nil {
		return errors.Wrapf(err, "getting config blob")
	}
	err = set.AddBlob(blob)
	if err != nil {
		return errors.Wrapf(err, "transferring config blob")
	}
	for _, l := range art.GetDescriptor().Layers {
		blob, err = art.GetBlob(l.Digest)
		if err != nil {
			return errors.Wrapf(err, "getting layer blob %s", l.Digest)
		}
		err = set.AddBlob(blob)
		if err != nil {
			return errors.Wrapf(err, "transferring layer blob %s", l.Digest)
		}
	}
	_, err = set.AddArtefact(art, tags...)
	if err != nil {
		return errors.Wrapf(err, "transferring image artefact")
	}
	return err
}

func EvaluateRefWithContext(ctx Context, ref string) (*RefSpec, NamespaceAccess, error) {
	parsed, err := ParseRef(ref)
	if err != nil {
		return nil, nil, err
	}
	spec := ocireg.NewRepositorySpec(parsed.Base())
	repo, err := ctx.RepositoryForSpec(spec)
	if err != nil {
		return nil, nil, err
	}
	ns, err := repo.LookupNamespace(parsed.Repository)
	return &parsed, ns, err
}
