package gardenerconfig

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	gardenercfg_cpi "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/errors"
)

var log = false

type Cipher string

const (
	Plaintext Cipher = "PLAINTEXT"
	AESECB    Cipher = "AES.ECB"
)

type Repository struct {
	ctx                       cpi.Context
	lock                      sync.RWMutex
	url                       string
	configType                gardenercfg_cpi.ConfigType
	cipher                    Cipher
	key                       []byte
	propagateConsumerIdentity bool
	creds                     map[string]cpi.Credentials
	fs                        vfs.FileSystem
}

func NewRepository(ctx cpi.Context, url string, configType gardenercfg_cpi.ConfigType, cipher Cipher, key []byte, propagateConsumerIdentity bool) (*Repository, error) {
	r := &Repository{
		ctx:                       ctx,
		url:                       url,
		configType:                configType,
		cipher:                    cipher,
		key:                       key,
		propagateConsumerIdentity: propagateConsumerIdentity,
		fs:                        vfsattr.Get(ctx),
	}
	if err := r.read(true); err != nil {
		return nil, err
	}
	return r, nil
}

var _ cpi.Repository = &Repository{}

func (r *Repository) ExistsCredentials(name string) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if err := r.read(false); err != nil {
		return false, err
	}

	return r.creds[name] != nil, nil
}

func (r *Repository) LookupCredentials(name string) (cpi.Credentials, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if err := r.read(false); err != nil {
		return nil, err
	}

	auth, ok := r.creds[name]
	if !ok {
		return nil, cpi.ErrUnknownCredentials(name)
	}

	return auth, nil
}

func (r *Repository) WriteCredentials(name string, creds cpi.Credentials) (cpi.Credentials, error) {
	return nil, errors.ErrNotSupported("write", "credentials", RepositoryType)
}

func (r *Repository) read(force bool) error {
	if !force && r.creds != nil {
		return nil
	}

	configReader, err := r.getRawConfig()
	if err != nil {
		return err
	}
	defer configReader.Close()

	r.creds = map[string]core.Credentials{}
	handler := gardenercfg_cpi.GetHandler(r.configType)
	if handler == nil {
		return errors.Newf("unable to find handler for config type %s", string(r.configType))
	}

	creds, err := handler.ParseConfig(configReader)
	if err != nil {
		return fmt.Errorf("unable to parse config: %w", err)
	}

	for _, cred := range creds {
		cred := cred
		if _, ok := r.creds[cred.Name()]; !ok {
			r.creds[cred.Name()] = cred.Properties()
		}
		if r.propagateConsumerIdentity {
			if log {
				fmt.Printf("propagate id %q\n", cred.ConsumerIdentity())
			}

			getCredentials := func() (cpi.Credentials, error) {
				return r.LookupCredentials(cred.Name())
			}
			cg := CredentialGetter{
				getCredentials: getCredentials,
			}
			r.ctx.SetCredentialsForConsumer(cred.ConsumerIdentity(), cg)
		}
	}

	return nil
}

func (r *Repository) getRawConfig() (io.ReadCloser, error) {
	u, err := url.Parse(r.url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %q: %w", r.url, err)
	}

	var reader io.ReadCloser
	if u.Scheme == "file" {
		f, err := r.fs.Open(u.Path)
		if err != nil {
			return nil, fmt.Errorf("unable to open file: %w", err)
		}
		reader = f
	} else {
		res, err := http.Get(u.String())
		if err != nil {
			return nil, fmt.Errorf("unable to get config from secret server: %w", err)
		}
		reader = res.Body
	}

	switch r.cipher {
	case AESECB:
		var srcBuf bytes.Buffer
		if _, err := io.Copy(&srcBuf, reader); err != nil {
			return nil, fmt.Errorf("unable to read body: %w", err)
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
		block, err := aes.NewCipher(r.key)
		if err != nil {
			return nil, fmt.Errorf("unable to create cipher: %w", err)
		}
		dst := make([]byte, srcBuf.Len())
		if err := ecbDecrypt(block, dst, srcBuf.Bytes()); err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewBuffer(dst)), nil
	case Plaintext:
		return reader, nil
	default:
		return nil, errors.ErrNotImplemented(string(r.cipher), RepositoryType)
	}
}

func ecbDecrypt(block cipher.Block, dst, src []byte) error {
	blockSize := block.BlockSize()
	if len(src)%blockSize != 0 {
		return fmt.Errorf("input must contain only full blocks (blocksize: %d; input length: %d)", blockSize, len(src))
	}
	if len(dst) < len(src) {
		return errors.New("destination is smaller than source")
	}
	for len(src) > 0 {
		block.Decrypt(dst, src[:blockSize])
		src = src[blockSize:]
		dst = dst[blockSize:]
	}
	return nil
}
