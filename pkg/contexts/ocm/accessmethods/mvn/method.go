package mvn

import (
	"bytes"
	"context"
	"crypto"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
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
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
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

	Artifact `json:",inline"`
}

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

var log = logging.DynamicLogger(identity.REALM)

// New creates a new Maven (mvn) repository access spec version v1.
func New(repository, groupId, artifactId, version string, options ...func(*AccessSpec)) *AccessSpec {
	accessSpec := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Repository:          repository,
		Artifact: Artifact{
			GroupId:    groupId,
			ArtifactId: artifactId,
			Version:    version,
			Classifier: "",
			Extension:  "",
		},
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

// GetReferenceHint returns the reference hint for the Maven (mvn) artifact.
func (a *AccessSpec) GetReferenceHint(_ accspeccpi.ComponentVersionAccess) string {
	return a.Serialize()
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
	return a.Repository + "/" + a.GavPath()
}

func (a *AccessSpec) ArtifactUrl() string {
	return a.Url(a.Repository)
}

func (a *AccessSpec) NewArtifact() *Artifact {
	return a.Artifact.Copy()
}

type meta struct {
	MimeType string      `json:"packaging"`
	HashType crypto.Hash `json:"hashType"`
	Hash     string      `json:"hash"`
	Bin      string      `json:"bin"`
}

func (a *AccessSpec) GetPackageMeta(ctx accspeccpi.Context) (*meta, error) {
	fs := vfsattr.Get(ctx)

	log := log.WithValues("BaseUrl", a.BaseUrl())
	fileMap, err := a.GavFiles(fs)
	if err != nil {
		return nil, err
	}

	if a.Classifier != "" {
		fileMap = filterByClassifier(fileMap, a.Classifier)
	}
	singleBinary := len(fileMap) == 1

	tempFs, err := osfs.NewTempFileSystem()
	if err != nil {
		return nil, err
	}
	defer vfs.Cleanup(tempFs)

	artifact := a.NewArtifact()
	metadata := meta{}
	for file, hash := range fileMap {
		artifact.ClassifierExtensionFrom(file)
		metadata.Bin = artifact.Url(a.Repository)
		log = log.WithValues("file", metadata.Bin)
		log.Debug("processing")
		metadata.MimeType = artifact.MimeType()
		if hash > 0 {
			metadata.HashType = hash
			metadata.Hash, err = getStringData(metadata.Bin+hashUrlExt(hash), fs)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot read %s digest of: %s", hash, metadata.Bin)
			}
		} else {
			log.Warn("no digest available")
		}

		// single binary dependency, this will never be a complete GAV - no maven uploader support!
		if a.Extension != "" || singleBinary && a.Classifier != "" {
			// in case you want to transport <packaging>pom</packaging>, then you should NOT set the extension
			return &metadata, nil
		}

		// download the artifact into the temporary file system
		out, err := tempFs.Create(file)
		if err != nil {
			return nil, err
		}
		defer out.Close()
		reader, err := getReader(metadata.Bin, fs)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		if hash > 0 {
			dreader := iotools.NewDigestReaderWithHash(hash, reader)
			_, err = io.Copy(out, dreader)
			if err != nil {
				return nil, err
			}
			sum := dreader.Digest().Encoded()
			if metadata.Hash != sum {
				return nil, errors.Newf("checksum mismatch for %s", metadata.Bin)
			}
		} else {
			_, err = io.Copy(out, reader)
			if err != nil {
				return nil, err
			}
		}
	}

	// pack all downloaded files into a tar.gz file
	tgz, err := vfs.TempFile(fs, "", Type+"-"+artifact.FileNamePrefix()+"-*.tar.gz")
	if err != nil {
		return nil, err
	}
	defer tgz.Close()

	dw := iotools.NewDigestWriterWith(digest.SHA256, tgz)
	defer dw.Close()
	err = tarutils.TgzFs(tempFs, dw)
	if err != nil {
		return nil, err
	}

	metadata.Bin = "file://" + tgz.Name()
	metadata.MimeType = mime.MIME_TGZ
	metadata.Hash = dw.Digest().Encoded()
	metadata.HashType = crypto.SHA256
	log.Debug("created", "file", metadata.Bin)

	return &metadata, nil
}

func filterByClassifier(fileMap map[string]crypto.Hash, classifier string) map[string]crypto.Hash {
	filtered := make(map[string]crypto.Hash)
	for file, hash := range fileMap {
		if strings.Contains(file, "-"+classifier+".") {
			filtered[file] = hash
		}
	}
	return filtered
}

func (a *AccessSpec) GavFiles(fs ...vfs.FileSystem) (map[string]crypto.Hash, error) {
	if strings.HasPrefix(a.Repository, "file://") && len(fs) > 0 {
		dir := a.Repository[7:]
		return gavFilesFromDisk(fs[0], dir)
	}
	return a.gavOnlineFiles()
}

func gavFilesFromDisk(fs vfs.FileSystem, dir string) (map[string]crypto.Hash, error) {
	files, err := tarutils.ListSortedFilesInDir(fs, dir, true)
	if err != nil {
		return nil, err
	}
	return filesAndHashes(files), nil
}

// gavOnlineFiles returns the files of the Maven (mvn) artifact in the repository and their available digests.
func (a *AccessSpec) gavOnlineFiles() (map[string]crypto.Hash, error) {
	log := log.WithValues("BaseUrl", a.BaseUrl())
	log.Debug("gavOnlineFiles")

	reader, err := getReader(a.BaseUrl(), nil)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Which files are listed in the repository?
	log.Debug("parse-html")
	htmlDoc, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}
	var fileList []string
	var process func(*html.Node)
	prefix := a.FileNamePrefix()
	process = func(node *html.Node) {
		// check if the node is an element node and the tag is "<a href="..." />"
		if node.Type == html.ElementNode && node.Data == "a" {
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

	return filesAndHashes(fileList), nil
}

func filesAndHashes(fileList []string) map[string]crypto.Hash {
	// Sort the list of files, to ensure always the same results for e.g. identical tar.gz files.
	sort.Strings(fileList)

	// Which hash files are available?
	result := make(map[string]crypto.Hash)
	for _, file := range fileList {
		if IsResource(file) {
			result[file] = bestAvailableHash(fileList, file)
			log.Debug("found", "file", file)
		}
	}
	return result
}

// bestAvailableHash returns the best available hash for the given file.
// It first checks for SHA-512, then SHA-256, SHA-1, and finally MD5. If nothing is found, it returns 0.
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
		return accessobj.CachedBlobAccessForWriter(c.GetContext(), a.MimeType(), accessio.NewDataAccessWriter(acc)), nil
	}
	// FIXME add Digest!
	return accspeccpi.NewDefaultMethodImpl(c, a, "", a.MimeType(), factory), nil
}

func getReader(url string, fs vfs.FileSystem) (io.ReadCloser, error) {
	if strings.HasPrefix(url, "file://") {
		path := url[7:]
		return fs.OpenFile(path, vfs.O_RDONLY, 0o600)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, io.LimitReader(resp.Body, 2000))
		if err != nil {
			return nil, errors.Newf("http %s error - %s", resp.Status, url)
		}
		return nil, errors.Newf("http %s error - %s returned: %s", resp.Status, url, buf.String())
	}
	return resp.Body, nil
}
