package mimeutils

import (
	"mime"

	"github.com/open-component-model/ocm/pkg/logging"
	ocmmime "github.com/open-component-model/ocm/pkg/mime"
)

var ocmTypes = map[string]string{
	// added entries
	".txt":    ocmmime.MIME_TEXT,
	".yaml":   ocmmime.MIME_YAML_OFFICIAL,
	".gzip":   ocmmime.MIME_GZIP,
	".tar":    ocmmime.MIME_TAR,
	".tgz":    ocmmime.MIME_TGZ,
	".tar.gz": ocmmime.MIME_TGZ,
	".pom":    ocmmime.MIME_XML,
	".zip":    ocmmime.MIME_GZIP,
	".jar":    ocmmime.MIME_JAR,
	".module": ocmmime.MIME_JSON, // gradle module metadata
}

func init() {
	for k, v := range ocmTypes {
		err := mime.AddExtensionType(k, v)
		if err != nil {
			logging.DynamicLogger(logging.DefineSubRealm("mimeutils")).Error("failed to add extension type", "extension", k, "type", v, "error", err)
		}
	}
}

// Deprecated: use mime.TypeByExtension instead.
func TypeByExtension(ext string) string {
	return mime.TypeByExtension(ext)
}
