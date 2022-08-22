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

package ocireg

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/docker/resolve"
)

// TODO: add cache

type dataAccess struct {
	accessio.NopCloser
	lock    sync.Mutex
	fetcher remotes.Fetcher
	desc    artdesc.Descriptor
	reader  io.ReadCloser
}

var _ cpi.DataAccess = (*dataAccess)(nil)

func NewDataAccess(fetcher remotes.Fetcher, digest digest.Digest, mimeType string, delayed bool) (*dataAccess, error) {
	var reader io.ReadCloser
	var err error
	desc := artdesc.Descriptor{
		MediaType: mimeType,
		Digest:    digest,
		Size:      accessio.BLOB_UNKNOWN_SIZE,
	}
	if !delayed {
		reader, err = fetcher.Fetch(dummyContext, desc)
		if err != nil {
			return nil, err
		}
	}
	return &dataAccess{
		fetcher: fetcher,
		desc:    desc,
		reader:  reader,
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
	return d.fetcher.Fetch(dummyContext, d.desc)
}

/*
func fetch(ctx context.Context, f remotes.Fetcher, desc *artdesc.Descriptor) ([]byte, error) {
	fmt.Printf("*** fetch %s %s\n", desc.MediaType, desc.Digest)
	if desc.Size == 0 {
		desc.Size = -1
	}
	return readAll(f.Fetch(ctx, *desc))
}
*/

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

func push(ctx context.Context, p resolve.Pusher, blob accessio.BlobAccess) error {
	desc := *artdesc.DefaultBlobDescriptor(blob)
	return pushData(ctx, p, desc, blob)
}

func pushData(ctx context.Context, p resolve.Pusher, desc artdesc.Descriptor, data accessio.DataAccess) error {
	key := remotes.MakeRefKey(ctx, desc)
	if desc.Size == 0 {
		desc.Size = -1
	}
	fmt.Printf("*** push %s %s: %s\n", desc.MediaType, desc.Digest, key)
	req, err := p.Push(ctx, desc, data)
	if err != nil {
		if errdefs.IsAlreadyExists(err) {
			fmt.Printf("*** %s %s: already exists\n", desc.MediaType, desc.Digest)
			return nil
		}
		return err
	}
	return req.Commit(ctx, desc.Size, desc.Digest)
}

var dummyContext = nologger()

func nologger() context.Context {
	ctx := context.Background()
	logger := logrus.New()
	logger.Level = logrus.ErrorLevel
	return log.WithLogger(ctx, logrus.NewEntry(logger))
}
