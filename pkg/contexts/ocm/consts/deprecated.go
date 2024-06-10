package consts

import (
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/extraid"
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
