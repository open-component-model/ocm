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

package docker

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/containers/image/v5/docker/daemon"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
)

var dummyContext = context.Background()

var pattern = regexp.MustCompile("^[0-9a-f]{12}$")

func ParseGenericRef(ref string) (string, string, error) {
	if strings.TrimSpace(ref) == "" {
		return "", "", fmt.Errorf("invalid docker reference %q", ref)
	}
	parts := strings.Split(ref, ":")
	if len(parts) > 2 {
		return "", "", fmt.Errorf("invalid docker reference %q", ref)
	}
	if len(parts) == 1 {
		// expect docker id
		if pattern.MatchString(parts[0]) {
			return "", parts[0], nil
		}
	}
	_, err := daemon.ParseReference(ref)
	if err != nil {
		return "", "", err
	}
	return parts[0], parts[1], nil
}

func ParseRef(name, version string) (types.ImageReference, error) {
	if version == "" || name == "" {
		id := version
		if id == "" {
			id = name
		}
		// check for docker daemon image id
		if pattern.MatchString(id) {
			// this definately no digest, but the library expects it this way
			return daemon.NewReference(digest.Digest(id), nil)
		}
		return nil, fmt.Errorf("no docker daemon image id")
	}
	return daemon.ParseReference(name + ":" + version)
}

func ImageId(art cpi.Artefact) digest.Digest {
	m, err := art.Manifest()
	if err != nil {
		return ""
	}
	return digest.Digest(m.Config.Digest.Hex()[:12])
}

// TODO add cache

type dataAccess struct {
	accessio.NopCloser
	lock   sync.Mutex
	info   types.BlobInfo
	src    types.ImageSource
	reader io.ReadCloser
}

var _ cpi.DataAccess = (*dataAccess)(nil)

func NewDataAccess(src types.ImageSource, info types.BlobInfo, delayed bool) (*dataAccess, error) {
	var reader io.ReadCloser
	var err error

	if !delayed {
		reader, _, err = src.GetBlob(context.Background(), info, nil)
		if err != nil {
			return nil, err
		}
	}
	return &dataAccess{
		info:   info,
		src:    src,
		reader: reader,
	}, nil
}

func (d *dataAccess) Get() ([]byte, error) {
	return readAll(d.Reader())
}

func (d *dataAccess) Reader() (io.ReadCloser, error) {
	d.lock.Lock()
	reader := d.reader
	d.reader = nil
	d.lock.Unlock()
	if reader != nil {
		return reader, nil
	}
	reader, _, err := d.src.GetBlob(context.Background(), d.info, nil)
	return reader, err
}

func readAll(reader io.ReadCloser, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return data, nil
}
