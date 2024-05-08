package annotations

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
)

// MAINARTIFACT_ANNOTATION is the name of the OCI manifest annotation used to describe
// the main artifact identity in an artifact set..
const MAINARTIFACT_ANNOTATION = artifactset.MAINARTIFACT_ANNOTATION

// TAGS_ANNOTATION is the name of the OCI manifest annotation used to describe a set of
// tags assigned to a manifest in an artifact set.
const TAGS_ANNOTATION = artifactset.TAGS_ANNOTATION

const TYPE_ANNOTATION = artifactset.TYPE_ANNOTATION

// OCITAG_ANNOTATION is the name of the OCI manifest annotation used to describe a tag.
const OCITAG_ANNOTATION = artifactset.OCITAG_ANNOTATION

// COMPVERS_ANNOTATION is the name of the OCI manifest annotation used to describe
// the OCM identity of the origin of an OCI artifact. This is the identity of a
// component version `<component name>:<component version>`.
const COMPVERS_ANNOTATION = "software.ocm/component-version"
