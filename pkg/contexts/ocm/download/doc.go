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

// Package download provides an API for resource download handlers.
// A download handler is used for downloading resoures. By default the native
// blob as provided by the access method is the resukt of a download.
// A download handler can influence the outbound blob format according
// to the concrete type of the resource.
// For example, a helm download for a helm artefact stored as oci artefact
// will not provide the oci format chosen for representing the artefact
// in OCI but a regular helm archive according to its specification.
// The sub package handlers provides dedicated packages for standard handlers.
//
// A downloader registry is stores as attribute ATTR_DOWNLOADER_HANDLERS
// for the OCM context, it is not a static part of the OCM context.
// The downloaders are basically used by clients requiring access
// to the effective resource content.
package download
