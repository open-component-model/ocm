package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/tech/helm"
	"ocm.software/ocm/api/tech/helm/loader"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	// MEDIA_TYPE is the media type of the HELM Chart artifact as tgz.
	// It is the definitive format for JFrog Uploads
	MEDIA_TYPE = helm.ChartMediaType

	// NAME of the Uploader and the Configuration
	NAME = "JFrogHelm"

	// VERSION of the Uploader TODO Increment once stable
	VERSION = "v1alpha1"

	// VERSIONED_NAME is the name of the Uploader including the version
	VERSIONED_NAME = NAME + runtime.VersionSeparator + VERSION

	// ID_TYPE is the type of the JFrog Artifactory credentials
	ID_TYPE = cpi.ID_TYPE
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

	// ReIndexAfterUpload is a flag to indicate if the chart should be reindexed after upload or not.
	// OPTIONAL: If not set, defaulted to false.
	ReIndexAfterUpload bool `json:"reindexAfterUpload,omitempty"`
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
	types = ppi.UploadFormats{VERSIONED_NAME: decoder}
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
	client := &http.Client{}
	// we do not want to double compress helm tgz files
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableCompression = true
	client.Transport = transport
	return &Uploader{
		UploaderBase: ppi.MustNewUploaderBase(NAME, "upload artifacts to JFrog HELM repositories by using the JFrog REST API."),
		Client:       client,
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
func (a *Uploader) Upload(
	ctx context.Context,
	p ppi.Plugin,
	arttype, mediaType, _, digest string,
	targetSpec ppi.UploadTargetSpec,
	creds credentials.Credentials,
	reader io.Reader,
) (ppi.AccessSpecProvider, error) {
	if arttype != artifacttypes.HELM_CHART {
		return nil, fmt.Errorf("unsupported artifact type %s", arttype)
	}

	spec, ok := targetSpec.(*JFrogHelmUploaderSpec)
	if !ok {
		return nil, fmt.Errorf("the type %T is not a valid target spec type", spec)
	}

	switch mediaType {
	case helm.ChartMediaType:
		// if it is a native chart tgz we can pass it on as is
	case artifactset.MediaType(artdesc.MediaTypeImageManifest):
		// if we have an artifact set (ocm custom version of index + locally colocated blobs as files, we need
		// to translate it. This translation is not perfect because the ociArtifactDigest that might be
		// generated in OCM is not the same as the one that is used within Artifactory, but Uploaders
		// do not have a way of providing back digest information to the caller.
		// TODO: At some point consider this for a plugin rework.
		var err error
		if reader, digest, err = ConvertArtifactSetWithOCIImageHelmChartToPlainTGZChart(reader); err != nil {
			return nil, fmt.Errorf("failed to convert OCI Helm Chart to plain TGZ: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported media type %s", mediaType)
	}

	var buf bytes.Buffer
	chart, err := loader.LoadArchive(io.TeeReader(reader, &buf))
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}
	spec.Name = chart.Metadata.Name
	spec.Version = chart.Metadata.Version
	reader = &buf

	// now based on the chart and repository we can upload it to the correct location.
	targetURL, err := ConvertTargetSpecToHelmUploadURL(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to convert target spec to URL: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, spec.GetTimeout())
	defer cancel()

	access, err := Upload(ctx, reader, a.Client, targetURL, creds, digest)
	if err != nil {
		return nil, fmt.Errorf("failed to upload: %w", err)
	}

	if spec.ReIndexAfterUpload {
		if err := ReindexChart(ctx, a.Client, spec.URL, spec.Repository, creds); err != nil {
			return nil, fmt.Errorf("failed to reindex chart: %w", err)
		}
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
	purl, err := ParseURLAllowNoScheme(spec.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	var info ppi.UploadTargetSpecInfo

	// By default, we identify an artifactory repository as a combination
	// of Host & Repository
	info.ConsumerId = credentials.ConsumerIdentity{
		ID_TYPE:       NAME,
		ID_HOSTNAME:   purl.Hostname(),
		ID_REPOSITORY: spec.Repository,
	}
	if purl.Port() != "" {
		info.ConsumerId.SetNonEmptyValue(ID_PORT, purl.Port())
	}

	return &info, nil
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
	requestURLParsed, err := ParseURLAllowNoScheme(spec.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse full request URL: %w", err)
	}
	requestURLParsed.Path = path.Join("artifactory", spec.Repository, fmt.Sprintf("%s-%s.tgz", spec.Name, spec.Version))
	return requestURLParsed, nil
}

// ParseURLAllowNoScheme is an adaptation / hack on url.Parse because
// url.Parse does not support parsing a URL without a prefixed scheme.
// However, we would like to accept these kind of URLs because we default them
// to "https://" out of convenience.
func ParseURLAllowNoScheme(urlToParse string) (*url.URL, error) {
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
