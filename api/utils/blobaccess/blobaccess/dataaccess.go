package blobaccess

import (
	"bytes"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/ioutils"
)

type _nopCloser = ioutils.NopCloser

////////////////////////////////////////////////////////////////////////////////

type readerAccess struct {
	_nopCloser
	reader func() (io.ReadCloser, error)
	origin string
}

var _ DataSource = (*readerAccess)(nil)

func DataAccessForReaderFunction(reader func() (io.ReadCloser, error), origin string) DataAccess {
	return &readerAccess{reader: reader, origin: origin}
}

func (a *readerAccess) Get() (data []byte, err error) {
	r, err := a.Reader()
	if err != nil {
		return nil, err
	}
	defer errors.PropagateError(&err, r.Close)

	buf := bytes.Buffer{}
	_, err = io.Copy(&buf, r)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %s", a.origin)
	}
	return buf.Bytes(), nil
}

func (a *readerAccess) Reader() (io.ReadCloser, error) {
	r, err := a.reader()
	if err != nil {
		return nil, errors.Wrapf(err, "errors getting reader for %s", a.origin)
	}
	return r, nil
}

func (a *readerAccess) Origin() string {
	return a.origin
}
