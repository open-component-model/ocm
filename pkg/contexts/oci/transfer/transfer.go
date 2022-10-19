// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

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
