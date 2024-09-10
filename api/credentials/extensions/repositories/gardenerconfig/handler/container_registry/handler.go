package container_registry

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ocm.software/ocm/api/credentials/cpi"
	gardenercfgcpi "ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
)

func init() {
	gardenercfgcpi.RegisterHandler(Handler{})
}

// config is the struct that describes the gardener config data structure.
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

func (h Handler) ConfigType() gardenercfgcpi.ConfigType {
	return gardenercfgcpi.ContainerRegistry
}

func (h Handler) ParseConfig(configReader io.Reader) ([]gardenercfgcpi.Credential, error) {
	config := &config{}
	if err := json.NewDecoder(configReader).Decode(&config); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}

	creds := []gardenercfgcpi.Credential{}
	for credentialName, credential := range config.ContainerRegistry {
		var (
			scheme string
			port   string
		)
		if credential.Host != "" {
			parsedHost, err := utils.ParseURL(credential.Host)
			if err != nil {
				return nil, fmt.Errorf("unable to parse host: %w", err)
			}
			scheme = parsedHost.Scheme
			port = parsedHost.Port()
		}

		for _, imgRef := range credential.ImageReferencePrefixes {
			parsedImgPrefix, err := utils.ParseURL(imgRef)
			if err != nil {
				return nil, fmt.Errorf("unable to parse image prefix: %w", err)
			}
			if parsedImgPrefix.Host == "index.docker.io" {
				parsedImgPrefix.Host = "docker.io"
			}

			consumerIdentity := cpi.ConsumerIdentity{
				cpi.ID_TYPE:            identity.CONSUMER_TYPE,
				hostpath.ID_HOSTNAME:   parsedImgPrefix.Host,
				hostpath.ID_PATHPREFIX: strings.Trim(parsedImgPrefix.Path, "/"),
			}
			consumerIdentity.SetNonEmptyValue(hostpath.ID_SCHEME, scheme)
			consumerIdentity.SetNonEmptyValue(hostpath.ID_PORT, port)

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

func newCredentialsFromContainerRegistryCredentials(auth *containerRegistryCredentials) cpi.Credentials {
	props := common.Properties{
		cpi.ATTR_USERNAME: auth.Username,
		cpi.ATTR_PASSWORD: auth.Password,
	}
	props.SetNonEmptyValue(cpi.ATTR_SERVER_ADDRESS, auth.Host)
	return cpi.NewCredentials(props)
}
