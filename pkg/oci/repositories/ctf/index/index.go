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

import (
	"sort"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
)

type RepositoryIndex struct {
	byDigest     map[digest.Digest][]*ArtefactMeta
	byRepository map[string]map[string]*ArtefactMeta
}

func NewRepositoryIndex() *RepositoryIndex {
	return &RepositoryIndex{
		byDigest:     map[digest.Digest][]*ArtefactMeta{},
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
	repos[m.Digest.String()] = &m
	repos[m.Tag] = &m
}

func (r *RepositoryIndex) HasArtefact(repo, tag string) bool {
	repos := r.byRepository[repo]
	if repos == nil {
		return false
	}
	m := repos[tag]
	return m != nil
}

func (r *RepositoryIndex) GetArtefacts(digest digest.Digest) []*ArtefactMeta {
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

func (r *RepositoryIndex) GetDescriptor() *ArtefactIndex {
	index := &ArtefactIndex{
		Versioned: specs.Versioned{SchemaVersion},
	}

	repos := make([]string, len(r.byRepository))
	i := 0
	for repo := range r.byRepository {
		repos[i] = repo
		i++
	}
	sort.Strings(repos)
	for _, name := range repos {
		repo := r.byRepository[name]
		versions := make([]string, len(repo))
		i := 0
		for vers := range repo {
			versions[i] = vers
			i++
		}
		sort.Strings(repos)

		for _, name := range versions {
			vers := repo[name]
			d := &ArtefactMeta{
				Repository: vers.Repository,
				Tag:        vers.Tag,
				Digest:     vers.Digest,
			}
			index.Index = append(index.Index, *d)
		}
	}
	return index
}
