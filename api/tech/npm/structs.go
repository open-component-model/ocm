package npm

type Dist struct {
	Integrity string `json:"integrity"`
	Shasum    string `json:"shasum"`
	Tarball   string `json:"tarball"`
}

type Version struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Dist    Dist   `json:"dist"`
}

type Project struct {
	Name    string             `json:"name"`
	Version map[string]Version `json:"versions"`
}
