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

package index

type RepositoryIndex struct {
	byDigest     map[string][]*ArtefactMeta
	byRepository map[string]map[string]*ArtefactMeta
}

func NewRepositoryIndex() *RepositoryIndex {
	return &RepositoryIndex{
		byDigest:     map[string][]*ArtefactMeta{},
		byRepository: map[string]map[string]*ArtefactMeta{},
	}
}

func (r *RepositoryIndex) AddArtefact(n *ArtefactMeta) {
	m := *n

	list := r.byDigest[m.Digest]
	if list == nil {
		list = []*ArtefactMeta{&m}
	} else {
		for _, e := range list {
			if *e == m {
				return
			}
		}
		list = append(list, &m)
	}
	r.byDigest[m.Digest] = list

	repos := r.byRepository[m.Repository]
	if len(repos) == 0 {
		repos = map[string]*ArtefactMeta{}
		r.byRepository[m.Repository] = repos
	}
	repos[m.Digest] = &m
	repos[m.Tag] = &m
}

func (r *RepositoryIndex) GetArtefacts(digest string) []*ArtefactMeta {
	return r.byDigest[digest]
}

func (r *RepositoryIndex) GetArtefact(repo, tag string) *ArtefactMeta {
	repos := r.byRepository[repo]
	if repos == nil {
		return nil
	}
	m := repos[tag]
	if m == nil {
		return nil
	}
	result := *m
	return &result
}
