package main

import (
	"encoding/base64"
	out "fmt"
	"io"
	"net/http"
	"os"
)

const (
	RequestTokenEnvKey = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	RequestURLEnvKey   = "ACTIONS_ID_TOKEN_REQUEST_URL"
)

type githubActions struct{}

func (ga *githubActions) Enabled() bool {
	if os.Getenv(RequestTokenEnvKey) == "" {
		return false
	}
	if os.Getenv(RequestURLEnvKey) == "" {
		return false
	}
	return true
}

func (ga *githubActions) Get(audience string) error {
	url := os.Getenv(RequestURLEnvKey) + "&audience=" + audience

	//nolint: all // yes
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "bearer "+os.Getenv(RequestTokenEnvKey))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	out.Printf("Token: %s\n", base64.StdEncoding.EncodeToString(data))
	return nil
}

func main() {
	p := &githubActions{}

	if p.Enabled() {
		err := p.Get("sigstore")
		if err != nil {
			out.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	} else {
		out.Printf("no enabled\n")
	}
}
