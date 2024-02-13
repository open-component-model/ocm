package npm

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	// REPOSITORY_TYPE is the type of the NPMConfig.
	REPOSITORY_TYPE    = "NPMConfig"
	REPOSITORY_TYPE_v1 = REPOSITORY_TYPE + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](REPOSITORY_TYPE))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](REPOSITORY_TYPE_v1, cpi.WithDescription(usage), cpi.WithFormatSpec(format)))
}

var usage = `
This repository type can be used to access credentials stored in a file
following the NPM config format (~/.npmrc). It take into account the
credentials helper section, also. If enabled, the described
credentials will be automatically assigned to appropriate consumer ids.
`

var format = `The repository specification supports the following fields:
` + listformat.FormatListElements("", listformat.StringElementDescriptionList{
	"npmrcFile", "*string*: the file path to a NPM config file",
	"propagateConsumerIdentity", "*bool*(optional): enable consumer id propagation",
})

// RepositorySpec describes a docker config based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	NpmrcFile                   string `json:"npmrcFile,omitempty"`
}

// NewRepositorySpec creates a new memory RepositorySpec.
func NewRepositorySpec(path string, prop ...bool) *RepositorySpec {
	p := false
	for _, e := range prop {
		p = p || e
	}
	if path == "" {
		path = "~/.npmrc"
	}
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(REPOSITORY_TYPE),
		NpmrcFile:           path,
	}
}

func (rs *RepositorySpec) GetType() string {
	return REPOSITORY_TYPE
}

func (rs *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	//	r := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories)
	//	repos, ok := r.(*Repositories)
	//	if !ok {
	//		return nil, fmt.Errorf("failed to assert type %T to Repositories", r)
	//	}
	//	return repos.GetRepository(ctx, a.DockerConfigFile, a.DockerConfig, true)
	return nil, fmt.Errorf("not implemented")
}

// ReadNpmConfigFile reads "~/.npmrc" file line by line, parse it and return the result as a map.
func ReadNpmConfigFile(path string) (map[string]string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new scanner and read the file line by line
	scanner := bufio.NewScanner(file)
	config := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		line, authFound := strings.CutPrefix(line, "//")
		if !authFound {
			// e.g. 'global=false'
			continue
		}
		// Split the line into key and value
		parts := strings.SplitN(line, "/:_authToken=", 2)
		if len(parts) == 2 {
			// TODO: should we directly add here the prefix 'https://'?
			config[parts[0]] = parts[1]
		}
	}

	// Check for errors
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}
