// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/helm"
	"github.com/open-component-model/ocm/pkg/helm/credentials"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type for a blob in an OCI repository.
const (
	Type   = "helm"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterAccessType(cpi.NewAccessSpecType(Type, &AccessSpec{}, cpi.WithDescription(usage)))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(TypeV1, &AccessSpec{}, cpi.WithFormatSpec(formatV1), cpi.WithConfigHandler(ConfigHandler())))
}

// New creates a new Helm Chart accessor for helm repositories.
func New(chart string, repourl string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		HelmChart:           chart,
		HelmRepository:      repourl,
	}
}

// AccessSpec describes the access for a oci registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// HelmRepository is the URL og the helm repository to load the chart from.
	HelmRepository string `json:"helmRepository"`

	// HelmChart if the name of the helm chart and its version separated by a colon.
	HelmChart string `json:"helmChart"`

	// CACert is an optional root TLS certificate
	CACert string `json:"caCert,omitempty"`

	// Keyring is an optional keyring to verify the chart.
	Keyring string `json:"keyring,omitempty"`
}

var _ cpi.AccessSpec = (*AccessSpec)(nil)

func (a *AccessSpec) Describe(ctx cpi.Context) string {
	return fmt.Sprintf("Helm chart %s in repository %s", a.HelmChart, a.HelmRepository)
}

func (s *AccessSpec) IsLocal(context cpi.Context) bool {
	return false
}

func (s *AccessSpec) GlobalAccessSpec(ctx cpi.Context) cpi.AccessSpec {
	return s
}

func (s *AccessSpec) GetMimeType() string {
	return helm.ChartMediaType
}

func (s *AccessSpec) AccessMethod(access cpi.ComponentVersionAccess) (cpi.AccessMethod, error) {
	return &accessMethod{comp: access, spec: s}, nil
}

////////////////////////////////////////////////////////////////////////////////

type accessMethod struct {
	lock sync.Mutex
	blob accessio.BlobAccess
	comp cpi.ComponentVersionAccess
	spec *AccessSpec

	acc helm.ChartAccess
}

var _ cpi.AccessMethod = (*accessMethod)(nil)

func (m *accessMethod) GetKind() string {
	return Type
}

func (m *accessMethod) AccessSpec() cpi.AccessSpec {
	return m.spec
}

func (m *accessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.blob != nil {
		m.blob.Close()
		m.acc.Close()
		m.blob = nil
	}
	return nil
}

func (m *accessMethod) Get() ([]byte, error) {
	return accessio.BlobData(m.getBlob())
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	return accessio.BlobReader(m.getBlob())
}

func (m *accessMethod) MimeType() string {
	return helm.ChartMediaType
}

func (m *accessMethod) getBlob() (cpi.BlobAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.blob != nil {
		return m.blob, nil
	}

	parts := strings.Split(m.spec.HelmChart, ":")
	if len(parts) != 2 {
		return nil, errors.ErrInvalid("helm chart ref", m.spec.HelmRepository)
	}
	acc, err := helm.DownloadChart(os.Stdout, parts[0], parts[1], m.spec.HelmRepository,
		helm.WithCredentials(credentials.GetCredentials(m.comp.GetContext(), m.spec.HelmRepository)),
		helm.WithKeyring([]byte(m.spec.Keyring)),
		helm.WithRootCert([]byte(m.spec.CACert)))
	if err != nil {
		return nil, err
	}
	m.blob, err = acc.Chart()
	if err != nil {
		acc.Close()
	}
	m.acc = acc
	return m.blob, nil
}
