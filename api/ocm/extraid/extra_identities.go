package extraid

import (
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

const (
	SystemIdentityName    = metav1.SystemIdentityName
	SystemIdentityVersion = metav1.SystemIdentityVersion

	// ExecutableOperatingSystem is the standard extra id for the operating system for an executable.
	ExecutableOperatingSystem = "os"
	// ExecutableArchitecture is the standard extra id for the architecture system for an executable.
	ExecutableArchitecture = "architecture"
)
