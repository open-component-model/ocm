package artifacthdlr

import (
	"fmt"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/tree"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

func Elem(e interface{}) oci.ArtifactAccess {
	return e.(*Object).Artifact
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	History    common.History
	Key        common.NameVersion
	Spec       oci.RefSpec
	AttachKind string
	Namespace  oci.NamespaceAccess
	Artifact   oci.ArtifactAccess
}

var (
	_ common.HistoryElement = (*Object)(nil)
	_ tree.Object           = (*Object)(nil)
	_ tree.Typed            = (*Object)(nil)
)

func (o *Object) GetHistory() common.History {
	return o.History
}

func (o *Object) GetKey() common.NameVersion {
	return o.Key
}

func (o *Object) GetKind() string {
	return o.AttachKind
}

func (o *Object) IsNode() *common.NameVersion {
	blob, err := o.Artifact.Blob()
	if err != nil {
		logging.DefaultContext().Logger().LogError(err, "failed to fetch blob from artifact")

		return nil
	}

	nv := common.NewNameVersion("", blob.Digest().String())
	return &nv
}

func (o *Object) AsManifest() interface{} {
	var digest string
	b, err := o.Artifact.Blob()
	if err == nil {
		digest = b.Digest().String()
	} else {
		digest = err.Error()
	}
	return &Manifest{
		Spec:     o.Spec,
		Digest:   digest,
		Manifest: o.Artifact.GetDescriptor(),
	}
}

func (o *Object) String() string {
	blob, err := o.Artifact.Blob()
	if err != nil {
		return ""
	}

	dig := blob.Digest()
	tag := "-"
	if o.Spec.Tag != nil {
		tag = *o.Spec.Tag
	}
	return fmt.Sprintf("%s [%s]: %s", dig, tag, o.History)
}

type Manifest struct {
	Spec     oci.RefSpec
	Digest   string
	Manifest *artdesc.Artifact
}

////////////////////////////////////////////////////////////////////////////////

func Key(a oci.ArtifactAccess) (common.NameVersion, error) {
	blob, err := a.Blob()
	if err != nil {
		return common.NameVersion{}, fmt.Errorf("unable to determine blob name: %w", err)
	}

	return common.NewNameVersion("", blob.Digest().String()), nil
}

type TypeHandler struct {
	octx     clictx.OCI
	session  oci.Session
	repobase oci.Repository
}

func NewTypeHandler(octx clictx.OCI, session oci.Session, repobase oci.Repository) utils.TypeHandler {
	return &TypeHandler{
		octx:     octx,
		session:  session,
		repobase: repobase,
	}
}

func (h *TypeHandler) Close() error {
	return h.session.Close()
}

func (h *TypeHandler) All() ([]output.Object, error) {
	return h.all(h.repobase)
}

func (h *TypeHandler) all(repo oci.Repository) ([]output.Object, error) {
	if repo == nil {
		return nil, nil
	}
	lister := repo.NamespaceLister()
	if lister == nil {
		return nil, nil
	}
	list, err := lister.GetNamespaces("", true)
	if err != nil {
		return nil, err
	}
	var result []output.Object
	for _, l := range list {
		part, err := h.get(repo, utils.StringSpec(l))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
		}
		result = append(result, part...)
	}
	output.Print(result, "all")
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	result, err := h.get(h.repobase, elemspec)
	output.Print(result, "get %s", elemspec)
	return result, err
}

func (h *TypeHandler) get(repo oci.Repository, elemspec utils.ElemSpec) ([]output.Object, error) {
	var namespace oci.NamespaceAccess
	var result []output.Object
	var err error

	name := elemspec.String()
	spec := oci.RefSpec{}
	if repo == nil {
		evaluated, err := h.session.EvaluateRef(h.octx.Context(), name)
		if err != nil {
			return nil, errors.Wrapf(err, "repository %q", name)
		}
		if evaluated.Namespace == nil {
			return h.all(evaluated.Repository)
		}
		spec = evaluated.Ref
		namespace = evaluated.Namespace
		if evaluated.Artifact != nil {
			key, err := Key(evaluated.Artifact)
			if err != nil {
				return nil, fmt.Errorf("unable to determine key for artifact %q: %w", name, err)
			}

			obj := &Object{
				Key:       key,
				Spec:      spec,
				Namespace: namespace,
				Artifact:  evaluated.Artifact,
			}
			result = append(result, obj)
			return result, nil
		}
	} else {
		art := &oci.ArtSpec{Repository: ""}
		if name != "" {
			art, err = oci.ParseArt(name)
			if err != nil {
				return nil, errors.Wrapf(err, "artifact reference %q", name)
			}
		}
		namespace, err = h.session.LookupNamespace(repo, art.Repository)
		if err != nil {
			return nil, errors.Wrapf(err, "reference %q", name)
		}
		spec.UniformRepositorySpec = *repo.GetSpecification().UniformRepositorySpec()
		spec.Repository = art.Repository
		spec.Tag = art.Tag
		spec.Digest = art.Digest
	}

	if spec.IsVersion() {
		a, err := namespace.GetArtifact(spec.Version())
		if err != nil {
			return nil, err
		}
		h.session.AddCloser(a)
		key, err := Key(a)
		if err != nil {
			return nil, fmt.Errorf("unable to determine key for artifact %q: %w", name, err)
		}
		obj := &Object{
			Key:       key,
			Spec:      spec,
			Namespace: namespace,
			Artifact:  a,
		}
		result = append(result, obj)
	} else {
		tags, err := namespace.ListTags()
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			a, err := namespace.GetArtifact(tag)
			if err != nil {
				return nil, err
			}
			h.session.AddCloser(a)
			t := tag
			s := spec
			s.Tag = &t
			key, err := Key(a)
			if err != nil {
				return nil, fmt.Errorf("unable to determine key for artifact %q: %w", name, err)
			}

			result = append(result, &Object{
				Key:       key,
				Spec:      s,
				Namespace: namespace,
				Artifact:  a,
			})
		}
	}
	return result, nil
}
