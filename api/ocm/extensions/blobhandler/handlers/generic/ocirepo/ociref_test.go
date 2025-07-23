package ocirepo_test

import (
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/generic/ocirepo"
	"ocm.software/ocm/api/utils/runtime"
	"testing"
)

// TestGenOciRef tests the GenOciRef function, which is used in the context of generic oci blob handler,
// in particular for the StoreBlob() operations. The function GenOciRef was introduced to align the way how
// generic/ocirepo.artifactHandler generates OCI references with the way how oci/ocirepo.artifactHandler does it.
func TestGenOciRef(t *testing.T) {
	tests := []struct {
		name, host, port, tag, version, namespace string
		wantRef                                   string
		wantErr                                   bool
	}{
		// TODO(ikhandamirov): Uncomment this valid test case once the oci.ParseArt() function accepts port numbers.
		//{
		//	"all fields set",
		//	"sub.example.com",
		//	"433",
		//	":latest",
		//	"@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		//	"repo/publisher/image",
		//	"sub.example.com:443/repo/publisher/image:latest@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		//	false,
		//},
		{
			"port not set",
			"sub.example.com",
			"",
			":latest",
			"@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"repo/publisher/image",
			"sub.example.com/repo/publisher/image:latest@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			false,
		},
		{
			"host not set",
			"",
			"",
			":latest",
			"@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"repo/publisher/image",
			"",
			true,
		},
		{
			"namespace not set",
			"sub.example.com",
			"",
			":latest",
			"@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"",
			"",
			true,
		},
		{
			"tag not set",
			"sub.example.com",
			"",
			"",
			"@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"repo/publisher/image",
			"sub.example.com/repo/publisher/image@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			false,
		},
		{
			"version not set",
			"sub.example.com",
			"",
			":latest",
			"",
			"repo/publisher/image",
			"sub.example.com/repo/publisher/image:latest",
			false,
		},
		{
			"tag and version not set",
			"sub.example.com",
			"",
			"",
			"",
			"repo/publisher/image",
			"sub.example.com/repo/publisher/image",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ocirepo.GenOciRef(tt.host, tt.port, tt.tag, tt.version, tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenOciRef() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ref != tt.wantRef {
				t.Errorf("GenOciRef() = %v, want %v", ref, tt.wantRef)
			}
		})
	}
}

// TestGetInfo tests the ociuploadattr.Attribute.GetInfo() function, which is in particular used in the context of
// generic oci blob handler for the StoreBlob() operations. The test validates that the uploader configuration is
// properly parsed and translated into required objects.
func TestGetInfo(t *testing.T) {
	tests := []struct {
		name                                                                          string
		spec                                                                          string
		wantRepoBaseURL                                                               string
		wantBaseType, wantBaseScheme, wantBaseHost, wantBasePort, wantNamespacePrefix string
	}{
		// This test corresponds to the following uploader configuration:
		// --uploader 'ocm/ociArtifacts={"repository":{"baseUrl":"sub.example.com","type":"OCIRegistry"}}'
		{
			"OCI registry with base URL",
			`{"repository":{"baseUrl":"sub.example.com","type":"OCIRegistry"}}`,
			"sub.example.com",
			"OCIRegistry",
			"https",
			"sub.example.com",
			"",
			"",
		},

		// This test corresponds to the following uploader configuration:
		// --uploader 'ocm/ociArtifacts={"repository":{"baseUrl":"sub.example.com:443","type":"OCIRegistry"},"namespacePrefix":"repo"}'
		{
			"OCI registry with base URL, port and namespace prefix",
			`{"repository":{"baseUrl":"sub.example.com:443","type":"OCIRegistry"},"namespacePrefix":"repo"}`,
			"sub.example.com:443",
			"OCIRegistry",
			"https",
			"sub.example.com",
			"443",
			"repo",
		},
	}

	ctx := ocm.DefaultContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attri, err := ociuploadattr.AttributeType{}.Decode([]byte(tt.spec), runtime.DefaultJSONEncoding)
			if err != nil {
				t.Errorf("failed to unmarshal spec: %v", err)
				return
			}

			attr, ok := attri.(*ociuploadattr.Attribute)
			if !ok {
				t.Errorf("decoded interdace is not of type *ociuploadattr.Attribute")
				return
			}

			repo, base, prefix, err := attr.GetInfo(ctx)
			if err != nil {
				t.Error(err)
				return
			}

			repoSpec, ok := repo.GetSpecification().(*ocireg.RepositorySpec)
			if !ok {
				t.Errorf("repository specification is not of type *ocireg.RepositorySpec")
			}
			if repoSpec.BaseURL != tt.wantRepoBaseURL {
				t.Errorf("GetInfo() repo base URL = %v, want %v", repoSpec.BaseURL, tt.wantRepoBaseURL)
			}

			if base.Type != tt.wantBaseType {
				t.Errorf("GetInfo() base type = %v, want %v", base.Type, tt.wantBaseType)
			}

			if base.Scheme != tt.wantBaseScheme {
				t.Errorf("GetInfo() base scheme = %v, want %v", base.Scheme, tt.wantBaseScheme)
			}

			host, port := base.HostPort()
			if host != tt.wantBaseHost {
				t.Errorf("GetInfo() base host = %v, want %v", host, tt.wantBaseHost)
			}
			if port != tt.wantBasePort {
				t.Errorf("GetInfo() base port = %v, want %v", port, tt.wantBasePort)
			}

			if prefix != tt.wantNamespacePrefix {
				t.Errorf("GetInfo() prefix = %v, want %v", prefix, tt.wantNamespacePrefix)
			}
		})
	}
}
