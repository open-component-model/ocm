package cc_config

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

	dockercred "github.com/docker/cli/cli/config/credentials"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/errors"
)

var log = false

type Cipher string

const (
	Plaintext Cipher = "PLAINTEXT"
	AESECB    Cipher = "AES.ECB"
)

type Repository struct {
	ctx          cpi.Context
	lock         sync.RWMutex
	url          string
	consumerType string
	cipher       Cipher
	key          []byte
	propagate    bool
	index        *IndexNode
	creds        map[string]cpi.Credentials
	fs           vfs.FileSystem
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

func NewRepository(ctx cpi.Context, url string, consumerType string, cipher Cipher, key []byte, propagate bool, fs vfs.FileSystem) *Repository {
	return &Repository{
		ctx:          ctx,
		url:          url,
		consumerType: consumerType,
		cipher:       cipher,
		key:          key,
		propagate:    propagate,
		fs:           fs,
	}
}

var _ cpi.Repository = &Repository{}

func (r *Repository) ExistsCredentials(name string) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.index.Find(name) == "", nil
}

func (r *Repository) LookupCredentials(name string) (cpi.Credentials, error) {
	if err := r.Read(false); err != nil {
		return nil, err
	}
	r.lock.RLock()
	defer r.lock.RUnlock()

	segment := r.index.Find(name)

	auth, ok := r.creds[segment]
	if !ok {
		return nil, cpi.ErrUnknownCredentials(name)
	}

	return auth, nil
}

func (r *Repository) WriteCredentials(name string, creds cpi.Credentials) (cpi.Credentials, error) {
	return nil, errors.ErrNotSupported("write", "credentials", CCConfigRepositoryType)
}

func (r *Repository) Read(force bool) error {
	r.lock.Lock()
	defer r.lock.Unlock()
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

	r.index = NewIndexNode()
	r.creds = map[string]cpi.Credentials{}

	// TODO: what is the propagate flag for? must it be used here?
	for _, credential := range config.ContainerRegistry {
		for _, imgPrefix := range credential.ImageReferencePrefixes {
			if _, ok := r.creds[imgPrefix]; ok {
				return errors.Newf("credentials for image prefix %s already exist", imgPrefix)
			}

			// TODO: remember weird behavior if protocol prefix is missing, maybe use/implement util function?
			url, err := url.Parse(imgPrefix)
			if err != nil {
				return err
			}
			hostname := dockercred.ConvertToHostname(url.Host)
			if hostname == "index.docker.io" {
				hostname = "docker.io"
			}

			id := cpi.ConsumerIdentity{
				cpi.ATTR_TYPE:          r.consumerType,
				identity.ID_HOSTNAME:   hostname,
				identity.ID_PATHPREFIX: url.Path,
			}

			var creds cpi.Credentials
			if log {
				fmt.Printf("propagate id %q\n", id)
			}
			creds = newCredentialsFromContainerRegistryCredentials(credential)
			r.ctx.SetCredentialsForConsumer(id, creds)

			r.index.Insert(imgPrefix)
			r.creds[imgPrefix] = creds
		}
	}

	return nil
}

func (r *Repository) getConfig() (io.ReadCloser, error) {
	u, err := url.Parse(r.url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url %q: %w", r.url, err)
	}

	if u.Scheme == "file" {
		f, err := r.fs.Open(u.Path)
		if err != nil {
			return nil, fmt.Errorf("unable to open file: %w", err)
		}
		return f, nil
	} else {
		res, err := http.Get(u.String())
		if err != nil {
			return nil, fmt.Errorf("unable to get config from secret server: %w", err)
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
			return nil, errors.ErrNotImplemented(string(r.cipher), CCConfigRepositoryType)
		}
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

type IndexNode struct {
	segment  string
	children []*IndexNode
}

func NewIndexNode() *IndexNode {
	return &IndexNode{
		segment:  "",
		children: []*IndexNode{},
	}
}

func (n *IndexNode) Insert(path string) {
	splitPath := strings.Split(path, "/")
	child := n.findSegment(splitPath[0])
	if child == nil {
		child = &IndexNode{
			segment: splitPath[0],
		}
		n.children = append(n.children, child)
	}
	if len(splitPath) <= 1 {
		return
	}
	child.Insert(strings.Join(splitPath[1:], "/"))
}

func (n *IndexNode) findSegment(segment string) *IndexNode {
	for _, child := range n.children {
		if child.segment == segment {
			return child
		}
	}
	return nil
}

func (n *IndexNode) Find(path string) string {
	splitPath := strings.Split(path, "/")
	child := n.findSegment(splitPath[0])
	if child == nil {
		return n.segment
	}
	childSegment := child.Find(strings.Join(splitPath[1:], "/"))
	if n.segment == "" {
		return childSegment
	} else {
		return strings.Join([]string{n.segment, childSegment}, "/")
	}
}
