package gardenerconfig

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
	"strings"
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/errors"
)

var log = false

type Cipher string
type ConfigType string

const (
	Plaintext Cipher = "PLAINTEXT"
	AESECB    Cipher = "AES.ECB"

	ContainerRegistry ConfigType = "container_registry"
)

var ConsumerTypes = map[ConfigType]string{
	ContainerRegistry: identity.CONSUMER_TYPE,
}

type Repository struct {
	ctx        cpi.Context
	lock       sync.RWMutex
	url        string
	configType ConfigType
	cipher     Cipher
	key        []byte
	propagate  bool
	creds      map[string]cpi.Credentials
	fs         vfs.FileSystem
}

// Config is the struct that describes the cc-config data structure
type Config struct {
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

func NewRepository(ctx cpi.Context, url string, configType ConfigType, cipher Cipher, key []byte, propagate bool) *Repository {
	return &Repository{
		ctx:        ctx,
		url:        url,
		configType: configType,
		cipher:     cipher,
		key:        key,
		propagate:  propagate,
		fs:         vfsattr.Get(ctx),
	}
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

	configReader, err := r.getConfig()
	if err != nil {
		return err
	}
	defer configReader.Close()

	config := &Config{}
	if err := json.NewDecoder(configReader).Decode(config); err != nil {
		return fmt.Errorf("unable to unmarshal config: %w", err)
	}

	r.creds = map[string]cpi.Credentials{}

	for credentialName, credential := range config.ContainerRegistry {
		credentialName := credentialName
		if _, ok := r.creds[credentialName]; ok {
			return errors.Newf("credential with name %s already exist", credentialName)
		}

		scheme, port := "", ""
		if credential.Host != "" {
			if !strings.Contains(credential.Host, "://") {
				credential.Host = "dummy://" + credential.Host
			}

			parsedHost, err := url.Parse(credential.Host)
			if err != nil {
				return fmt.Errorf("unable to parse host: %w", err)
			}
			scheme = parsedHost.Scheme
			port = parsedHost.Port()
		}

		for _, imgPrefix := range credential.ImageReferencePrefixes {
			parsedImgPrefix, err := ParseImageRef(imgPrefix)
			if err != nil {
				return fmt.Errorf("unable to parse image prefix: %w", err)
			}

			id := cpi.ConsumerIdentity{
				cpi.ATTR_TYPE:          ConsumerTypes[r.configType],
				identity.ID_SCHEME:     scheme,
				identity.ID_HOSTNAME:   parsedImgPrefix.Host,
				identity.ID_PORT:       port,
				identity.ID_PATHPREFIX: parsedImgPrefix.Path,
			}

			var creds cpi.Credentials
			if log {
				fmt.Printf("propagate id %q\n", id)
			}
			creds = newCredentialsFromContainerRegistryCredentials(credential)
			r.creds[credentialName] = creds

			if r.propagate {
				getCredentials := func() (credentials.Credentials, error) {
					return r.LookupCredentials(credentialName)
				}

				cg := CredentialGetter{
					getCredentials: getCredentials,
				}

				r.ctx.SetCredentialsForConsumer(id, cg)
			}
		}
	}

	return nil
}

func ParseImageRef(imgRef string) (*url.URL, error) {
	if !strings.Contains(imgRef, "://") {
		imgRef = "dummy://" + imgRef
	}

	parsedImgRef, err := url.Parse(imgRef)
	if err != nil {
		return nil, err
	}

	if parsedImgRef.Host == "index.docker.io" {
		parsedImgRef.Host = "docker.io"
	}

	return parsedImgRef, nil
}

func (r *Repository) getConfig() (io.ReadCloser, error) {
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

		return ioutil.NopCloser(bytes.NewBuffer(dst)), nil
	case Plaintext:
		return reader, nil
	default:
		return nil, errors.ErrNotImplemented(string(r.cipher), RepositoryType)
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
