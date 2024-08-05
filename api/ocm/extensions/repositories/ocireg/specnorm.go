package ocireg

import (
	"strings"

	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
)

func init() {
	genericocireg.RegisterSpecificationNormalizer(ocireg.Type, Normalize)
	genericocireg.RegisterSpecificationNormalizer(ocireg.TypeV1, Normalize)
}

func Normalize(s *genericocireg.RepositorySpec) {
	if os, ok := s.RepositorySpec.(*ocireg.RepositorySpec); ok {
		if s.SubPath == "" {
			scheme := ""
			baseURL := os.BaseURL
			if idx := strings.Index(baseURL, "://"); idx > 0 {
				scheme = baseURL[:idx+3]
				baseURL = baseURL[idx+3:]
			}
			if idx := strings.Index(baseURL, "/"); idx > 0 {
				s.SubPath = baseURL[idx+1:]
				os.BaseURL = scheme + baseURL[:idx]
			}
		}
	}
}
