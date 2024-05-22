package mvn

import (
	"bytes"
	"context"
	"crypto"
	"fmt"
	"github.com/open-component-model/ocm/pkg/maven"
	"io"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/maven/identity"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
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

	maven.Coordinates `json:",inline"`
}

// Option defines the interface function "ApplyTo()".
type Option = optionutils.Option[*AccessSpec]

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

var log = logging.DynamicLogger(identity.REALM)

// New creates a new Maven (mvn) repository access spec version v1.
func New(repository, groupId, artifactId, version string, options ...Option) *AccessSpec {
	accessSpec := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Repository:          repository,
		Coordinates: maven.Coordinates{
			GroupId:    groupId,
			ArtifactId: artifactId,
			Version:    version,
			Classifier: "",
			Extension:  "",
		},
	}
	optionutils.ApplyOptions(accessSpec, options...)
	return accessSpec
}

// classifier Option for Maven (mvn) Coordinates.
type classifier string

func (c classifier) ApplyTo(a *AccessSpec) {
	a.Classifier = string(c)
}

// WithClassifier sets the classifier of the Maven (mvn) artifact.
func WithClassifier(c string) Option {
	return classifier(c)
}

// extension Option for Maven (mvn) Coordinates.
type extension string

func (e extension) ApplyTo(a *AccessSpec) {
	a.Extension = string(e)
}

// WithExtension sets the extension of the Maven (mvn) artifact.
func WithExtension(e string) Option {
	return extension(e)
}

func (a *AccessSpec) Describe(_ accspeccpi.Context) string {
	return fmt.Sprintf("Maven (mvn) package '%s' in repository '%s' path '%s'", a.Coordinates.String(), a.Repository, a.Coordinates.FilePath())
}

