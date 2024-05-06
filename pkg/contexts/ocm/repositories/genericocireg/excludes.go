package genericocireg

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/docker"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/empty"
)

var Excludes = []string{
	docker.Type,
	artifactset.Type,
	empty.Type,
}
