package transfer

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/generics"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/tools/transfer/filters"
	"ocm.software/ocm/api/utils/logging"
)

func TransferArtifact(art cpi.ArtifactAccess, set cpi.ArtifactSink, tags ...string) error {
	_, err := TransferArtifactWithFilter(art, set, nil, tags...)
	return err
}

func TransferArtifactWithFilter(art cpi.ArtifactAccess, set cpi.ArtifactSink, filter filters.Filter, tags ...string) (*digest.Digest, error) {
	if art.GetDescriptor().IsIndex() {
		return TransferIndexWithFilter(art.IndexAccess(), set, filter, tags...)
	} else {
		if filter != nil && !filter.Accept(art, nil) {
			return nil, errors.ErrNoMatch(cpi.KIND_OCIARTIFACT, art.Digest().String())
		}
		return generics.Pointer(art.Digest()), TransferManifest(art.ManifestAccess(), set, tags...)
	}
}

func TransferIndex(art cpi.IndexAccess, set cpi.ArtifactSink, tags ...string) error {
	_, err := TransferIndexWithFilter(art, set, nil, tags...)
	return err
}

func TransferIndexWithFilter(art cpi.IndexAccess, set cpi.ArtifactSink, filter filters.Filter, tags ...string) (dig *digest.Digest, err error) {
	logging.Logger().Debug("transfer OCI index", "digest", art.Digest())
	defer func() {
		logging.Logger().Debug("transfer OCI index done", "error", logging.ErrorMessage(err))
	}()

	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	// provide a local copy of the inbound artifact (index)
	// which keeps track of the original serialized form,
	// which is used if the manifest is left unchanged after
	// filtering.
	modified, err := cpi.NewArtifact(art)
	if err != nil {
		return nil, err
	}

	// don't add this (index) object later.
	// it implements the same interface, but would
	// provide a new serialization, which might differ
	// from the original one even if it is unchanged,
	// which would change the digest.
	index, err := modified.Index()
	if err != nil {
		return nil, err
	}

	ign := 0
	for i, l := range art.GetDescriptor().Manifests {
		loop := finalize.Nested()
		logging.Logger().Debug("indexed manifest", "digest", l.Digest, "size", l.Size)
		art, err := art.GetArtifact(l.Digest)
		if err != nil {
			return nil, errors.Wrapf(err, "getting indexed artifact %s", l.Digest)
		}
		loop.Close(art)
		if filter == nil || filter.Accept(art, l.Platform) {
			err = TransferArtifact(art, set)
			if err != nil {
				return nil, errors.Wrapf(err, "transferring indexed artifact %s", l.Digest)
			}
			dig = generics.Pointer(l.Digest)
		} else {
			index.Manifests = append(index.Manifests[:i-ign], index.Manifests[i-ign+1:]...)
			ign++
		}
		err = loop.Finalize()
		if err != nil {
			return nil, err
		}
	}

	if filter != nil {
		switch len(art.GetDescriptor().Manifests) - ign {
		case 0:
			return nil, errors.ErrNoMatch(cpi.KIND_OCIARTIFACT, art.Digest().String())
		case 1:
			if len(tags) > 0 {
				err := set.AddTags(*dig, tags...)
				if err != nil {
					return nil, err
				}
			}
			return dig, nil
		}
	}

	_, err = set.AddArtifact(modified, tags...)
	if err != nil {
		return nil, errors.Wrapf(err, "transferring index artifact")
	}
	return generics.Pointer(modified.Digest()), err
}

func TransferManifest(art cpi.ManifestAccess, set cpi.ArtifactSink, tags ...string) (err error) {
	logging.Logger().Debug("transfer OCI manifest", "digest", art.Digest())
	defer func() {
		logging.Logger().Debug("transfer OCI manifest done", "error", logging.ErrorMessage(err))
	}()

	blob, err := art.GetConfigBlob()
	if err != nil {
		return errors.Wrapf(err, "getting config blob")
	}
	err = set.AddBlob(blob)
	blob.Close()
	if err != nil {
		return errors.Wrapf(err, "transferring config blob")
	}
	for i, l := range art.GetDescriptor().Layers {
		logging.Logger().Debug("layer", "digest", "digest", l.Digest, "size", l.Size, "index", i)
		blob, err = art.GetBlob(l.Digest)
		if err != nil {
			return errors.Wrapf(err, "getting layer blob %s", l.Digest)
		}
		err = set.AddBlob(blob)
		blob.Close()
		if err != nil {
			return errors.Wrapf(err, "transferring layer blob %s", l.Digest)
		}
	}
	blob, err = set.AddArtifact(art, tags...)
	if err != nil {
		return errors.Wrapf(err, "transferring image artifact")
	}
	return blob.Close()
}
