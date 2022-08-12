package container_registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	gardenercfg_cpi "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
)

func init() {
	gardenercfg_cpi.RegisterHandler(Handler{})
}

// config is the struct that describes the gardener config data structure
type config struct {
	ContainerRegistry map[string]*containerRegistryCredentials `json:"container_registry"`
}

// containerRegistryCredentials describes the container registry credentials struct as defined by the gardener config.
type containerRegistryCredentials struct {
	Username               string   `json:"username"`
	Password               string   `json:"password"`
	Privileges             string   `json:"privileges"`
	Host                   string   `json:"host,omitempty"`
	ImageReferencePrefixes []string `json:"image_reference_prefixes,omitempty"`
}

type Handler struct{}

func (h Handler) ConfigType() gardenercfg_cpi.ConfigType {
	return gardenercfg_cpi.ContainerRegistry
}

func (h Handler) ParseConfig(configReader io.Reader) ([]gardenercfg_cpi.Credential, error) {
	config := &config{}
	if err := json.NewDecoder(configReader).Decode(&config); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	creds := []gardenercfg_cpi.Credential{}
	for credentialName, credential := range config.ContainerRegistry {
		var (
			scheme string
			port   string
		)
		if credential.Host != "" {
			if !strings.Contains(credential.Host, "://") {
				credential.Host = "dummy://" + credential.Host
			}

			parsedHost, err := url.Parse(credential.Host)
			if err != nil {
				return nil, fmt.Errorf("unable to parse host: %w", err)
			}
			scheme = parsedHost.Scheme
			port = parsedHost.Port()
		}

		for _, imgRef := range credential.ImageReferencePrefixes {
			parsedImgPrefix, err := parseImageRef(imgRef)
			if err != nil {
				return nil, fmt.Errorf("unable to parse image prefix: %w", err)
			}

			consumerIdentity := cpi.ConsumerIdentity{
				cpi.ATTR_TYPE:          identity.CONSUMER_TYPE,
				identity.ID_HOSTNAME:   parsedImgPrefix.Host,
				identity.ID_PATHPREFIX: strings.Trim(parsedImgPrefix.Path, "/"),
			}
			consumerIdentity.SetNonEmptyValue(identity.ID_SCHEME, scheme)
			consumerIdentity.SetNonEmptyValue(identity.ID_PORT, port)

			c := credentials{
				name:             credentialName,
				consumerIdentity: consumerIdentity,
				properties:       newCredentialsFromContainerRegistryCredentials(credential),
			}

			creds = append(creds, c)
		}
	}

	return creds, nil
}

func parseImageRef(imgRef string) (*url.URL, error) {
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

func newCredentialsFromContainerRegistryCredentials(auth *containerRegistryCredentials) cpi.Credentials {
	props := common.Properties{
		cpi.ATTR_USERNAME: auth.Username,
		cpi.ATTR_PASSWORD: auth.Password,
	}
	props.SetNonEmptyValue(cpi.ATTR_SERVER_ADDRESS, auth.Host)
	return cpi.NewCredentials(props)
}
