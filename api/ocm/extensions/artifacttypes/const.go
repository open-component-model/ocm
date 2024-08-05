package resourcetypes

const (
	KIND_ARTIFACT_TYPE = "artifact type"
	KIND_RESOURCE_TYPE = "resource type"
)

const (
	// OCI_ARTIFACT describes a generic OCI artifact following the
	//   [open containers image specification](https://github.com/opencontainers/image-spec/blob/main/spec.md).
	OCI_ARTIFACT = "ociArtifact"
	// OCI_IMAGE describes an OCIArtifact containing an image.
	OCI_IMAGE = "ociImage"
	// HELM_CHART describes a helm chart, either stored as OCI artifact or as tar
	// blob (tar media type).
	HELM_CHART = "helmChart"
	// NPM_PACKAGE describes a Node.js (npm) package.
	NPM_PACKAGE = "npmPackage"
	// MAVEN_PACKAGE describes the complete content addressed by a GAV.
	// The term Maven Package is introduced in the context of ocm since the term Maven Artifact (as used by Maven
	// itself) is quite ambiguous since it may refer either to the complete content addressed by a GAV or to a single
	// file addressed by a GAVCE.
	MAVEN_PACKAGE = "mavenPackage"
	// BLUEPRINT describes a Gardener Landscaper blueprint which is an artifact used in its installations describing
	// how to deploy a software component.
	BLUEPRINT        = "landscaper.gardener.cloud/blueprint"
	BLUEPRINT_LEGACY = "blueprint"
	// BLOB describes any anonymous untyped blob data.
	BLOB = "blob"
	// DIRECTORY_TREE describes a directory structure.
	DIRECTORY_TREE = "directoryTree"
	// FILESYSTEM describes a directory structure stored as archive (tar, tgz).
	FILESYSTEM        = DIRECTORY_TREE
	FILESYSTEM_LEGACY = "filesystem"
	// EXECUTABLE describes an OS executable.
	EXECUTABLE = "executable"
	// PLAIN_TEXT describes plain text.
	PLAIN_TEXT = "plainText"
	// OCM_PLUGIN describes an OS executable OCM plugin.
	OCM_PLUGIN = "ocmPlugin"

	// OCM_FILE describes a generic file or unspecified byte stream.
	OCM_FILE = "file"
	// OCM_YAML describes a YAML file.
	OCM_YAML = "yaml"
	// OCM_JSON describes a JSON file.
	OCM_JSON = "json"
	// OCM_XML describes a XML file.
	OCM_XML = "xml"
)
