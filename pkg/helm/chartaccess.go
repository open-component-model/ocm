// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

const (
	ChartMediaType      = "vnd.cncf.helm.chart.content.v1.tar+gzip"
	ProvenanceMediaType = "vnd.cncf.helm.chart.provenance.v1.prov"
)

type ChartAccess interface {
	io.Closer
	Chart() (accessio.BlobAccess, error)
	Prov() (accessio.BlobAccess, error)
}

func newFileAccess(c *chartAccess, path string, mime string) accessio.BlobAccess {
	c.refcnt++
	return accessio.ReferencingBlobAccess(accessio.BlobAccessForFile(mime, path, c.fs), c.unref)
}

type chartAccess struct {
	lock sync.Mutex

	closed bool
	refcnt int

	fs    vfs.FileSystem
	root  string
	chart string
	prov  string
}

var _ ChartAccess = (*chartAccess)(nil)

func newTempChartAccess(fss ...vfs.FileSystem) (*chartAccess, error) {
	fs := accessio.FileSystem(fss...)

	temp, err := vfs.TempDir(fs, "", "helmchart")
	if err != nil {
		return nil, err
	}
	return &chartAccess{
		fs:   fs,
		root: temp,
	}, nil
}

func NewChartAccessByFiles(chart, prov string, fss ...vfs.FileSystem) ChartAccess {
	return &chartAccess{
		fs:    accessio.FileSystem(fss...),
		chart: chart,
		prov:  prov,
	}
}

func (c *chartAccess) unref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.refcnt == 0 {
		return fmt.Errorf("oops: refcount is already zero")
	}
	c.refcnt--
	return nil
}

func (c *chartAccess) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.refcnt > 0 {
		return errors.ErrStillInUse("chart access")
	}

	defer func() { c.closed = true }()

	if c.root != "" && !c.closed {
		return os.RemoveAll(c.root)
	}
	return nil
}

func (c *chartAccess) Chart() (accessio.BlobAccess, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return nil, accessio.ErrClosed
	}

	return newFileAccess(c, c.chart, ChartMediaType), nil
}

func (c *chartAccess) Prov() (accessio.BlobAccess, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return nil, accessio.ErrClosed
	}
	return newFileAccess(c, c.prov, ProvenanceMediaType), nil
}
