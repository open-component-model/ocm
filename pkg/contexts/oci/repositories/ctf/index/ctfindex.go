// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"encoding/json"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
)

const SchemaVersion = 1

type ArtefactIndex struct {
	specs.Versioned
	Index []ArtefactMeta `json:"artefacts"`
}

type ArtefactMeta struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag,omitempty"`
	Digest     digest.Digest `json:"digest,omitempty"`
}

func Decode(data []byte) (*ArtefactIndex, error) {
	var d ArtefactIndex

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func Encode(d *ArtefactIndex) ([]byte, error) {
	return json.Marshal(d)
}
