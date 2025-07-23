package oci

import (
	"fmt"
	"strings"

	"oras.land/oras-go/v2/registry"

	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/grammar"
	ociidentity "ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils/runtime"
)

func AsTags(tag string) []string {
	if tag != "" {
		return []string{tag}
	}
	return nil
}

func StandardOCIRef(host, repository, version string) string {
	sep := grammar.TagSeparator
	i := strings.Index(version, grammar.DigestSeparator)
	if i > 1 {
		return fmt.Sprintf("%s%s%s%s%s", host, grammar.RepositorySeparator, repository, sep, version)
	}
	if ok, _ := artdesc.IsDigest(version); ok {
		sep = grammar.DigestSeparator
		if strings.HasPrefix(version, sep) {
			sep = ""
		}
	}
	return fmt.Sprintf("%s%s%s%s%s", host, grammar.RepositorySeparator, repository, sep, version)
}

func IsIntermediate(spec RepositorySpec) bool {
	if s, ok := spec.(IntermediateRepositorySpecAspect); ok {
		return s.IsIntermediate()
	}
	return false
}

func IsUnknown(r RepositorySpec) bool {
	return runtime.IsUnknown(r)
}

func GetConsumerIdForRef(ref string) (cpi.ConsumerIdentity, error) {
	r, err := ParseRef(ref)
	if err != nil {
		return nil, err
	}
	return ociidentity.GetConsumerId(r.Host, r.Repository), nil
}

func IsValidReference(ref string) (bool, error) {
	_, err := registry.ParseReference(ref)
	return err == nil, err
}
