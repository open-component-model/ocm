package genericocireg

import (
	"github.com/open-component-model/ocm/api/oci/extensions/repositories/artifactset"
	"github.com/open-component-model/ocm/api/oci/extensions/repositories/docker"
	"github.com/open-component-model/ocm/api/oci/extensions/repositories/empty"
)

var Excludes = []string{
	docker.Type,
	artifactset.Type,
	empty.Type,
}
