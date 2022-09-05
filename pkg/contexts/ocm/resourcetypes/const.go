// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package resourcetypes

const (
	// OCI_ARTEFACT describes a generic OCI artefact following the
	//   [open containers image specification](https://github.com/opencontainers/image-spec/blob/main/spec.md).
	OCI_ARTEFACT = "ociArtefact"
	// OCI_IMAGE describes an OCIArtefact containing an image.
	OCI_IMAGE = "ociImage"
	// HELM_CHART describes a helm chart, either stored as OCI artefact or as tar
	// blob (tar media type).
	HELM_CHART = "helmChart"
	// BLOB describes any anonymous untyped blob data.
	BLOB = "blob"
	// FILESYSTEM describes a directory structure stored as archive (tar, tgz).
	FILESYSTEM = "`filesytem"
)
