package maven

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/ioutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"golang.org/x/net/html"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/tarutils"
)

type FileMeta struct {
	MimeType string
	HashType crypto.Hash
	Hash     string
	Location *Location
}

type Repository struct {
	Location
}

func NewFileRepository(path string, fss ...vfs.FileSystem) *Repository {
	return &Repository{Location{
		path: path,
		fs:   utils.FileSystem(fss...),
	}}
}

func NewUrlRepository(repoUrl string, fss ...vfs.FileSystem) (*Repository, error) {
	u, err := url.ParseRequestURI(repoUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "file" {
		if u.Host != "" && u.Host != "localhost" {
			return nil, errors.Newf("named host not supported for url file scheme: %q", repoUrl)
		}
		return NewFileRepository(u.Path, fss...), nil
	}
	return &Repository{Location{
		url: repoUrl,
	}}, nil
}

func (r *Repository) Url() (string, error) {
	if r.url != "" {
		return r.url, nil
	}
	p, err := vfs.Canonical(r.fs, r.path, false)
	if err != nil {
		return "", err
	}
	return "file://localhost" + p, nil
}

// Body is the response struct of a deployment from the MVN repository (JFrog Artifactory).
type Body struct {
	Repo        string            `json:"repo"`
	Path        string            `json:"path"`
	DownloadUri string            `json:"downloadUri"`
	Uri         string            `json:"uri"`
	MimeType    string            `json:"mimeType"`
	Size        string            `json:"size"`
	Checksums   map[string]string `json:"checksums"`
}

func (r *Repository) Download(coords *Coordinates, creds Credentials, enforceVerification ...bool) (io.ReadCloser, error) {
	files, err := r.GavFiles(coords, creds)
	if err != nil {
		return nil, err
	}
	algorithm, ok := files[coords.FileName()]
	if !ok {
		return nil, errors.ErrNotFound("file", coords.FileName(), coords.GAV())
	}

	var digest string
	loc := coords.Location(r)
	if algorithm != 0 {
		digestFile := loc.AddExtension(HashExt(algorithm))
		reader, err := digestFile.GetReader(creds)
		if err != nil {
			return nil, err
		}
		digestData, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		digest = string(digestData)
	} else {
		if general.Optional(enforceVerification...) {
			return nil, fmt.Errorf("unable to verify, no digest available in target repository")
		}
	}

	reader, err := loc.GetReader(creds)
	if err != nil {
		return nil, err
	}
	if algorithm != 0 {
		reader = iotools.VerifyingReaderWithHash(reader, algorithm, digest)
	}
	return reader, nil
}

func (r *Repository) Upload(coords *Coordinates, reader ioutils.DupReadCloser, creds Credentials, hashes iotools.Hashes) (rerr error) {
	finalize := finalizer.Finalizer{}
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	loc := coords.Location(r)
	if r.IsFileSystem() {
		err := loc.fs.MkdirAll(vfs.Dir(loc.fs, loc.path), 0o755)
		if err != nil {
			return err
		}
		f, err := loc.fs.OpenFile(loc.path, vfs.O_WRONLY|vfs.O_CREATE|vfs.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		finalize.Close(f)

		_, err = io.Copy(f, reader)
		if err != nil {
			return err
		}

		for algorithm := range hashes {
			digest := hashes.GetString(algorithm)
			p := loc.path + "." + HashExt(algorithm)
			err = vfs.WriteFile(loc.fs, p, []byte(digest), 0o644)
			if err != nil {
				return err
			}
		}
		return nil
	}
	reader, err := reader.Dup()
	if rerr != nil {
		return err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, loc.String(), reader)
	if err != nil {
		return err
	}
	if creds != nil {
		err = creds.SetForRequest(req)
		if err != nil {
			return err
		}
	}
	// give the remote server a chance to decide based upon the checksum policy
	for k, v := range hashes.AsHttpHeader() {
		req.Header[k] = v
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	finalize.Close(resp.Body)

	// Check the response
	if resp.StatusCode != http.StatusCreated {
		all, e := io.ReadAll(resp.Body)
		if e != nil {
			return e
		}
		return fmt.Errorf("http (%d) - failed to upload coords: %s", resp.StatusCode, string(all))
	}
	Log.Debug("uploaded", "coords", coords, "extension", coords.Extension, "classifier", coords.Classifier)

	// Validate the response - especially the hash values with the ones we've tried to send
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var artifactBody Body
	err = json.Unmarshal(respBody, &artifactBody)
	if err != nil {
		return err
	}

	algorithm := bestAvailableHash(slices.Collect(maps.Keys(hashes)))
	digest := hashes.GetString(algorithm)
	remoteDigest := artifactBody.Checksums[strings.ReplaceAll(strings.ToLower(algorithm.String()), "-", "")]
	if remoteDigest == "" {
		Log.Warn("no checksum found for algorithm, we can't guarantee that the coords has been uploaded correctly", "algorithm", algorithm.String())
	} else if remoteDigest != digest {
		return errors.New("failed to upload coords: checksums do not match")
	}
	Log.Debug("digests are ok", "remoteDigest", remoteDigest, "digest", digest)
	return err
}

func (r *Repository) GetFileMeta(c *Coordinates, file string, hash crypto.Hash, creds Credentials) (*FileMeta, error) {
	coords := c.Copy()
	err := coords.SetClassifierExtensionBy(file)
	if err != nil {
		return nil, err
	}
	metadata := &FileMeta{
		Location: coords.Location(r),
		MimeType: coords.MimeType(),
	}
	log := Log.WithValues("file", metadata.Location.String())
	log.Debug("processing")
	if hash > 0 {
		metadata.HashType = hash
		metadata.Hash, err = metadata.Location.GetHash(creds, hash)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read %s digest of: %s", hash, metadata.Location)
		}
	} else {
		log.Warn("no digest available")
	}
	return metadata, nil
}

func (r *Repository) GavFiles(coords *Coordinates, creds Credentials) (map[string]crypto.Hash, error) {
	if r.path != "" {
		return gavFilesFromDisk(r.fs, coords.GavLocation(r).path)
	}
	return gavOnlineFiles(r, coords, creds)
}

func gavFilesFromDisk(fs vfs.FileSystem, dir string) (map[string]crypto.Hash, error) {
	files, err := tarutils.ListSortedFilesInDir(fs, dir, true)
	if err != nil {
		return nil, err
	}
	return filesAndHashes(files), nil
}

// gavOnlineFiles returns the files of the Maven artifact in the repository and their available digests.
func gavOnlineFiles(repo *Repository, coords *Coordinates, creds Credentials) (map[string]crypto.Hash, error) {
	log := Log.WithValues("RepoUrl", repo.String(), "GAV", coords.GavPath())
	log.Debug("gavOnlineFiles")

	tweakUrlAndUserAgent := func(loc *Location, req *http.Request) {
		if loc != nil && !strings.HasSuffix(loc.url, "/") {
			loc.url += "/"
		}
		if req != nil {
			req.Header.Set("User-Agent", "Mozilla")
		}
	}

	loc := coords.GavLocation(repo)
	reader, err := loc.GetReader(creds, tweakUrlAndUserAgent)
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
	prefix := coords.FileNamePrefix()
	process = func(node *html.Node) {
		// check if the node is an element node and the tag is "<a href="..." />"
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attribute := range node.Attr {
				if attribute.Key == "href" {
					attribute.Val = strings.TrimPrefix(attribute.Val, loc.String()) // make the href relative
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
	// Which hash files are available?
	result := make(map[string]crypto.Hash, len(fileList)/2)
	for _, file := range fileList {
		if IsResource(file) {
			result[file] = bestAvailableHashForFile(fileList, file)
			log.Debug("found", "file", file)
		}
	}
	return result
}

type Location struct {
	url  string
	path string
	fs   vfs.FileSystem
}

func (l *Location) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

func (l *Location) IsFileSystem() bool {
	return l.path != ""
}

func (l *Location) AddPath(path string) *Location {
	result := *l
	var p *string
	if result.url != "" {
		p = &result.url
	} else {
		p = &result.path
	}

	if !strings.HasSuffix(*p, "/") {
		*p += "/"
	}
	*p += path
	return &result
}

func (l *Location) AddExtension(ext string) *Location {
	result := *l
	var p *string
	if result.url != "" {
		p = &result.url
	} else {
		p = &result.path
	}

	*p += "." + ext
	return &result
}

func (l *Location) String() string {
	return general.Conditional(l.path != "", l.path, l.url)
}

func (l *Location) GetHash(creds Credentials, hash crypto.Hash) (string, error) {
	// getStringData reads all data from the given URL and returns it as a string.
	r, err := l.AddExtension(HashExt(hash)).GetReader(creds)
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

func (l *Location) GetReader(creds Credentials, tweakIndexOf ...func(loc *Location, req *http.Request)) (io.ReadCloser, error) {
	if l.path != "" {
		return l.fs.OpenFile(l.path, vfs.O_RDONLY, 0o600)
	}

	if tweakIndexOf != nil {
		tweakIndexOf[0](l, nil) // tweak the URL if necessary
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, l.url, nil)
	if err != nil {
		return nil, err
	}
	if creds != nil {
		err = creds.SetForRequest(req)
		if err != nil {
			return nil, err
		}
	}
	if tweakIndexOf != nil {
		tweakIndexOf[0](nil, req) // tweak the request if necessary
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
			Log.Error("http", "code", resp.Status, "repo", l.url, "body", buf.String())
		}
		return nil, errors.Newf("http %s error - %s", resp.Status, l.url)
	}
	return resp.Body, nil
}

type Credentials interface {
	SetForRequest(req *http.Request) error
}

type BasicAuthCredentials struct {
	Username string
	Password string
}

func (b *BasicAuthCredentials) SetForRequest(req *http.Request) error {
	req.SetBasicAuth(b.Username, b.Password)
	return nil
}
