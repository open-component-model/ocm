package uploaders

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/helm"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	NAME    = "JFrog"
	VERSION = "v1"

	ID_HOSTNAME   = hostpath.ID_HOSTNAME
	ID_PORT       = hostpath.ID_PORT
	ID_REPOSITORY = "repository"
)

type Config struct {
}

func GetConfig(raw json.RawMessage) (interface{}, error) {
	var cfg Config

	err := json.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

type HelmTargetSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// URL is the hostname of the JFrog instance
	URL string `json:"url"`
	url *url.URL

	// Repository is the repository to upload to
	Repository string `json:"repository"`

	ChartName string `json:"name"`
	// Version is the version of the chart
	ChartVersion string `json:"version"`
}

var types ppi.UploadFormats

func init() {
	decoder, err := runtime.NewDirectDecoder[runtime.TypedObject](&HelmTargetSpec{})
	if err != nil {
		panic(err)
	}
	types = ppi.UploadFormats{NAME + runtime.VersionSeparator + VERSION: decoder}
}

type Uploader struct {
	ppi.UploaderBase
}

var _ ppi.Uploader = (*Uploader)(nil)

func New() ppi.Uploader {
	return &Uploader{
		UploaderBase: ppi.MustNewUploaderBase(NAME, "upload artifacts to JFrog"),
	}
}

func (a *Uploader) Decoders() ppi.UploadFormats {
	return types
}

func (a *Uploader) ValidateSpecification(_ ppi.Plugin, spec ppi.UploadTargetSpec) (*ppi.UploadTargetSpecInfo, error) {
	var info ppi.UploadTargetSpecInfo
	my, ok := spec.(*HelmTargetSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", spec)
	}

	var err error
	if my.url, err = url.Parse(my.URL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	info.ConsumerId = credentials.ConsumerIdentity{
		cpi.ID_TYPE:   NAME,
		ID_HOSTNAME:   my.url.Hostname(),
		ID_PORT:       my.url.Port(),
		ID_REPOSITORY: my.Repository,
	}
	return &info, nil
}

func (a *Uploader) Upload(p ppi.Plugin, artifactType, mediatype, _ string, repo ppi.UploadTargetSpec, creds credentials.Credentials, reader io.Reader) (ppi.AccessSpecProvider, error) {
	cfg, err := p.GetConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "can't get config for access method %s", mediatype)
	}

	if artifactType != artifacttypes.HELM_CHART {
		return nil, fmt.Errorf("unsupported artifact type %s", artifactType)
	}

	if cfg != nil {
		_, ok := cfg.(Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type %T", cfg)
		}
	}

	my := repo.(*HelmTargetSpec)

	requestURL := path.Join(my.url.String(), "artifactory", my.Repository, fmt.Sprintf("%s-%s.tgz", my.ChartName, my.ChartVersion))

	req, err := http.NewRequest(http.MethodPost, requestURL, reader)
	if err != nil {
		return nil, err
	}

	if creds.ExistsProperty(credentials.ATTR_TOKEN) {
		req.Header.Set("Authorization", "Bearer "+creds.GetProperty(credentials.ATTR_TOKEN))
	} else {
		var user, pass string
		if creds.ExistsProperty(credentials.ATTR_USERNAME) {
			user = creds.GetProperty(credentials.ATTR_USERNAME)
		}
		if creds.ExistsProperty(credentials.ATTR_PASSWORD) {
			pass = creds.GetProperty(credentials.ATTR_PASSWORD)
		}
		req.SetBasicAuth(user, pass)
	}

	client := http.DefaultClient

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to store blob in artifactory: %w", err)
	}
	defer response.Body.Close()

	uploadResponse := &ArtifactoryUploadResponse{}
	if err := json.NewDecoder(response.Body).Decode(uploadResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	switch artifactType {
	case artifacttypes.HELM_CHART:
		spec, err := uploadResponse.ToHelmAccessSpec()
		if err != nil {
			return nil, err
		}
		return func() ppi.AccessSpec {
			return spec
		}, nil
	default:
		return nil, fmt.Errorf("unsupported artifact type %s", artifactType)
	}
}

type ArtifactoryUploadResponse struct {
	Repo        string `json:"repo,omitempty"`
	Path        string `json:"path,omitempty"`
	Created     string `json:"created,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	DownloadUri string `json:"downloadUri,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Size        string `json:"size,omitempty"`
	Checksums   struct {
		Sha1   string `json:"sha1,omitempty"`
		Sha256 string `json:"sha256,omitempty"`
		Md5    string `json:"md5,omitempty"`
	} `json:"checksums,omitempty"`
	Uri string `json:"uri"`
}

func (r *ArtifactoryUploadResponse) URL() string {
	if r.DownloadUri != "" {
		return r.DownloadUri
	}
	return r.Uri
}

func (r *ArtifactoryUploadResponse) ToHelmAccessSpec() (ppi.AccessSpec, error) {
	u := r.URL()
	urlp, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	chart := path.Base(urlp.Path)
	chart = strings.TrimSuffix(chart, path.Ext(chart))
	chart = strings.ReplaceAll(chart, "-", ":")

	repo := path.Join(urlp.Host, "artifactory", "api", "helm", r.Repo)

	return helm.New(chart, repo), nil
}
