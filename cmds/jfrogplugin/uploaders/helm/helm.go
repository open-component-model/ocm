package helm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/containerd/containerd/reference"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	NAME = "JFrogHelm"

	// VERSION of the Uploader TODO Increment once stable
	VERSION = "v1alpha1"

	// ID_HOSTNAME is the hostname of the artifactory server to upload to
	ID_HOSTNAME = hostpath.ID_HOSTNAME
	// ID_PORT is the port of the artifactory server to upload to
	ID_PORT = hostpath.ID_PORT
	// ID_REPOSITORY is the repository name in JFrog Artifactory to upload to
	ID_REPOSITORY = "repository"

	// DEFAULT_TIMEOUT is the default timeout for http requests issued by the uploader.
	DEFAULT_TIMEOUT = time.Minute
)

type JFrogHelmUploaderSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// URL is the hostname of the JFrog instance.
	// Required for correct reference to Artifactory.
	URL string `json:"url"`

	// Repository is the repository to upload to.
	// Required for correct reference to Artifactory.
	Repository string `json:"repository"`

	JFrogHelmChart `json:",inline"`

	// Timeout is the maximum duration the upload of the chart can take
	// before aborting and failing.
	// OPTIONAL: If not set, set to the internal DEFAULT_TIMEOUT.
	Timeout *time.Duration `json:"timeout,omitempty"`
}

type JFrogHelmChart struct {
	// ChartName is the desired name of the chart in the repository.
	// OPTIONAL: If not set, defaulted from the passed Hint.
	Name string `json:"name,omitempty"`
	// Version is the desired version of the chart
	// OPTIONAL: If not set, defaulted from the passed Hint.
	Version string `json:"version,omitempty"`
}

func (s *JFrogHelmUploaderSpec) GetTimeout() time.Duration {
	if s.Timeout == nil {
		return DEFAULT_TIMEOUT
	}
	return *s.Timeout
}

var types ppi.UploadFormats

func init() {
	decoder, err := runtime.NewDirectDecoder[runtime.TypedObject](&JFrogHelmUploaderSpec{})
	if err != nil {
		panic(err)
	}
	types = ppi.UploadFormats{NAME + runtime.VersionSeparator + VERSION: decoder}
}

func (a *Uploader) Decoders() ppi.UploadFormats {
	return types
}

type Uploader struct {
	ppi.UploaderBase
	*http.Client
}

var _ ppi.Uploader = (*Uploader)(nil)

func New() ppi.Uploader {
	return &Uploader{
		UploaderBase: ppi.MustNewUploaderBase(NAME, "upload artifacts to JFrog HELM repositories by using the JFrog REST API."),
		Client:       http.DefaultClient,
	}
}

func (a *Uploader) ValidateSpecification(_ ppi.Plugin, spec ppi.UploadTargetSpec) (*ppi.UploadTargetSpecInfo, error) {
	targetSpec, ok := spec.(*JFrogHelmUploaderSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", spec)
	}

	info, err := ConvertTargetSpecToInfo(targetSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to convert target spec to info: %w", err)
	}

	return info, nil
}

// Upload uploads any artifact that is of type artifacttypes.HELM_CHART.
// Process:
//  1. introspect the JFrogHelmUploaderSpec (cast from ppi.UploadTargetSpec) and hint parameter
//     (the hint is expected to be an OCI style reference, such as `repo/comp:version`)
//  2. building an Artifactory Style JFrog Upload URL out of it (see ConvertTargetSpecToHelmUploadURL),
//  3. creating a request respecting the passed credentials based on SetHeadersFromCredentials
//  4. uploading the passed blob as is (expected to be a tgz byte stream)
//  5. intepreting the JFrog API response, and converting it from ArtifactoryUploadResponse to ppi.AccessSpec
func (a *Uploader) Upload(_ ppi.Plugin, artifactType, _, hint string, targetSpec ppi.UploadTargetSpec, creds credentials.Credentials, reader io.Reader) (ppi.AccessSpecProvider, error) {
	if artifactType != artifacttypes.HELM_CHART {
		return nil, fmt.Errorf("unsupported artifact type %s", artifactType)
	}

	spec, ok := targetSpec.(*JFrogHelmUploaderSpec)
	if !ok {
		return nil, fmt.Errorf("the type %T is not a valid target spec type", spec)
	}

	if err := EnsureSpecWithHelpFromHint(spec, hint); err != nil {
		return nil, fmt.Errorf("could not ensure spec to be ready for upload: %w", err)
	}

	targetURL, err := ConvertTargetSpecToHelmUploadURL(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to convert target spec to URL: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), spec.GetTimeout())
	defer cancel()

	access, err := Upload(ctx, reader, a.Client, targetURL, creds)
	if err != nil {
		return nil, fmt.Errorf("failed to upload: %w", err)
	}

	return func() ppi.AccessSpec {
		return access
	}, nil
}

