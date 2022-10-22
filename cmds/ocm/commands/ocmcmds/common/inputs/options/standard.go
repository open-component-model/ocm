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

package options

import (
	"github.com/open-component-model/ocm/pkg/clisupport"
)

var PathOption = clisupport.NewStringOptionType("inputPath", "path field for input")
var MediaTypeOption = clisupport.NewStringOptionType("inputMediatype", "media type for input")
var CompressOption = clisupport.NewBoolOptionType("inputCompress", "compress option for input")

var ExcludeOption = clisupport.NewStringArrayOptionType("inputExcludes", "excludes (path) for inputs")
var IncludeOption = clisupport.NewStringArrayOptionType("inputIncludes", "includes (path) for inputs")

var PreserveDirOption = clisupport.NewBoolOptionType("inputPreserveDir", "preserve directory in archive for inputs")
var FollowSymlinksOption = clisupport.NewBoolOptionType("inputFollowSymlinks", "follow symbolic links during archive creation for inputs")

var HintOption = clisupport.NewStringOptionType("inputHint", "(repository) hint local artifacts for inputs")
var VariantsOption = clisupport.NewStringArrayOptionType("inputVariants", "(platform) variants for inputs")

var LibrariesOption = clisupport.NewStringArrayOptionType("inputLibraries", "library path for inputs")

var VersionOption = clisupport.NewStringArrayOptionType("inputVersion", "version info for inputs")

var ValuesOption = clisupport.NewValueMapOptionType("inputValues", "YAML based generic values for inputs")
