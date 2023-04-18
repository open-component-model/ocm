// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"net/url"
	"os"
	"strings"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
)

const CONSUMER_TYPE = "Github"

const GITHUB = "github.com"

func init() {
	t := os.Getenv("GITHUB_TOKEN")
	if t != "" {
		host := GITHUB
		port := ""
		us := os.Getenv("GITHUB_SERVER_URL")
		if us != "" {
			u, err := url.Parse(us)
			if err != nil {
				host = u.Host
			}
		}
		if idx := strings.Index(host, ":"); idx > 0 {
			port = host[idx+1:]
			host = host[:idx]
		}

		id := cpi.ConsumerIdentity{
			identity.ID_TYPE:     CONSUMER_TYPE,
			identity.ID_HOSTNAME: host,
		}
		if port != "" {
			id[identity.ID_PORT] = port
		}

		if src, err := cpi.DefaultContext.GetCredentialsForConsumer(id); err != nil || src == nil {
			creds := cpi.NewCredentials(common.Properties{cpi.ATTR_TOKEN: t})
			cpi.DefaultContext.SetCredentialsForConsumer(id, creds)
		}
	}
}
