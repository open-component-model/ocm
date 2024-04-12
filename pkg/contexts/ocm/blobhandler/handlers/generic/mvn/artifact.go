package mvn

import "strings"

// Artifact holds the typical Maven coordinates groupId, artifactId, version and packaging.
type Artifact struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Packaging  string `xml:"packaging"`
}

// GAV returns the GAV coordinates of the Maven Artifact.
func (a *Artifact) GAV() string {
	return a.GroupId + ":" + a.ArtifactId + ":" + a.Version
}

// Path returns the Maven Artifact's path within a repository.
func (a *Artifact) Path() string {
	return a.GroupPath() + "/" + a.ArtifactId + "/" + a.Version + "/" + a.ArtifactId + "-" + a.Version + "." + a.Packaging
}

// GroupPath returns GroupId with `/` instead of `.`.
func (a *Artifact) GroupPath() string {
	return strings.ReplaceAll(a.GroupId, ".", "/")
}

// Purl returns the Package URL of the Maven Artifact.
func (a *Artifact) Purl() string {
	return "pkg:maven/" + a.GroupId + "/" + a.ArtifactId + "@" + a.Version
}

// Body is the response struct of a deployment from the MVN repository (JFrog Artifactory).
type Body struct {
	Repo        string `json:"repo"`
	Path        string `json:"path"`
	DownloadUri string `json:"downloadUri"`
	Uri         string `json:"uri"`
	MimeType    string `json:"mimeType"`
	Size        string `json:"size"`
	Checksums   struct {
		Md5    string `json:"md5"`
		Sha1   string `json:"sha1"`
		Sha256 string `json:"sha256"`
		Sha512 string `json:"sha512"`
	} `json:"checksums"`
}
