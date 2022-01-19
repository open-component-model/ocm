// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ociclient_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/gardener/ocm/pkg/oci/client"
	"github.com/gardener/ocm/pkg/oci/client/credentials"
	"github.com/gardener/ocm/pkg/oci/client/oci"
)

var _ = Describe("client", func() {

	Context("Client", func() {

		It("should push and pull an oci artifact", func() {
			ctx := context.Background()
			defer ctx.Done()

			ref := testenv.Addr + "/test/artifact:v0.0.1"
			manifest := uploadTestManifest(ctx, ref)

			res, err := client.GetManifest(ctx, ref)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Config).To(Equal(manifest.Config))
			Expect(res.Layers).To(Equal(manifest.Layers))

			compareManifestToTestManifest(ctx, ref, res)
		}, 20)

		It("should push and pull an oci image index", func() {
			ctx := context.Background()
			defer ctx.Done()

			indexRef := testenv.Addr + "/image-index/1/img:v0.0.1"
			index := uploadTestIndex(ctx, indexRef)

			actualArtifact, err := client.GetOCIArtifact(ctx, indexRef)
			Expect(err).ToNot(HaveOccurred())

			Expect(actualArtifact.IsManifest()).To(BeFalse())
			Expect(actualArtifact.IsIndex()).To(BeTrue())
			compareImageIndices(actualArtifact.GetIndex(), index)
		}, 20)

		It("should push and pull an empty oci image index", func() {
			ctx := context.Background()
			defer ctx.Done()

			ref := testenv.Addr + "/image-index/2/empty-img:v0.0.1"
			index := oci.Index{
				Manifests: []*oci.Manifest{},
				Annotations: map[string]string{
					"test": "test",
				},
			}

			tmp, err := oci.NewIndexArtifact(&index)
			Expect(err).ToNot(HaveOccurred())

			err = client.PushOCIArtifact(ctx, ref, tmp)
			Expect(err).ToNot(HaveOccurred())

			actualArtifact, err := client.GetOCIArtifact(ctx, ref)
			Expect(err).ToNot(HaveOccurred())

			Expect(actualArtifact.IsManifest()).To(BeFalse())
			Expect(actualArtifact.IsIndex()).To(BeTrue())
			compareImageIndices(actualArtifact.GetIndex(), &index)
		}, 20)

		It("should push and pull an oci image index with only 1 manifest and no platform information", func() {
			ctx := context.Background()
			defer ctx.Done()

			ref := testenv.Addr + "/image-index/3/img:v0.0.1"
			manifest1Ref := testenv.Addr + "/image-index/1/img-platform-1:v0.0.1"
			manifest := uploadTestManifest(ctx, manifest1Ref)
			index := oci.Index{
				Manifests: []*oci.Manifest{
					{
						Data: manifest,
					},
				},
				Annotations: map[string]string{
					"test": "test",
				},
			}

			tmp, err := oci.NewIndexArtifact(&index)
			Expect(err).ToNot(HaveOccurred())

			err = client.PushOCIArtifact(ctx, ref, tmp)
			Expect(err).ToNot(HaveOccurred())

			actualArtifact, err := client.GetOCIArtifact(ctx, ref)
			Expect(err).ToNot(HaveOccurred())

			Expect(actualArtifact.IsManifest()).To(BeFalse())
			Expect(actualArtifact.IsIndex()).To(BeTrue())
			compareImageIndices(actualArtifact.GetIndex(), &index)
		}, 20)

		It("should copy an oci artifact", func() {
			ctx := context.Background()
			defer ctx.Done()

			ref := testenv.Addr + "/test/artifact:v0.0.1"
			manifest := uploadTestManifest(ctx, ref)

			newRef := testenv.Addr + "/new/artifact:v0.0.1"
			Expect(ociclient.Copy(ctx, client, ref, newRef)).To(Succeed())

			res, err := client.GetManifest(ctx, newRef)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Config).To(Equal(manifest.Config))
			Expect(res.Layers).To(Equal(manifest.Layers))

			var configBlob bytes.Buffer
			Expect(client.Fetch(ctx, ref, res.Config, &configBlob)).To(Succeed())
			Expect(configBlob.String()).To(Equal("test"))

			var layerBlob bytes.Buffer
			Expect(client.Fetch(ctx, ref, res.Layers[0], &layerBlob)).To(Succeed())
			Expect(layerBlob.String()).To(Equal("test-config"))
		}, 20)

		It("should copy an oci image index", func() {
			ctx := context.Background()
			defer ctx.Done()

			ref := testenv.Addr + "/copy/image-index/src/img:v0.0.1"
			index := uploadTestIndex(ctx, ref)

			newRef := testenv.Addr + "/copy/image-index/tgt/img:v0.0.1"
			Expect(ociclient.Copy(ctx, client, ref, newRef)).To(Succeed())

			actualArtifact, err := client.GetOCIArtifact(ctx, newRef)
			Expect(err).ToNot(HaveOccurred())

			Expect(actualArtifact.IsManifest()).To(BeFalse())
			Expect(actualArtifact.IsIndex()).To(BeTrue())
			compareImageIndices(actualArtifact.GetIndex(), index)

			for _, manifest := range actualArtifact.GetIndex().Manifests {
				compareManifestToTestManifest(ctx, newRef, manifest.Data)
			}
		}, 20)

	})

	Context("ExtendedClient", func() {
		Context("ListTags", func() {

			var (
				server  *httptest.Server
				host    string
				handler func(http.ResponseWriter, *http.Request)
				makeRef = func(repo string) string {
					return fmt.Sprintf("%s/%s", host, repo)
				}
			)

			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					handler(writer, request)
				}))

				hostUrl, err := url.Parse(server.URL)
				Expect(err).ToNot(HaveOccurred())
				host = hostUrl.Host
			})

			AfterEach(func() {
				server.Close()
			})

			It("should return a list of tags", func() {
				var (
					ctx        = context.Background()
					repository = "myproject/repo/myimage"
				)
				defer ctx.Done()
				handler = func(w http.ResponseWriter, req *http.Request) {
					if req.URL.Path == "/v2/" {
						// first auth discovery call by the library
						w.WriteHeader(200)
						return
					}
					Expect(req.URL.String()).To(Equal("/v2/myproject/repo/myimage/tags/list?n=1000"))
					w.WriteHeader(200)
					_, _ = w.Write([]byte(`
{
  "tags": [ "0.0.1", "0.0.2" ]
}
`))
				}

				client, err := ociclient.NewClient(logr.Discard(),
					ociclient.AllowPlainHttp(true),
					ociclient.WithKeyring(credentials.New()))
				Expect(err).ToNot(HaveOccurred())
				tags, err := client.ListTags(ctx, makeRef(repository))
				Expect(err).ToNot(HaveOccurred())
				Expect(tags).To(ConsistOf("0.0.1", "0.0.2"))
			})

		})

		Context("ListRepositories", func() {

			var (
				server  *httptest.Server
				host    string
				handler func(http.ResponseWriter, *http.Request)
				makeRef = func(repo string) string {
					return fmt.Sprintf("%s/%s", host, repo)
				}
			)

			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					handler(writer, request)
				}))

				hostUrl, err := url.Parse(server.URL)
				Expect(err).ToNot(HaveOccurred())
				host = hostUrl.Host
			})

			AfterEach(func() {
				server.Close()
			})

			It("should return a list of repositories", func() {
				var (
					ctx        = context.Background()
					repository = "myproject/repo"
				)
				defer ctx.Done()
				handler = func(w http.ResponseWriter, req *http.Request) {
					if req.URL.Path == "/v2/" {
						// first auth discovery call by the library
						w.WriteHeader(200)
						return
					}
					Expect(req.URL.String()).To(Equal("/v2/_catalog?n=1000"))
					w.WriteHeader(200)
					_, _ = w.Write([]byte(`
{
  "repositories": [ "myproject/repo/image1", "myproject/repo/image2" ]
}
`))
				}

				client, err := ociclient.NewClient(logr.Discard(),
					ociclient.AllowPlainHttp(true),
					ociclient.WithKeyring(credentials.New()))
				Expect(err).ToNot(HaveOccurred())
				repos, err := client.ListRepositories(ctx, makeRef(repository))
				Expect(err).ToNot(HaveOccurred())
				Expect(repos).To(ConsistOf(makeRef("myproject/repo/image1"), makeRef("myproject/repo/image2")))
			})

		})
	})

})

