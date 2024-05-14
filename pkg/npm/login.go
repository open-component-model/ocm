package npm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/npm/identity"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/logging"
)

var REALM = identity.REALM

// Login to npm registry (URL) and retrieve bearer token.
func Login(registry string, username string, password string, email string) (string, error) {
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

// BearerToken retrieves the bearer token for the given repository URL and package name.
// Either it's setup in the credentials or it will login to the registry and retrieve it.
func BearerToken(ctx cpi.ContextProvider, repoUrl string, pkgName string) (string, error) {
	// get credentials and TODO cache it
	cred := identity.GetCredentials(ctx, repoUrl, pkgName)
	if cred == nil {
		return "", fmt.Errorf("no credentials found for %s. Couldn't upload '%s'", repoUrl, pkgName)
	}
	log := logging.Context().Logger(identity.REALM)
	log.Debug("found credentials")

	// check if token exists, if not login and retrieve token
	token := cred[identity.ATTR_TOKEN]
	if token != "" {
		log.Debug("token found, skipping login")
		return token, nil
	}

	// use user+pass+mail from credentials to login and retrieve bearer token
	username := cred[identity.ATTR_USERNAME]
	password := cred[identity.ATTR_PASSWORD]
	email := cred[identity.ATTR_EMAIL]
	if username == "" || password == "" || email == "" {
		return "", fmt.Errorf("credentials for %s are invalid. Username, password or email missing! Couldn't upload '%s'", repoUrl, pkgName)
	}
	log = log.WithValues("user", username, "repo", repoUrl)
	log.Debug("login")

	// TODO: check different kinds of .npmrc content
	return Login(repoUrl, username, password, email)
}

// Authorize the given request with the bearer token for the given repository URL and package name.
// If the token is empty (login failed or credentials not found), it will not be set.
func Authorize(req *http.Request, ctx cpi.ContextProvider, repoUrl string, pkgName string) {
	token, err := BearerToken(ctx, repoUrl, pkgName)
	if err != nil {
		log := logging.Context().Logger(identity.REALM)
		log.Debug("Couldn't authorize", "error", err.Error(), "repo", repoUrl, "package", pkgName)
	} else if token != "" {
		req.Header.Set("authorization", "Bearer "+token)
	}
}
