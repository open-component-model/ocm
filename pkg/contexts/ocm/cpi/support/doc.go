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

/*
Package support provides a standard implementation for the object type set
required to implement the OCM repository interface.

This implementation is based on three interfaces that have to implemented:

- BlobContainer
  is used to provide access to blob data
- ComponentVersionContainer
  is used to provide access to component version for  component.

The function NewComponentVersionAccess can be used to create an
object implementing the complete ComponentVersionAccess contract.
*/
package support
