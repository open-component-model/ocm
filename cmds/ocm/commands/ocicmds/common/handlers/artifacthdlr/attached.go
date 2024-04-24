// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package artifacthdlr

import (
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/generics"
)

func Attachment(d digest.Digest, suffix string) string {
	return strings.Replace(d.String(), ":", "-", 1) + "." + suffix
}

var ExplodeAttached = processing.Explode(explodeAttached)

func explodeAttached(o interface{}) []interface{} {
	obj, ok := o.(*Object)
	if !ok {
		return nil
	}
	result := []interface{}{o}
	blob, err := obj.Artifact.Blob()
	if err != nil {
		return result
	}
	dig := blob.Digest()
	prefix := Attachment(dig, "")
	list, err := obj.Namespace.ListTags()
	hist := generics.AppendedSlice(obj.History, common.NewNameVersion("", dig.String()))
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
						// this is questionable behaviour. :think:
						return nil
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