func (_ *AccessSpec) IsLocal(accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(_ accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

// GetReferenceHint returns the reference hint for the Maven (mvn) artifact.
func (a *AccessSpec) GetReferenceHint(_ accspeccpi.ComponentVersionAccess) string {
	return a.String()
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

func (a *AccessSpec) GetCoordinates() *maven.Coordinates {
	return a.Coordinates.Copy()
}

type meta struct {
	MimeType string      `json:"packaging"`
	HashType crypto.Hash `json:"hashType"`
	Hash     string      `json:"hash"`
	Bin      string      `json:"bin"`
}

func update(a *AccessSpec, file string, hash crypto.Hash, metadata *meta, ctx accspeccpi.Context, fs vfs.FileSystem) error {
	artifact := a.GetCoordinates()
	err := artifact.SetClassifierExtensionBy(file)
	if err != nil {
		return err
	}
	metadata.Bin = artifact.Url(a.Repository)
	log := log.WithValues("file", metadata.Bin)
	log.Debug("processing")
	metadata.MimeType = artifact.MimeType()
	if hash > 0 {
		metadata.HashType = hash
		metadata.Hash, err = getStringData(ctx, metadata.Bin+hashUrlExt(hash), fs)
		if err != nil {
			return errors.Wrapf(err, "cannot read %s digest of: %s", hash, metadata.Bin)
		}
	} else {
		log.Warn("no digest available")
	}
	return nil
}

func (a *AccessSpec) GetPackageMeta(ctx accspeccpi.Context) (*meta, error) {
	fs := vfsattr.Get(ctx)

	log := log.WithValues("BaseUrl", a.BaseUrl())
	fileMap, err := a.GavFiles(ctx, fs)
	if err != nil {
		return nil, err
	}

	if a.Classifier != "" {
		fileMap = filterByClassifier(fileMap, a.Classifier)
	}

	switch l := len(fileMap); {
	case l <= 0:
		return nil, errors.New("no maven artifact files found")
	case l == 1 && (a.Extension != "" || a.Classifier != ""):
		metadata := meta{}
		for file, hash := range fileMap {
			update(a, file, hash, &metadata, ctx, fs)
		}
		return &metadata, nil
		// default: continue below with: create tempFs where all files can be downloaded to and packed together as tar.gz
	}

	if (a.Extension == "") != (a.Classifier == "") { // XOR
		log.Warn("Either classifier or extension have been specified, which results in an incomplete GAV!")
	}
	tempFs, err := osfs.NewTempFileSystem()
	if err != nil {
		return nil, err
	}
	defer vfs.Cleanup(tempFs)

	metadata := meta{}
	for file, hash := range fileMap {
		update(a, file, hash, &metadata, ctx, fs)

		// download the artifact into the temporary file system
		e := func() error {
			out, err := tempFs.Create(file)
			if err != nil {
				return err
			}
			defer out.Close()
			reader, err := getReader(ctx, metadata.Bin, fs)
			if err != nil {
				return err
			}
			defer reader.Close()
			if hash > 0 {
				dreader := iotools.NewDigestReaderWithHash(hash, reader)
				_, err = io.Copy(out, dreader)
				if err != nil {
					return err
				}
				sum := dreader.Digest().Encoded()
				if metadata.Hash != sum {
					return errors.Newf("%s digest mismatch: expected %s, found %s", metadata.HashType, metadata.Hash, sum)
				}
			} else {
				_, err = io.Copy(out, reader)
				return err
			}
			return err
		}()
		if e != nil {
			return nil, e
		}
	}

	// pack all downloaded files into a tar.gz file
	tgz, err := vfs.TempFile(fs, "", Type+"-"+a.GetCoordinates().FileNamePrefix()+"-*.tar.gz")
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
	for file := range fileMap {
		if !strings.Contains(file, "-"+classifier+".") {
			delete(fileMap, file)
		}
	}
	return fileMap
}

func (a *AccessSpec) GavFiles(ctx accspeccpi.Context, fs ...vfs.FileSystem) (map[string]crypto.Hash, error) {
	if strings.HasPrefix(a.Repository, "file://") {
		dir := path.Join(a.Repository[7:], a.GavPath())
		return gavFilesFromDisk(utils.FileSystem(fs...), dir)
	}
	return a.gavOnlineFiles(ctx)
}

func gavFilesFromDisk(fs vfs.FileSystem, dir string) (map[string]crypto.Hash, error) {
	files, err := tarutils.ListSortedFilesInDir(fs, dir, true)
	if err != nil {
		return nil, err
	}
	return filesAndHashes(files), nil
}

// gavOnlineFiles returns the files of the Maven (mvn) artifact in the repository and their available digests.
func (a *AccessSpec) gavOnlineFiles(ctx accspeccpi.Context) (map[string]crypto.Hash, error) {
	log := log.WithValues("BaseUrl", a.BaseUrl())
	log.Debug("gavOnlineFiles")

	reader, err := getReader(ctx, a.BaseUrl(), nil)
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
	result := make(map[string]crypto.Hash, len(fileList)/2)
	for _, file := range fileList {
		if maven.IsResource(file) {
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

func getStringData(ctx accspeccpi.Context, url string, fs vfs.FileSystem) (string, error) {
	// getStringData reads all data from the given URL and returns it as a string.
	r, err := getReader(ctx, url, fs)
	if err != nil {
		return "", err
	}
	defer r.Close()
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
			return getReader(c.GetContext(), meta.Bin, vfsattr.Get(c.GetContext()))
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

func getReader(ctx accspeccpi.Context, url string, fs vfs.FileSystem) (io.ReadCloser, error) {
	if strings.HasPrefix(url, "file://") {
		path := url[7:]
		return fs.OpenFile(path, vfs.O_RDONLY, 0o600)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	err = identity.BasicAuth(req, ctx, url, "")
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
		if err == nil {
			log.Error("http", "code", resp.Status, "url", url, "body", buf.String())
		}
		return nil, errors.Newf("http %s error - %s", resp.Status, url)
	}
	return resp.Body, nil
}
