package main

import (
	"ocm.software/ocm/api/credentials"
)

func obfuscate(creds credentials.Credentials) string {
	if creds == nil {
		return "no credentials"
	}
	props := creds.Properties()
	if pw, ok := props[credentials.ATTR_PASSWORD]; ok {
		if len(pw) > 5 {
			pw = pw[:5] + "***"
		} else {
			pw = "***"
		}
		props = props.Copy()
		props[credentials.ATTR_PASSWORD] = pw
	}
	return props.String()
}
