package main_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	helmaccess "ocm.software/ocm/api/ocm/extensions/accessmethods/helm"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/tech/helm/loader"
	"ocm.software/ocm/api/utils/runtime"
	. "ocm.software/ocm/cmds/jfrogplugin/testhelper"
	"ocm.software/ocm/cmds/jfrogplugin/uploaders/helm"
)

var _ = Describe(helm.VERSIONED_NAME, func() {
	var (
		env                  *TestEnv
		server               *httptest.Server
		reindexedAfterUpload bool
		user, pass           = "foo", "bar"
		creds                = string(Must(json.Marshal(credentials.DirectCredentials{
			credentials.ATTR_USERNAME: user,
			credentials.ATTR_PASSWORD: pass,
		})))
	)

	BeforeEach(func() {
		env = Must(NewTestEnv())
		DeferCleanup(env.Cleanup)
	})

	It("Upload Validate Invalid Spec", func(ctx SpecContext) {
		env.CLI.Command().SetContext(ctx)
		err := env.Execute("upload", "validate", "--artifactType=bla", "abc", "def")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("error unmarshaling JSON")))
	})

	It("Validate Spec OK and empty", func(ctx SpecContext) {
		env.CLI.Command().SetContext(ctx)

		uploadSpec := &helm.JFrogHelmUploaderSpec{
			ObjectVersionedType: runtime.ObjectVersionedType{Type: helm.VERSIONED_NAME},
		}

		Expect(env.Execute(
			"upload", "validate", "--artifactType", artifacttypes.HELM_CHART,
			helm.NAME,
			string(Must(json.Marshal(uploadSpec))),
		)).To(Succeed())

		Expect(env.CLI.GetOutput()).To(Not(BeEmpty()))
		var info ppi.UploadTargetSpecInfo
		Expect(json.Unmarshal(env.CLI.GetOutput(), &info)).To(Succeed())
		Expect(info.ConsumerId.Type()).To(Equal(helm.NAME))
	})

	It("Validate Upload Spec OK with full identity based on Artifact Set containing OCI Image", func(ctx SpecContext) {
		env.CLI.Command().SetContext(ctx)

		purl := Must(url.Parse("https://ocm.software:5501/my-artifactory"))
		uploadSpec := &helm.JFrogHelmUploaderSpec{
			ObjectVersionedType: runtime.ObjectVersionedType{Type: helm.VERSIONED_NAME},
			URL:                 purl.String(),
			Repository:          "my-repo",
		}

		Expect(env.Execute("upload", "validate",
			"--artifactType", artifacttypes.HELM_CHART,
			helm.NAME,
			string(Must(json.Marshal(uploadSpec))),
		)).To(Succeed())

		var info ppi.UploadTargetSpecInfo
		output := env.CLI.GetOutput()
		Expect(output).To(Not(BeEmpty()))
		Expect(json.Unmarshal(output, &info)).To(Succeed())

		Expect(info.ConsumerId.Type()).To(Equal(helm.NAME))
		Expect(info.ConsumerId.Match(credentials.ConsumerIdentity{
			helm.ID_TYPE:       helm.NAME,
			helm.ID_HOSTNAME:   purl.Hostname(),
			helm.ID_PORT:       purl.Port(),
			helm.ID_REPOSITORY: "my-repo",
		})).To(BeTrue(), "the identity should contain all attributes relevant to"+
			" match the correct repository for a resource transfer")
	})

	BeforeEach(func() {
		reindexedAfterUpload = false
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.String(), "reindex") {
				w.WriteHeader(http.StatusOK)
				reindexedAfterUpload = true
				return
			}

			if r.Method != http.MethodPut {
				http.Error(w, fmt.Sprintf("expected %s request, got %s", http.MethodPut, r.Method), http.StatusBadRequest)
			}

			u, p, ok := r.BasicAuth()
			if !ok {
				http.Error(w, fmt.Sprintf("expected basic auth header, got %s,%s but expected %s,%s", u, p, user, pass), http.StatusBadRequest)
			}

			data := Must(io.ReadAll(r.Body))
			unzipped, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to unzip data: %v", err), http.StatusBadRequest)
			}
			var buf bytes.Buffer
			io.Copy(&buf, unzipped)
			unzipped.Close()
			bufBytes := buf.Bytes()
			var compressed bytes.Buffer
			writer := gzip.NewWriter(&compressed)
			io.Copy(writer, bytes.NewReader(bufBytes))
			writer.Close()
			data = compressed.Bytes()

			chart, err := loader.LoadArchive(bytes.NewReader(data))
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to load chart: %s", err.Error()), http.StatusBadRequest)
			}

			resData, err := json.Marshal(&helm.ArtifactoryUploadResponse{
				DownloadUri: fmt.Sprintf("%s/my-repo/%s-%s.tgz", server.URL, chart.Name(), chart.Metadata.Version),
				Path:        "/mocked/chart.tgz",
				CreatedBy:   "mocked",
				Created:     time.Now().Format(time.RFC3339),
				Repo:        "my-repo",
				MimeType:    helm.MEDIA_TYPE,
				Checksums: helm.ArtifactoryUploadChecksums{
					Sha256: r.Header.Get("X-Checksum-Sha256"),
				},
				Size: strconv.Itoa(len(data)),
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to marshal response: %v", err), http.StatusInternalServerError)
			}

			// mimic the upload response
			w.WriteHeader(http.StatusCreated)
			if _, err := io.Copy(w, bytes.NewReader(resData)); err != nil {
				Fail(fmt.Sprintf("failed to write response: %v", err))
			}
		}))

		DeferCleanup(server.Close)
	})

	It("Upload Artifact Set to Server (no reindex) with Basic Auth", func(ctx SpecContext) {
		env.CLI.Command().SetContext(ctx)

		testDataPath := Must(filepath.Abs("../../api/ocm/extensions/download/handlers/helm/testdata/test-chart-oci-artifact.tgz"))
		testDataFile := Must(os.OpenFile(testDataPath, os.O_RDONLY, 0o400))
		DeferCleanup(testDataFile.Close)
		testData := Must(io.ReadAll(testDataFile))

		purl := Must(helm.ParseURLAllowNoScheme(server.URL))
		uploadSpec := &helm.JFrogHelmUploaderSpec{
			ObjectVersionedType: runtime.ObjectVersionedType{Type: helm.VERSIONED_NAME},
			URL:                 purl.String(),
			Repository:          "my-repo",
		}

		env.CLI.SetInput(io.NopCloser(bytes.NewReader(testData)))
		Expect(env.Execute("upload", "put",
			"--artifactType", artifacttypes.HELM_CHART,
			"--mediaType", artifactset.MediaType(artdesc.MediaTypeImageManifest),
			"--credentials", creds,
			helm.NAME,
			string(Must(json.Marshal(uploadSpec)))),
		).To(Succeed())

		var spec helmaccess.AccessSpec
		Expect(json.Unmarshal(env.GetOutput(), &spec)).To(Succeed())
		Expect(spec).To(Not(BeNil()))
		Expect(spec.HelmRepository).To(Equal(fmt.Sprintf("%s/artifactory/api/helm/%s", server.URL, "my-repo")))
		Expect(spec.HelmChart).To(ContainSubstring(":"), "helm chart is separated with version")
		splitChart := strings.Split(spec.HelmChart, ":")
		Expect(splitChart).To(HaveLen(2), "helm chart is separated with version")
		Expect(splitChart[0]).To(Equal("test-chart"), "the chart name should be test-chart")
		Expect(splitChart[1]).To(Equal("0.1.0"), "the chart version should be 0.1.0")
		Expect(reindexedAfterUpload).To(BeFalse(), "the server should not have been re-indexed as it wasn't requested explicitly")
	})

	It("Upload Artifact Set to Server (with reindex) with Basic Auth", func(ctx SpecContext) {
		env.CLI.Command().SetContext(ctx)

		testDataPath := Must(filepath.Abs("../../api/ocm/extensions/download/handlers/helm/testdata/test-chart-oci-artifact.tgz"))
		testDataFile := Must(os.OpenFile(testDataPath, os.O_RDONLY, 0o400))
		DeferCleanup(testDataFile.Close)
		testData := Must(io.ReadAll(testDataFile))

		purl := Must(helm.ParseURLAllowNoScheme(server.URL))
		uploadSpec := &helm.JFrogHelmUploaderSpec{
			ObjectVersionedType: runtime.ObjectVersionedType{Type: helm.VERSIONED_NAME},
			URL:                 purl.String(),
			Repository:          "my-repo",
			ReIndexAfterUpload:  true,
		}

		env.CLI.SetInput(io.NopCloser(bytes.NewReader(testData)))
		Expect(env.Execute("upload", "put",
			"--artifactType", artifacttypes.HELM_CHART,
			"--mediaType", artifactset.MediaType(artdesc.MediaTypeImageManifest),
			"--credentials", creds,
			helm.NAME,
			string(Must(json.Marshal(uploadSpec)))),
		).To(Succeed())

		var spec helmaccess.AccessSpec
		Expect(json.Unmarshal(env.GetOutput(), &spec)).To(Succeed())
		Expect(spec).To(Not(BeNil()))
		Expect(spec.HelmRepository).To(Equal(fmt.Sprintf("%s/artifactory/api/helm/%s", server.URL, "my-repo")))
		Expect(spec.HelmChart).To(ContainSubstring(":"), "helm chart is separated with version")
		splitChart := strings.Split(spec.HelmChart, ":")
		Expect(splitChart).To(HaveLen(2), "helm chart is separated with version")
		Expect(splitChart[0]).To(Equal("test-chart"), "the chart name should be test-chart")
		Expect(splitChart[1]).To(Equal("0.1.0"), "the chart version should be 0.1.0")
		Expect(reindexedAfterUpload).To(BeTrue(), "the server should not have been re-indexed as it wasn't requested explicitly")
	})
})
