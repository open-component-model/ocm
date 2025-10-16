package artifacthdlr

import (
	"strings"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/logging"
	"github.com/opencontainers/go-digest"

	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
)

func Attachment(d digest.Digest, suffix string) string {
	return strings.Replace(d.String(), ":", "-", 1) + "." + suffix
}

var ExplodeAttached = processing.Explode(explodeAttached)

func explodeAttached(o interface{}) []interface{} {
	// internal function must be called correctly, otherwise early panic
	obj := o.(*Object)
	result := []interface{}{o}
	blob, err := obj.Artifact.Blob()
	if err != nil {
		logging.DefaultContext().Logger().LogError(err, "failed to fetch blob from artifact")

		return nil
	}
	dig := blob.Digest()
	prefix := Attachment(dig, "")
	list, err := obj.Namespace.ListTags()
	hist := sliceutils.CopyAppend(obj.History, common.NewNameVersion("", dig.String()))
	if err == nil {
		for _, l := range list {
			if strings.HasPrefix(l, prefix) {
				a, err := obj.Namespace.GetArtifact(l)
				if err == nil {
					t := l
					s := obj.Spec
					s.Tag = &t
					s.Digest = nil
					key, err := Key(a)
					if err != nil {
						// this list ignores errors as this segment only happens when err == nil.
						continue
					}

					att := &Object{
						History:    hist,
						Key:        key,
						Spec:       s,
						AttachKind: l[len(prefix):],
						Namespace:  obj.Namespace,
						Artifact:   a,
					}
					result = append(result, explodeAttached(att)...)
				}
			}
		}
	}
	output.Print(result, "attached %s", dig)
	return result
}
