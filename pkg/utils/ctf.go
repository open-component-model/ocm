package utils

import (
	"fmt"
	"strings"
)

// CTFComponentArchiveFilename returns the name of the component archive file in the ctf.
func CTFComponentArchiveFilename(name, version string) string {
	return fmt.Sprintf("%s-%s.tar", strings.ReplaceAll(name, "/", "_"), version)
}
