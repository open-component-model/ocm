// inspired by @cloverstd - https://gist.github.com/cloverstd/7355e95424d59256123a1093f76f78a6

package npm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	//nolint:gosec // older npm (prior to v5) uses sha1
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Package struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Readme      string `json:"readme"`
	Description string `json:"description"`
	Dist        struct {
		Integrity string `json:"integrity"`
		Shasum    string `json:"shasum"`
		Tarball   string `json:"tarball"`
	} `json:"dist"`
}

type Attachment struct {
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
	Length      int    `json:"length"`
}

type Body struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DistTags    struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
	Versions    map[string]*Package    `json:"versions"`
	Readme      string                 `json:"readme"`
	Attachments map[string]*Attachment `json:"_attachments"`
}

func NewAttachment(data []byte) *Attachment {
	return &Attachment{
		ContentType: "application/octet-stream",
		Data:        data,
		Length:      len(data),
	}
}

// Login to npm registry (URL) and retrieve bearer token.
func login(registry string, username string, password string, email string) (string, error) {
	data := map[string]interface{}{
		"_id":      "org.couchdb.user:" + username,
		"name":     username,
		"email":    email,
		"password": password,
		"type":     "user",
	}
	marshal, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, registry+"/-/user/org.couchdb.user:"+url.PathEscape(username), bytes.NewReader(marshal))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		all, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("%d, %s", resp.StatusCode, string(all))
	}
	var token struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}
	return token.Token, nil
}

func createSha512(data []byte) string {
	hash := sha512.New()
	hash.Write(data)
	return "sha512-" + base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func createSha1(data []byte) string {
	hash := sha1.New() //nolint:gosec // older npm (prior to v5) uses sha1
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// Read package.json and README.md from tarball to create Package object.
func prepare(data []byte) (*Package, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	var (
		pkgData []byte
		readme  []byte
	)
	for {
		thr, e := tr.Next()
		if e != nil {
			if errors.Is(e, io.EOF) {
				break
			}
			return nil, e
		}
		if pkgData != nil && readme != nil {
			break
		}
		switch thr.Name {
		case "package/package.json":
			pkgData, err = io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("read package.json failed, %w", err)
			}
		case "package/README.md":
			readme, err = io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("read README.md failed, %w", err)
			}
		}
	}

	// fetch some information from package.json
	if len(pkgData) == 0 {
		return nil, fmt.Errorf("package.json is empty")
	}
	var pkgJson map[string]interface{}
	err = json.Unmarshal(pkgData, &pkgJson)
	if err != nil {
		return nil, fmt.Errorf("read package.json failed, %w", err)
	}
	if pkgJson["name"] == "" {
		return nil, fmt.Errorf("package.json's name is empty")
	}
	if pkgJson["version"] == "" {
		return nil, fmt.Errorf("package.json's version is empty")
	}

	// create package object
	var pkg Package
	pkg.Name = pkgJson["name"].(string)
	pkg.Version = pkgJson["version"].(string)
	pkg.Description = pkgJson["description"].(string)
	pkg.Readme = string(readme)
	pkg.Dist.Shasum = createSha1(data)
	pkg.Dist.Integrity = createSha512(data)
	return &pkg, nil
}
