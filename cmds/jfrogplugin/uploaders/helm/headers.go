package helm

import (
	"net/http"

	"ocm.software/ocm/api/credentials"
)

func SetHeadersFromCredentials(req *http.Request, creds credentials.Credentials) {
	if creds == nil {
		return
	}
	if creds.ExistsProperty(credentials.ATTR_TOKEN) {
		req.Header.Set("Authorization", "Bearer "+creds.GetProperty(credentials.ATTR_TOKEN))
	} else {
		var user, pass string
		if creds.ExistsProperty(credentials.ATTR_USERNAME) {
			user = creds.GetProperty(credentials.ATTR_USERNAME)
		}
		if creds.ExistsProperty(credentials.ATTR_PASSWORD) {
			pass = creds.GetProperty(credentials.ATTR_PASSWORD)
		}
		req.SetBasicAuth(user, pass)
	}
}
