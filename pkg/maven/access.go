// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maven

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"github.com/cloudflare/cfssl/log"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Repository struct {
	Location
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

func NewFileRepository(path string, fss ...vfs.FileSystem) *Repository {
	return &Repository{Location{
		path: path,
		fs:   utils.FileSystem(fss...),
	}}
}

func NewUrlRepository(repoUrl string, fss ...vfs.FileSystem) (*Repository, error) {
	u, err := url.Parse(repoUrl)
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

type FileMeta struct {
	MimeType string
	HashType crypto.Hash
	Hash     string
	Location *Location
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

func (l *Location) GetReader(creds Credentials) (io.ReadCloser, error) {
	if l.path != "" {
		return l.fs.OpenFile(l.path, vfs.O_RDONLY, 0o600)
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

// gavOnlineFiles returns the files of the Maven (mvn) artifact in the repository and their available digests.
func gavOnlineFiles(repo *Repository, coords *Coordinates, creds Credentials) (map[string]crypto.Hash, error) {
	log := Log.WithValues("RepoUrl", repo.String(), "GAV", coords.GavPath())
	log.Debug("gavOnlineFiles")

	reader, err := coords.GavLocation(repo).GetReader(creds)
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
			result[file] = bestAvailableHash(fileList, file)
			log.Debug("found", "file", file)
		}
	}
	return result
}
