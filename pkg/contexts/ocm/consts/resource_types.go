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

package consts

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

const (
	// OCIArtefact describes a generic OCI artefact follwoing the
	//   [open containers image specification](https://github.com/opencontainers/image-spec/blob/main/spec.md)
	OCIArtefact = resourcetypes.OCI_ARTEFACT
	// OCIImage describes an OCIArtefact containing an image
	OCIImage = resourcetypes.OCI_IMAGE
	// HelmChart describes a helm chart, either stored as OCI artefact or as tar blob (tar media type)
	HelmChart = resourcetypes.HELM_CHART
	// blob describes any anonymous untyped blob data
	Blob = resourcetypes.BLOB
	// FileSystemContent describes a directory structure stored as archive (tar, tgz)
	FileSystemContent = resourcetypes.FILESYSTEM_CONTENT
)
