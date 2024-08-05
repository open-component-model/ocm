package genericocireg

import (
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/oci/extensions/repositories/docker"
	"ocm.software/ocm/api/oci/extensions/repositories/empty"
)

var Excludes = []string{
	docker.Type,
	artifactset.Type,
	empty.Type,
}
