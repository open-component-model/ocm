package mvn

import (
	"bytes"
	"context"
	"crypto"
	"fmt"
	"io"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/mvn/identity"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type of Maven (mvn) repository.
const (
	Type   = "mvn"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// AccessSpec describes the access for a Maven (mvn) artifact.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Repository is the base URL of the Maven (mvn) repository.
	Repository string `json:"repository"`
	// ArtifactId is the name of Maven (mvn) artifact.
	GroupId string `json:"groupId"`
	// ArtifactId is the name of Maven (mvn) artifact.
	ArtifactId string `json:"artifactId"`
	// Version of the Maven (mvn) artifact.
	Version string `json:"version"`
	// Classifier of the Maven (mvn) artifact.
	Classifier string `json:"classifier"`
	// Extension of the Maven (mvn) artifact.
	Extension string `json:"extension"`
}

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

// New creates a new Maven (mvn) repository access spec version v1.
func New(repository, groupId, artifactId, version string, options ...func(*AccessSpec)) *AccessSpec {
	accessSpec := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Repository:          repository,
		GroupId:             groupId,
		ArtifactId:          artifactId,
		Version:             version,
		Classifier:          "",
		Extension:           "jar",
	}
	for _, option := range options {
		option(accessSpec)
	}
	return accessSpec
}

// WithClassifier sets the classifier of the Maven (mvn) artifact.
func WithClassifier(classifier string) func(*AccessSpec) {
	return func(a *AccessSpec) {
		a.Classifier = classifier
	}
}

// WithExtension sets the extension of the Maven (mvn) artifact.
func WithExtension(extension string) func(*AccessSpec) {
	return func(a *AccessSpec) {
		a.Extension = extension
	}
}

func (a *AccessSpec) Describe(_ accspeccpi.Context) string {
	return fmt.Sprintf("Maven (mvn) package %s:%s:%s in repository %s", a.GroupId, a.ArtifactId, a.Version, a.Repository)
}

