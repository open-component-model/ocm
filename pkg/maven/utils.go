package maven

import (
	"crypto"
	"slices"
	"strings"
)

// HashExt returns the 'maven' hash extension for the given hash.
// Maven usually uses sha1, sha256, sha512, md5 instead of SHA-1, SHA-256, SHA-512, MD5.
func HashExt(h crypto.Hash) string {
	return strings.ReplaceAll(strings.ToLower(h.String()), "-", "")
}

var hashes = [5]crypto.Hash{crypto.SHA512, crypto.SHA256, crypto.SHA1, crypto.MD5}

// bestAvailableHashForFile returns the best available hash for the given file.
// It first checks for SHA-512, then SHA-256, SHA-1, and finally MD5. If nothing is found, it returns 0.
func bestAvailableHashForFile(list []string, filename string) crypto.Hash {
	for _, hash := range hashes {
		if slices.Contains(list, filename+"."+HashExt(hash)) {
			return hash
		}
	}
	return 0
}

// bestAvailableHashForFile returns the best available hash for the given file.
// It first checks for SHA-512, then SHA-256, SHA-1, and finally MD5. If nothing is found, it returns 0.
func bestAvailableHash(list []crypto.Hash) crypto.Hash {
	for _, hash := range hashes {
		if slices.Contains(list, hash) {
			return hash
		}
	}
	return 0
}