// ConvertTargetSpecToInfo converts the JFrogHelmUploaderSpec
// to a valid info block containing the consumer ID used
// in the library to identify the correct credentials that need to
// be passed to it.
func ConvertTargetSpecToInfo(spec *JFrogHelmUploaderSpec) (*ppi.UploadTargetSpecInfo, error) {
	purl, err := parseURLAllowNoScheme(spec.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	var info ppi.UploadTargetSpecInfo

	// By default, we identify an artifactory repository as a combination
	// of Host & Repository
	info.ConsumerId = credentials.ConsumerIdentity{
		cpi.ID_TYPE:   NAME,
		ID_HOSTNAME:   purl.Hostname(),
		ID_REPOSITORY: spec.Repository,
	}
	if purl.Port() != "" {
		info.ConsumerId.SetNonEmptyValue(ID_PORT, purl.Port())
	}

	return &info, nil
}

// EnsureSpecWithHelpFromHint introspects the hint and fills the target spec based on it.
// It makes sure that the spec can be used to access a JFrog Artifactory HELM Repository.
func EnsureSpecWithHelpFromHint(spec *JFrogHelmUploaderSpec, hint string) error {
	if refFromHint, err := reference.Parse(hint); err == nil {
		if refFromHint.Digest() != "" && refFromHint.Object == "" {
			return fmt.Errorf("the hint contained a valid reference but it was a digest, so it cannot be used to deduce a version of the helm chart: %s", refFromHint)
		}
		if spec.Version == "" {
			spec.Version = refFromHint.Object
		}
		if spec.Name == "" {
			spec.Name = path.Base(refFromHint.Locator)
		}
	}
	if spec.Name == "" {
		return fmt.Errorf("the chart name could not be deduced from the hint (%s) or the config (%s)", hint, spec)
	}
	if spec.Version == "" {
		return fmt.Errorf("the chart version could not be deduced from the hint (%s) or the config (%s)", hint, spec)
	}
	return nil
}

// ConvertTargetSpecToHelmUploadURL interprets the JFrogHelmUploaderSpec into a valid REST API Endpoint URL to upload to.
// It requires a valid ChartName and ChartVersion to determine the correct URL endpoint.
//
//	See https://jfrog.com/help/r/jfrog-rest-apis/deploy-artifact for the URL endpoint
//	See https://jfrog.com/help/r/jfrog-artifactory-documentation/deploying-artifacts for artifact deployment reference
//	See https://jfrog.com/help/r/jfrog-artifactory-documentation/use-the-jfrog-helm-client for the HELM Client reference.
//
// Example:
//
//	JFrogHelmUploaderSpec.URL => demo.jfrog.ocm.software
//	JFrogHelmUploaderSpec.Repository => my-charts
//	JFrogHelmUploaderSpec.ChartName => podinfo
//	JFrogHelmUploaderSpec.ChartVersion => 0.0.1
//
// will result in
//
//	url.URL => https://demo.jfrog.ocm.software/artifactory/my-charts/podinfo-0.0.1.tgz
func ConvertTargetSpecToHelmUploadURL(spec *JFrogHelmUploaderSpec) (*url.URL, error) {
	requestURL := path.Join(spec.URL, "artifactory", spec.Repository, fmt.Sprintf("%s-%s.tgz", spec.Name, spec.Version))
	requestURLParsed, err := parseURLAllowNoScheme(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse full request URL: %w", err)
	}
	return requestURLParsed, nil
}

// parseURLAllowNoScheme is an adaptation / hack on url.Parse because
// url.Parse does not support parsing a URL without a prefixed scheme.
// However, we would like to accept these kind of URLs because we default them
// to "https://" out of convenience.
func parseURLAllowNoScheme(urlToParse string) (*url.URL, error) {
	const dummyScheme = "dummy"
	if !strings.Contains(urlToParse, "://") {
		urlToParse = dummyScheme + "://" + urlToParse
	}
	parsedURL, err := url.Parse(urlToParse)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme == dummyScheme {
		parsedURL.Scheme = ""
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}
	return parsedURL, nil
}
