package artifacthdlr

import (
	"ocm.software/ocm/api/oci"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/common/output"
)

////////////////////////////////////////////////////////////////////////////////

func ClosureExplode(opts *output.Options, e interface{}) []interface{} {
	return traverse(common.History{}, e.(*Object), opts.Context)
}

func traverse(hist common.History, o *Object, octx out.Context) []output.Object {
	blob, err := o.Artifact.Blob()
	if err != nil {
		out.Errf(octx, "unable to get artifact blob: %s", err)

		return nil
	}
	key := common.NewNameVersion("", blob.Digest().String())
	if err := hist.Add(oci.KIND_OCIARTIFACT, key); err != nil {
		out.Errf(octx, "unable to add artifact to history: %s", err)

		return nil
	}
	result := []output.Object{o}
	if o.Artifact.IsIndex() {
		refs := o.Artifact.IndexAccess().GetDescriptor().Manifests

		found := map[common.NameVersion]bool{}
		for _, ref := range refs {
			key := common.NewNameVersion("", ref.Digest.String())
			if found[key] {
				continue // skip same ref wit different attributes for recursion
			}
			found[key] = true
			nested, err := o.Namespace.GetArtifact(key.GetVersion())
			if err != nil {
				out.Errf(octx, "Warning: lookup nested artifact %q [%s]: %s\n", ref.Digest, hist, err)
			}
			version, err := Key(nested)
			if err != nil {
				out.Errf(octx, "Failed to find nested key %q [%s]: %s\n", ref.Digest, hist, err)
			}
			obj := &Object{
				History: hist.Copy(),
				Key:     version,
				Spec: oci.RefSpec{
					UniformRepositorySpec: o.Spec.UniformRepositorySpec,
					ArtSpec: oci.ArtSpec{
						Repository: o.Spec.Repository,
						ArtVersion: oci.ArtVersion{Digest: &ref.Digest},
					},
				},
				Namespace: o.Namespace,
				Artifact:  nested,
			}
			result = append(result, traverse(hist, obj, octx)...)
		}
	}
	output.Print(result, "traverse %s", blob.Digest())
	return result
}
