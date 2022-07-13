package secretserver

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Cipher string

const (
	Plaintext Cipher = "PLAINTEXT"
	AESECB    Cipher = "AES.ECB"
)

type Repository struct {
	lock       sync.RWMutex
	url        string
	configName string
	cipher     Cipher
	key        []byte
	config     *SecretServerConfig
}

// SecretServerConfig is the struct that describes the secret server data structure
type SecretServerConfig struct {
	ContainerRegistry map[string]*ContainerRegistryCredentials `json:"container_registry"`
}

// ContainerRegistryCredentials describes the container registry credentials struct as given by the cc secrets server.
type ContainerRegistryCredentials struct {
	Username               string   `json:"username"`
	Password               string   `json:"password"`
	Privileges             string   `json:"privileges"`
	Host                   string   `json:"host,omitempty"`
	ImageReferencePrefixes []string `json:"image_reference_prefixes,omitempty"`
}

func NewRepository(url string, configName string, cipher Cipher, key []byte) *Repository {
	return &Repository{
		url:        url,
		configName: configName,
		cipher:     cipher,
		key:        key,
	}
}

var _ cpi.Repository = &Repository{}

func (r *Repository) ExistsCredentials(name string) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	_, ok := r.config.ContainerRegistry[name]
	return ok, nil
}

func (r *Repository) LookupCredentials(name string) (cpi.Credentials, error) {
	if err := r.Read(false); err != nil {
		return nil, err
	}
	r.lock.RLock()
	defer r.lock.RUnlock()

	auth, ok := r.config.ContainerRegistry[name]
	if !ok {
		return nil, cpi.ErrUnknownCredentials(name)
	}

	return newCredentialsFromContainerRegistryCredentials(auth), nil
}

func (r *Repository) WriteCredentials(name string, creds cpi.Credentials) (cpi.Credentials, error) {
	return nil, errors.ErrNotSupported("write", "credentials", SecretServerRepositoryType)
}

func (r *Repository) Read(force bool) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if !force && r.config != nil {
		return nil
	}

	configReader, err := r.getConfigFromSecretServer()
	if err != nil {
		return err
	}
	defer configReader.Close()

	config := &SecretServerConfig{}
	if err := json.NewDecoder(configReader).Decode(config); err != nil {
		return fmt.Errorf("unable to unmarshal config: %w", err)
	}
	r.config = config

	return nil
}

func (r *Repository) getConfigFromSecretServer() (io.ReadCloser, error) {
	u, err := url.Parse(r.url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %q: %w", r.url, err)
	}
	u.Path = filepath.Join(u.Path, r.configName)

	res, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("unable to get config: %w", err)
	}
	reader := res.Body

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

		return ioutil.NopCloser(bytes.NewBuffer(dst)), nil
	case Plaintext:
		return reader, nil
	default:
		return nil, errors.ErrNotImplemented(string(r.cipher), SecretServerRepositoryType)
	}
}

// ecbDecrypt decrypts ecb data
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

func newCredentialsFromContainerRegistryCredentials(auth *ContainerRegistryCredentials) cpi.Credentials {
	props := common.Properties{
		cpi.ATTR_USERNAME: auth.Username,
		cpi.ATTR_PASSWORD: auth.Password,
	}
	props.SetNonEmptyValue(cpi.ATTR_SERVER_ADDRESS, auth.Host)
	return cpi.NewCredentials(props)
}