func uploadTestManifest(ctx context.Context, ref string) *ocispecv1.Manifest {
	data := []byte("test")
	layerData := []byte("test-config")
	manifest := &ocispecv1.Manifest{
		Config: ocispecv1.Descriptor{
			MediaType: "text/plain",
			Digest:    digest.FromBytes(data),
			Size:      int64(len(data)),
		},
		Layers: []ocispecv1.Descriptor{
			{
				MediaType: "text/plain",
				Digest:    digest.FromBytes(layerData),
				Size:      int64(len(layerData)),
			},
		},
	}
	store := ociclient.GenericStore(func(ctx context.Context, desc ocispecv1.Descriptor, writer io.Writer) error {
		switch desc.Digest.String() {
		case manifest.Config.Digest.String():
			_, err := writer.Write(data)
			return err
		default:
			_, err := writer.Write(layerData)
			return err
		}
	})
	Expect(client.PushManifest(ctx, ref, manifest, ociclient.WithStore(store))).To(Succeed())
	return manifest
}

func compareManifestToTestManifest(ctx context.Context, ref string, manifest *ocispecv1.Manifest) {
	var configBlob bytes.Buffer
	Expect(client.Fetch(ctx, ref, manifest.Config, &configBlob)).To(Succeed())
	Expect(configBlob.String()).To(Equal("test"))

	var layerBlob bytes.Buffer
	Expect(client.Fetch(ctx, ref, manifest.Layers[0], &layerBlob)).To(Succeed())
	Expect(layerBlob.String()).To(Equal("test-config"))
}

