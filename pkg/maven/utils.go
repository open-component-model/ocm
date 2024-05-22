// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maven

import (
	"crypto"
	. "github.com/mandelsoft/goutils/regexutils"
	"slices"
	"strings"
)

// HashUrlExt returns the 'maven' hash extension for the given hash.
// Maven usually uses sha1, sha256, sha512, md5 instead of SHA-1, SHA-256, SHA-512, MD5.
func HashUrlExt(h crypto.Hash) string {
	return "." + strings.ReplaceAll(strings.ToLower(h.String()), "-", "")
}

func FilterByCoordinates(fileMap map[string]crypto.Hash, coords *Coordinates) map[string]crypto.Hash {
	exp := Literal(coords.ArtifactId + "-" + coords.Version)
	if coords.Classifier != "" {
		exp = Sequence(exp, Literal("-"+coords.Classifier))
	}
	if coords.Extension != "" {
		exp = Sequence(exp, Literal("."+coords.Extension))
	} else {
		exp = Sequence(exp, Literal("."), Match(".*"))
	}
	exp = Anchored(exp)
	for file := range fileMap {
		if !exp.MatchString(file) {
			delete(fileMap, file)
		}
	}
	return fileMap
}

var hashes = [5]crypto.Hash{crypto.SHA512, crypto.SHA256, crypto.SHA1, crypto.MD5}

// bestAvailableHash returns the best available hash for the given file.
// It first checks for SHA-512, then SHA-256, SHA-1, and finally MD5. If nothing is found, it returns 0.
func bestAvailableHash(list []string, filename string) crypto.Hash {
	for _, hash := range hashes {
		if slices.Contains(list, filename+HashUrlExt(hash)) {
			return hash
		}
	}
	return 0
}
