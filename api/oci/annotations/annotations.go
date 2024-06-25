package annotations

// MAINARTIFACT_ANNOTATION is the name of the OCI manifest annotation used to describe
// the main artifact identity in an artifact set.
const MAINARTIFACT_ANNOTATION = "software.ocm/main"

// TAGS_ANNOTATION is the name of the OCI manifest annotation used to describe a set of
// tags assigned to a manifest in an artifact set.
const TAGS_ANNOTATION = "software.ocm/tags"

// TYPE_ANNOTATION is the name of the OCI manifest annotation used to describe the type
// (not yet used).
const TYPE_ANNOTATION = "software.ocm/type"

// OCITAG_ANNOTATION is the name of the OCI manifest annotation used to describe a tag.
const OCITAG_ANNOTATION = "org.opencontainers.image.ref.name"

// COMPVERS_ANNOTATION is the name of the OCI manifest annotation used to describe
// the OCM identity of the origin of an OCI artifact. This is the identity of a
// component version `<component name>:<component version>`.
const COMPVERS_ANNOTATION = "software.ocm/component-version"
