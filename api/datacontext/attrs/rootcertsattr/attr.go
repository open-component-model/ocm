package rootcertsattr

import (
	"crypto/x509"
	"encoding/json"
	"sync"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ATTR_KEY   = "github.com/mandelsoft/ocm/rootcerts"
	ATTR_SHORT = "rootcerts"
)

type (
	Context         = datacontext.AttributesContext
	ContextProvider = datacontext.ContextProvider
)

func init() {
	datacontext.RegisterAttributeType(ATTR_KEY, AttributeType{}, ATTR_SHORT)
}

type AttributeType struct{}

func (a AttributeType) Name() string {
	return ATTR_KEY
}

func (a AttributeType) Description() string {
	return `
*JSON*
General root certificate settings given as JSON document with the following
format:

<pre>
{
  "rootCertificates": [
     {
       "data": ""&lt;base64>"
     },
     {
       "path": ""&lt;file path>"
     }
  ]
}
</pre>

One of following data fields are possible:
- <code>data</code>:       base64 encoded binary data
- <code>stringdata</code>: plain text data
- <code>path</code>:       a file path to read the data from
`
}

func (a AttributeType) Encode(v interface{}, marshaller runtime.Marshaler) ([]byte, error) {
	attr, ok := v.(*Attribute)
	if !ok {
		return nil, errors.ErrInvalid("certificate attribute")
	}
	cfg := New()

	attr.lock.Lock()
	defer attr.lock.Unlock()

	for _, c := range attr.rootCertificates {
		data := signutils.CertificateToPem(c)
		cfg.AddRootCertificateData(data)
	}

	return json.Marshal(cfg)
}

func (a AttributeType) Decode(data []byte, unmarshaller runtime.Unmarshaler) (interface{}, error) {
	var value Config
	err := unmarshaller.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}

	attr := &Attribute{}
	err = value.ApplyToAttribute(attr)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

////////////////////////////////////////////////////////////////////////////////

type Attribute struct {
	lock             sync.Mutex
	rootCertificates []*x509.Certificate
}

func (a *Attribute) RegisterRootCertificates(cert signutils.GenericCertificateChain) error {
	certs, err := signutils.GetCertificateChain(cert, false)
	if err != nil {
		return err
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	a.rootCertificates = append(a.rootCertificates, certs...)
	return nil
}

func (a *Attribute) HasRootCertificates() bool {
	a.lock.Lock()
	defer a.lock.Unlock()
	return len(a.rootCertificates) > 0
}

func (a *Attribute) GetRootCertPool(system bool) *x509.CertPool {
	var pool *x509.CertPool

	if system {
		pool, _ = x509.SystemCertPool()
	}
	if pool == nil {
		pool = x509.NewCertPool()
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	for _, c := range a.rootCertificates {
		pool.AddCert(c)
	}
	return pool
}

////////////////////////////////////////////////////////////////////////////////

func Get(ctx ContextProvider) *Attribute {
	return ctx.AttributesContext().GetAttributes().GetOrCreateAttribute(ATTR_KEY, func(ctx datacontext.Context) interface{} {
		return &Attribute{}
	}).(*Attribute)
}

func Set(ctx ContextProvider, attribute *Attribute) error {
	return ctx.AttributesContext().GetAttributes().SetAttribute(ATTR_KEY, attribute)
}