func (_ *AccessSpec) IsLocal(accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(_ accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

// GetReferenceHint returns the reference hint for the Maven (mvn) artifact. In the following form:
// groupId:artifactId:version:classifier:extension
func (a *AccessSpec) GetReferenceHint(_ accspeccpi.ComponentVersionAccess) string {
	return a.GroupId + ":" + a.ArtifactId + ":" + a.Version + ":" + a.Classifier + ":" + a.Extension
}

func (_ *AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(c accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(newMethod(c, a))
}

func (a *AccessSpec) GetInexpensiveContentVersionIdentity(access accspeccpi.ComponentVersionAccess) string {
	meta, _ := a.GetPackageMeta(access.GetContext())
	if meta != nil {
		return meta.Hash
	}
	return ""
}

func (a *AccessSpec) BaseUrl() string {
	return a.Repository + "/" + a.AsArtifact().GavPath()
}

func (a *AccessSpec) ArtifactUrl() string {
	return a.Repository + a.AsArtifact().Path()
}

func (a *AccessSpec) AsArtifact() *Artifact {
	return &Artifact{
		GroupId:    a.GroupId,
		ArtifactId: a.ArtifactId,
		Version:    a.Version,
		Classifier: a.Classifier,
		Extension:  a.Extension,
	}
}

type meta struct {
	Packaging string      `json:"packaging"`
	HashType  crypto.Hash `json:"hashType"`
	Hash      string      `json:"hash"`
	Asc       string      `json:"asc"`
	Bin       string      `json:"bin"`
}

func (a *AccessSpec) GetPackageMeta(ctx accspeccpi.Context) (*meta, error) {
	// this is how the usual maven repository structure looks like
	baseUrl := a.Repository + path.Join("/", strings.ReplaceAll(a.GroupId, ".", "/"), a.ArtifactId, a.Version, a.ArtifactId+"-"+a.Version)
	//fs := vfsattr.Get(ctx)

	// first let's read the pom file and check which binary artifact we need to read
	log := logging.Context().Logger(identity.REALM)
	log.Debug("Reading ", "repository", baseUrl)

	/*
		pom, err := readPom(urlPrefix+"pom", fs)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read GAV: %s", baseUrl)
		}

		hashType, hash, err := getBestHashValue(urlPrefix+pom.Packaging, fs)
		if err != nil {
			log.Debug("Could not find any hash value for ", "file", urlPrefix+pom.Packaging)
		}
		asc, err := getStringData(urlPrefix+pom.Packaging+".asc", fs)
		if err != nil {
			log.Debug("No signing info found for ", "file", urlPrefix+pom.Packaging)
		}


	*/
	var metadata meta = meta{
		Packaging: "",
		HashType:  crypto.MD5,
		Hash:      "hash",
		Bin:       "urlPrefix + pom.Packaging",
		Asc:       "asc",
	}

	return &metadata, nil
}

func (a *AccessSpec) GetGAVFiles() (map[string]crypto.Hash, error) {
	// Create a new request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, a.BaseUrl(), nil)
	if err != nil {
		return nil, err
	}
	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Check the response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s - %s", a.BaseUrl(), resp.Status)
	}
	htmlDoc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	// Which files are listed in the repository?
	var fileList []string
	var process func(*html.Node)
	prefix := a.AsArtifact().FilePrefix()
	process = func(node *html.Node) {
		// check if the node is an element node and the tag is "<a href="..." />"
		if node.Type == html.ElementNode && node.Data == "a" {
			// iterate over the attr and read the href attribute
			for _, attribute := range node.Attr {
				if attribute.Key == "href" {
					// check if the href starts with artifactId-version
					if strings.HasPrefix(attribute.Val, prefix) {
						fileList = append(fileList, attribute.Val)
					}
				}
			}
		}
		for nextChild := node.FirstChild; nextChild != nil; nextChild = nextChild.NextSibling {
			process(nextChild) // recursive call!
		}
	}
	process(htmlDoc)

	var filesAndHashes map[string]crypto.Hash
	filesAndHashes = make(map[string]crypto.Hash)
	for _, file := range fileList {
		if IsArtifact(file) {
			filesAndHashes[file] = bestAvailableHash(fileList, file)
		}
	}

	sort.Strings(fileList)
	return filesAndHashes, nil
}

func bestAvailableHash(list []string, filename string) crypto.Hash {
	hashes := [5]crypto.Hash{crypto.SHA512, crypto.SHA256, crypto.SHA1, crypto.MD5}
	for _, hash := range hashes {
		if slices.Contains(list, filename+hashUrlExt(hash)) {
			return hash
		}
	}
	return 0
}

////////////////////////////////////////////////////////////////////////////////

/*
type project struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Classifier  string `xml:"packaging"`
}

func readPom(url string, fs vfs.FileSystem) (*project, error) {
	reader, err := getReader(url, fs)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, io.LimitReader(reader, 200000))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get version metadata for %s", url)
	}
	var pom project
	err = xml.Unmarshal(buf.Bytes(), &pom)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal version metadata for %s", url)
	}
	if pom.Classifier == "" {
		pom.Classifier = "jar"
	}
	return &pom, nil
}
*/

// getStringData reads all data from the given URL and returns it as a string.
func getStringData(url string, fs vfs.FileSystem) (string, error) {
	r, err := getReader(url, fs)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// getBestHashValue returns the best hash value (SHA-512, SHA-256, SHA-1, MD5) for the given artifact (POM, JAR, ...).
func getBestHashValue(url string, fs vfs.FileSystem) (crypto.Hash, string, error) {
	arr := [5]crypto.Hash{crypto.SHA512, crypto.SHA256, crypto.SHA1, crypto.MD5}
	log := logging.Context().Logger(identity.REALM)
	for i := 0; i < len(arr); i++ {
		v, err := getStringData(url+hashUrlExt(arr[i]), fs)
		if v != "" {
			log.Debug("found hash ", "url", url+hashUrlExt(arr[i]))
			return arr[i], v, err
		}
		if err != nil {
			log.Debug("hash file not found", "url", url+hashUrlExt(arr[i]))
		}
	}
	return 0, "", errors.New("no hash value found")
}

// hashUrlExt returns the 'maven' hash extension for the given hash.
// Maven usually uses sha1, sha256, sha512, md5 instead of SHA-1, SHA-256, SHA-512, MD5.
func hashUrlExt(h crypto.Hash) string {
	return "." + strings.ReplaceAll(strings.ToLower(h.String()), "-", "")
}

func newMethod(c accspeccpi.ComponentVersionAccess, a *AccessSpec) (accspeccpi.AccessMethodImpl, error) {
	factory := func() (blobaccess.BlobAccess, error) {
		meta, err := a.GetPackageMeta(c.GetContext())
		if err != nil {
			return nil, err
		}

		reader := func() (io.ReadCloser, error) {
			return getReader(meta.Bin, vfsattr.Get(c.GetContext()))
		}
		if meta.Hash != "" {
			getreader := reader
			reader = func() (io.ReadCloser, error) {
				readCloser, err := getreader()
				if err != nil {
					return nil, err
				}
				return iotools.VerifyingReaderWithHash(readCloser, meta.HashType, meta.Hash), nil
			}
		}
		acc := blobaccess.DataAccessForReaderFunction(reader, meta.Bin)
		return accessobj.CachedBlobAccessForWriter(c.GetContext(), mime.MIME_JAR, accessio.NewDataAccessWriter(acc)), nil
	}
	return accspeccpi.NewDefaultMethodImpl(c, a, "", mime.MIME_JAR, factory), nil
}

func getReader(url string, fs vfs.FileSystem) (io.ReadCloser, error) {
	c := &http.Client{}

	if strings.HasPrefix(url, "file://") {
		path := url[7:]
		return fs.OpenFile(path, vfs.O_RDONLY, 0o600)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, io.LimitReader(resp.Body, 2000))
		if err != nil {
			return nil, errors.Newf("version meta data request %s provides %s", url, resp.Status)
		}
		return nil, errors.Newf("version meta data request %s provides %s: %s", url, resp.Status, buf.String())
	}
	return resp.Body, nil
}
