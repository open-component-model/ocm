package github_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	_ "ocm.software/ocm/api/datacontext/config"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	me "ocm.software/ocm/api/ocm/extensions/accessmethods/github"
	"ocm.software/ocm/api/tech/github/identity"
)

type mockDownloader struct {
	expected []byte
	err      error
}

func (m *mockDownloader) Download(w io.WriterAt) error {
	if _, err := w.WriteAt(m.expected, 0); err != nil {
		return fmt.Errorf("failed to write to mock writer: %w", err)
	}
	return m.err
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

var _ = Describe("Method", func() {
	var (
		ctx                 ocm.Context
		expectedBlobContent []byte
		err                 error
		defaultLink         string
		accessSpec          *me.AccessSpec
		fs                  vfs.FileSystem
		expectedURL         string
		clientFn            func(url string) *http.Client
	)

	BeforeEach(func() {
		ctx = ocm.New()
		expectedBlobContent, err = os.ReadFile(filepath.Join("testdata", "repo.tar.gz"))
		Expect(err).ToNot(HaveOccurred())
		defaultLink = "https://github.com/test/test/sha?token=token"
		expectedURL = "https://api.github.com/repos/test/test/tarball/7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f"

		clientFn = func(url string) *http.Client {
			return NewTestClient(func(req *http.Request) *http.Response {
				if req.URL.String() != url {
					Fail(fmt.Sprintf("failed to match url to expected url. want: %s; got: %s", expectedURL, req.URL.String()))
				}
				return &http.Response{
					StatusCode: http.StatusFound,
					Status:     http.StatusText(http.StatusFound),
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					Header: http.Header{
						"Location": []string{defaultLink},
					},
				}
			})
		}

		accessSpec = me.New(
			"https://github.com/test/test",
			"",
			"7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f",
			me.WithClient(clientFn(expectedURL)),
			me.WithDownloader(&mockDownloader{
				expected: expectedBlobContent,
			}),
		)
		fs, err = osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		vfsattr.Set(ctx, fs)
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: "/tmp", Filesystem: fs})
	})

	AfterEach(func() {
		vfs.Cleanup(fs)
	})

	It("provides consumer id", func() {
		m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
		Expect(err).ToNot(HaveOccurred())
		Expect(credentials.GetProvidedConsumerId(m)).To(Equal(credentials.NewConsumerIdentity(identity.CONSUMER_TYPE,
			identity.ID_HOSTNAME, "github.com",
			identity.ID_PATHPREFIX, "test/test")))
	})

	It("downloads artifacts", func() {
		m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
		Expect(err).ToNot(HaveOccurred())
		content, err := m.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(content).To(Equal(expectedBlobContent))
	})

	When("the commit sha is of an invalid length", func() {
		It("errors", func() {
			accessSpec := me.New(
				"hostname",
				"",
				"not-a-sha",
				me.WithClient(clientFn(expectedURL)),
				me.WithDownloader(&mockDownloader{
					expected: expectedBlobContent,
				}),
			)
			m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).To(MatchError(ContainSubstring("commit is not a SHA")))
			if m != nil {
				m.Close()
			}
		})
	})

	When("the commit sha is of the right length but contains invalid characters", func() {
		It("errors", func() {
			accessSpec := me.New(
				"hostname",
				"1234",
				"refs/heads/veryinteresting_branch_namess",
				me.WithClient(clientFn(expectedURL)),
				me.WithDownloader(&mockDownloader{
					expected: expectedBlobContent,
				}),
			)
			m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).To(MatchError(ContainSubstring("commit contains invalid characters for a SHA")))
			if m != nil {
				m.Close()
			}
		})
	})

	When("credentials are provided", func() {
		BeforeEach(func() {
			clientFn = func(url string) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					if v, ok := req.Header["Authorization"]; ok {
						Expect(v).To(ContainElement("Bearer test"))
					} else {
						Fail("Authorization header not found in request")
					}
					if req.URL.String() != url {
						Fail(fmt.Sprintf("failed to match url to expected url. want: %s; got: %s", expectedURL, req.URL.String()))
					}
					return &http.Response{
						StatusCode: http.StatusFound,
						Status:     http.StatusText(http.StatusFound),
						// Must be set to non-nil value or it panics
						Body: io.NopCloser(bytes.NewBufferString(`{}`)),
						Header: http.Header{
							"Location": []string{defaultLink},
						},
					}
				})
			}
			accessSpec = me.New(
				"https://github.com/test/test",
				"",
				"7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f",
				me.WithClient(clientFn(expectedURL)),
				me.WithDownloader(&mockDownloader{
					expected: expectedBlobContent,
				}),
			)
		})
		It("can use those to access private repos", func() {
			mcc := ocm.New(datacontext.MODE_INITIAL)
			src := &mockCredSource{
				Context: mcc.CredentialsContext(),
				cred: credentials.DirectCredentials{
					credentials.ATTR_TOKEN: "test",
				},
			}
			mcc.CredentialsContext().SetCredentialsForConsumer(credentials.NewConsumerIdentity(identity.CONSUMER_TYPE), src)
			m, err := accessSpec.AccessMethod(&mockComponentVersionAccess{
				ocmContext: mcc,
			})
			Expect(err).ToNot(HaveOccurred())
			_, err = m.Get()
			Expect(err).ToNot(HaveOccurred())
			m.Close()
			Expect(src.called).To(BeTrue())
		})
	})

	When("GetCredentialsForConsumer returns an error", func() {
		It("errors", func() {
			mcc := ocm.New(datacontext.MODE_INITIAL)
			src := &mockCredSource{
				Context: mcc.CredentialsContext(),
				err:     fmt.Errorf("danger will robinson"),
			}
			mcc.CredentialsContext().SetCredentialsForConsumer(credentials.NewConsumerIdentity(identity.CONSUMER_TYPE), src)
			_, err := accessSpec.AccessMethod(&mockComponentVersionAccess{
				ocmContext: mcc,
			})
			Expect(err).To(MatchError(ContainSubstring("danger will robinson")))
			Expect(src.called).To(BeTrue())
		})
	})

	When("an enterprise repo URL is provided", func() {
		It("uses that domain and includes api/v3 in the request URL", func() {
			expectedURL = "https://github.tools.sap/api/v3/repos/test/test/tarball/25d9a3f0031c0b42e9ef7ab0117c35378040ef82"
			spec := me.New("https://github.tools.sap/test/test", "", "25d9a3f0031c0b42e9ef7ab0117c35378040ef82", me.WithClient(clientFn(expectedURL)))
			_, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("hostname is different from github.com", func() {
		It("will use an enterprise client", func() {
			expectedURL = "https://custom/api/v3/repos/test/test/tarball/25d9a3f0031c0b42e9ef7ab0117c35378040ef82"
			spec := me.New("https://github.tools.sap/test/test", "custom", "25d9a3f0031c0b42e9ef7ab0117c35378040ef82", me.WithClient(clientFn(expectedURL)))
			_, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("repoURL doesn't have an https prefix", func() {
		It("will add one", func() {
			expectedURL = "https://api.github.com/repos/test/test/tarball/25d9a3f0031c0b42e9ef7ab0117c35378040ef82"
			spec := me.New("github.com/test/test", "", "25d9a3f0031c0b42e9ef7ab0117c35378040ef82", me.WithClient(clientFn(expectedURL)))
			_, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

type mockComponentVersionAccess struct {
	ocm.ComponentVersionAccess
	ocmContext ocm.Context
}

func (m *mockComponentVersionAccess) GetContext() ocm.Context {
	return m.ocmContext
}

type mockCredSource struct {
	credentials.Context
	cred   credentials.Credentials
	called bool
	err    error
}

func (m *mockCredSource) Credentials(credentials.Context, ...credentials.CredentialsSource) (credentials.Credentials, error) {
	m.called = true
	return m.cred, m.err
}