func uploadTestIndex(ctx context.Context, indexRef string) *oci.Index {
	splitted := strings.Split(indexRef, ":")
	indexRepo := strings.Join(splitted[0:len(splitted)-1], ":")
	tag := splitted[len(splitted)-1]

	manifest1Ref := fmt.Sprintf("%s-platform-1:%s", indexRepo, tag)
	manifest2Ref := fmt.Sprintf("%s-platform-2:%s", indexRepo, tag)
	manifest1 := uploadTestManifest(ctx, manifest1Ref)
	manifest2 := uploadTestManifest(ctx, manifest2Ref)
	index := oci.Index{
		Manifests: []*oci.Manifest{
			{
				Descriptor: ocispecv1.Descriptor{
					Platform: &ocispecv1.Platform{
						Architecture: "amd64",
						OS:           "linux",
					},
				},
				Data: manifest1,
			},
			{
				Descriptor: ocispecv1.Descriptor{
					Platform: &ocispecv1.Platform{
						Architecture: "amd64",
						OS:           "windows",
					},
				},
				Data: manifest2,
			},
		},
		Annotations: map[string]string{
			"test": "test",
		},
	}

	tmp, err := oci.NewIndexArtifact(&index)
	Expect(err).ToNot(HaveOccurred())

	Expect(client.PushOCIArtifact(ctx, indexRef, tmp)).To(Succeed())
	return &index
}

func compareImageIndices(actualIndex *oci.Index, expectedIndex *oci.Index) {
	Expect(actualIndex.Annotations).To(Equal(expectedIndex.Annotations))
	Expect(len(actualIndex.Manifests)).To(Equal(len(expectedIndex.Manifests)))

	for i := 0; i < len(actualIndex.Manifests); i++ {
		actualManifest := actualIndex.Manifests[i]
		expectedManifest := expectedIndex.Manifests[i]

		expectedManifestBytes, err := json.Marshal(expectedManifest.Data)
		Expect(err).ToNot(HaveOccurred())

		Expect(actualManifest.Descriptor.MediaType).To(Equal(ocispecv1.MediaTypeImageManifest))
		Expect(actualManifest.Descriptor.Digest).To(Equal(digest.FromBytes(expectedManifestBytes)))
		Expect(actualManifest.Descriptor.Size).To(Equal(int64(len(expectedManifestBytes))))
		Expect(actualManifest.Descriptor.Platform).To(Equal(expectedManifest.Descriptor.Platform))
		Expect(actualManifest.Data).To(Equal(expectedManifest.Data))
	}
}
