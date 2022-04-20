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

package artefacthdlr

import (
	"strings"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/opencontainers/go-digest"
)

func Attachment(d digest.Digest, suffix string) string {
	return strings.Replace(d.String(), ":", ".", 1) + "." + suffix
}

var ExplodeAttached = processing.Explode(explodeAttached)

func explodeAttached(o interface{}) []interface{} {
	obj := o.(*Object)
	result := []interface{}{o}
	blob, _ := obj.Artefact.Blob()
	dig := blob.Digest()
	prefix := Attachment(dig, "")
	list, err := obj.Namespace.ListTags()
	hist := append(obj.History.Copy(), common.NewNameVersion("", dig.String()))
	if err == nil {
		for _, l := range list {
			if strings.HasPrefix(l, prefix) {
				a, err := obj.Namespace.GetArtefact(l)
				if err == nil {
					t := l
					s := obj.Spec
					s.Tag = &t
					s.Digest = nil
					att := &Object{
						History:    hist,
						Spec:       s,
						AttachKind: l[len(prefix):],
						Namespace:  obj.Namespace,
						Artefact:   a,
					}
					result = append(result, explodeAttached(att)...)
				}
			}
		}
	}
	output.Print(result, "attached %s", dig)
	return result
}
