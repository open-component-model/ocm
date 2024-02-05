package npmjs

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

func login(registry, username, password string, email string) (string, error) {
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
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return "", err
	}
	return token.Token, nil
}

type Package struct {
	pkg            map[string]interface{}
	Name           string
	Version        string
	Readme         string
	ReadmeFilename string
	GitHead        string
	Description    string
	ID             string
	NodeVersion    string
	NpmVersion     string
	Dist           struct {
		Integrity string `json:"integrity"`
		Shasum    string `json:"shasum"`
		Tarball   string `json:"tarball"`
	}
}

func (p *Package) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.pkg)
}

func (p *Package) MarshalJSON() ([]byte, error) {
	p.pkg["readme"] = p.Readme
	p.pkg["readmeFilename"] = p.ReadmeFilename
	p.pkg["gitHead"] = p.GitHead
	p.pkg["_id"] = p.ID
	p.pkg["_nodeVersion"] = p.NodeVersion
	p.pkg["_npmVersion"] = p.NpmVersion
	p.pkg["dist"] = p.Dist
	return json.Marshal(p.pkg)
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

func createIntegrity(data []byte) string {
	hash := sha512.New()
	hash.Write(data)
	return "sha512-" + base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func createShasum(data []byte) string {
	hash := sha1.New() //nolint:gosec // older npm (prior to v5) uses sha1
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

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
	if len(pkgData) == 0 {
		return nil, fmt.Errorf("package.json is empty")
	}
	var pkg Package
	err = json.Unmarshal(pkgData, &pkg)
	if err != nil {
		return nil, fmt.Errorf("read package.json failed, %w", err)
	}
	var meta struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(pkgData, &meta); err != nil {
		return nil, fmt.Errorf("read package.json version and name failed, %w", err)
	}
	if meta.Name == "" {
		return nil, fmt.Errorf("package.json's name is empty")
	}
	if meta.Version == "" {
		return nil, fmt.Errorf("package.json's version is empty")
	}
	pkg.Version = meta.Version
	pkg.Description = meta.Description
	pkg.Name = meta.Name
	pkg.Readme = string(readme)
	pkg.ReadmeFilename = "README.md"
	pkg.ID = meta.Name + meta.Version
	pkg.Dist.Shasum = createShasum(data)
	pkg.Dist.Integrity = createIntegrity(data)
	return &pkg, nil
}
