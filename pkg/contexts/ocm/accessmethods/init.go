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

package accessmethods

import (
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localociblob"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/none"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartefact"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/s3"
)
