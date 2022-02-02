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

package format

import (
	"time"
)

const (
	DirMode  = 0755
	FileMode = 0644
)

var (
	ModTime = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	// The artefact descriptor name for artefact format
	ArtefactSetDescriptorFileName = "artefact-descriptor.json"
	// BlobsDirectoryName is the name of the directory holding the artefact archives
	BlobsDirectoryName = "blobs"
	// ArtefactIndexFileName is the artefact index descriptor name for CommanTransportFormat
	ArtefactIndexFileName = "artefact-index.json"
	// ArtefactsDirectoryName is the name of the directory holding the artefact archives
	ArtefactsDirectoryName = "artefacts"
)
