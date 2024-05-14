package mime

import (
	"mime"

	"github.com/open-component-model/ocm/pkg/logging"
)

const (
	MIME_TEXT  = "text/plain"
	MIME_OCTET = "application/octet-stream"

	MIME_JSON          = "application/x-json"
	MIME_JSON_ALT      = "text/json" // no utf8
	MIME_JSON_OFFICIAL = "application/json"
	MIME_XML           = "application/xml"
	MIME_YAML          = "application/x-yaml"
	MIME_YAML_ALT      = "text/yaml" // no utf8
	MIME_YAML_OFFICIAL = "application/yaml"

	MIME_GZIP    = "application/gzip"
	MIME_TAR     = "application/x-tar"
	MIME_TGZ     = "application/x-tgz"
	MIME_TGZ_ALT = MIME_TAR + "+gzip"

	MIME_JAR = "application/x-jar"
)

func init() {
	ocmTypes := map[string]string{
		// added entries
		".txt":    MIME_TEXT,
		".yaml":   MIME_YAML_OFFICIAL,
		".gzip":   MIME_GZIP,
		".tar":    MIME_TAR,
		".tgz":    MIME_TGZ,
		".tar.gz": MIME_TGZ,
		".pom":    MIME_XML,
		".zip":    MIME_GZIP,
		".jar":    MIME_JAR,
		".module": MIME_JSON, // gradle module metadata
	}

	for k, v := range ocmTypes {
		err := mime.AddExtensionType(k, v)
		if err != nil {
			logging.DynamicLogger(logging.DefineSubRealm("mimeutils")).Error("failed to add extension type", "extension", k, "type", v, "error", err)
		}
	}
}
