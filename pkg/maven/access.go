// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maven

import (
	"bytes"
	"context"
	"crypto"
	"github.com/cloudflare/cfssl/log"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"path"
	"strings"
)

type FileMeta struct {
	MimeType string
	HashType crypto.Hash
	Hash     string
	Url      string
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

func GetHash(url string, creds Credentials, hash crypto.Hash, fss ...vfs.FileSystem) (string, error) {
	// getStringData reads all data from the given URL and returns it as a string.
	r, err := GetReader(url+HashUrlExt(hash), creds, fss...)
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

func GetReader(url string, creds Credentials, fss ...vfs.FileSystem) (io.ReadCloser, error) {
	if strings.HasPrefix(url, "file://") {
		fs := utils.FileSystem(fss...)
		path := url[7:]
		return fs.OpenFile(path, vfs.O_RDONLY, 0o600)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
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
			Log.Error("http", "code", resp.Status, "url", url, "body", buf.String())
		}
		return nil, errors.Newf("http %s error - %s", resp.Status, url)
	}
	return resp.Body, nil
}

func GetFileMeta(repoUrl string, c *Coordinates, file string, hash crypto.Hash, creds Credentials, fss ...vfs.FileSystem) (*FileMeta, error) {
	coords := c.Copy()
	err := coords.SetClassifierExtensionBy(file)
	if err != nil {
		return nil, err
	}
	metadata := &FileMeta{
		Url:      coords.Url(repoUrl),
		MimeType: coords.MimeType(),
	}
	log := Log.WithValues("file", metadata.Url)
	log.Debug("processing")
	if hash > 0 {
		metadata.HashType = hash
		metadata.Hash, err = GetHash(metadata.Url, creds, hash, fss...)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read %s digest of: %s", hash, metadata.Url)
		}
	} else {
		log.Warn("no digest available")
	}
	return metadata, nil
}

func GavFiles(repoUrl string, coords *Coordinates, creds Credentials, fss ...vfs.FileSystem) (map[string]crypto.Hash, error) {
	if strings.HasPrefix(repoUrl, "file://") {
		dir := path.Join(repoUrl[7:], coords.GavPath())
		return gavFilesFromDisk(utils.FileSystem(fss...), dir)
	}
	return gavOnlineFiles(repoUrl, coords, creds)
}

func gavFilesFromDisk(fs vfs.FileSystem, dir string) (map[string]crypto.Hash, error) {
	files, err := tarutils.ListSortedFilesInDir(fs, dir, true)
	if err != nil {
		return nil, err
	}
	return filesAndHashes(files), nil
}

// gavOnlineFiles returns the files of the Maven (mvn) artifact in the repository and their available digests.
func gavOnlineFiles(repoUrl string, coords *Coordinates, creds Credentials) (map[string]crypto.Hash, error) {
	log := Log.WithValues("BaseUrl", repoUrl)
	log.Debug("gavOnlineFiles")

	reader, err := GetReader(coords.GavUrl(repoUrl), creds, nil)
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
