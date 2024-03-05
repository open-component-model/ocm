// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
)

type _Artifact = artdesc.Artifact

type modifiedArtifact struct {
	state accessobj.State
	*_Artifact
}

var _ Artifact = (*modifiedArtifact)(nil)

// NewArtifact provides a copy of the given artifact,
// which will keep track of the original serialization.
// This one is used as long as the artifact is unchanged.
// (the returned underlying implementation of the Artifact interface
// might differ from the original one).
func NewArtifact(art Artifact) (Artifact, error) {
	blob, err := art.Blob()
	if err != nil {
		return nil, err
	}
	state, err := accessobj.NewBlobStateForBlob(accessobj.ACC_WRITABLE, blob, NewArtifactStateHandler())
	if err != nil {
		return nil, err
	}

	return &modifiedArtifact{
		_Artifact: state.GetState().(*_Artifact),
		state:     state,
	}, nil
}

func (a *modifiedArtifact) Blob() (BlobAccess, error) {
	return a.state.GetBlob()
}

func (a *modifiedArtifact) Digest() digest.Digest {
	blob, _ := a.Blob()
	if blob != nil {
		return blob.Digest()
	}
	return ""
}
