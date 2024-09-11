package consts

import (
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extraid"
)

const (
	// Deprecated: use extraid.SystemIdentityName.
	SystemIdentityName = metav1.SystemIdentityName
	// Deprecated: use extraid.SystemIdentityVersion .
	SystemIdentityVersion = metav1.SystemIdentityVersion

	// Deprecated: use extraid.ExecutableOperatingSystem .
	ExecutableOperatingSystem = extraid.ExecutableOperatingSystem
	// Deprecated: use extraid.ExecutableArchitecture .
	ExecutableArchitecture = extraid.ExecutableArchitecture
)
