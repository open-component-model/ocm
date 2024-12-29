package maven

import (
	metav1 "ocm.software/ocm/api/ocm/refhints"
)

const ReferenceHintType = "maven"

// HINT_REFERENCE is the single attribute describing the OCI reference.
// for OCI hints.
const HINT_REFERENCE = metav1.HINT_REFERENCE

func ReferenceHint(ref string, implicit ...bool) metav1.ReferenceHint {
	return metav1.New(ReferenceHintType, ref, implicit...)
}
