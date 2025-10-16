package npm

import (
	"bufio"
	"os"
	"strings"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/utils"
)

type npmConfig map[string]string

// readNpmConfigFile reads "~/.npmrc" file line by line, parse it and return the result as a npmConfig.
func readNpmConfigFile(path string) (npmConfig, string, error) {
	path, err := utils.ResolvePath(path)
	if err != nil {
		return nil, path, errors.Wrapf(err, "cannot resolve path %q", path)
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, path, err
	}
	defer file.Close()

	// Create a new scanner and read the file line by line
	scanner := bufio.NewScanner(file)
	cfg := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		line, authFound := strings.CutPrefix(line, "//")
		if !authFound {
			// e.g. 'global=false'
			continue
		}
		// Split the line into key and value
		parts := strings.SplitN(line, ":_authToken=", 2)
		if len(parts) == 2 {
			if strings.HasSuffix(parts[0], "/") {
				cfg[parts[0][:len(parts[0])-1]] = parts[1]
			} else {
				cfg[parts[0]] = parts[1]
			}
		}
	}

	// Check for errors
	if err = scanner.Err(); err != nil {
		return nil, path, err
	}

	return cfg, path, nil
}
