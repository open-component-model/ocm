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

package install

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

const (
	PathTOI         = "/toi"
	Inputs          = "inputs"
	Outputs         = "outputs"
	PathExec        = PathTOI + "/run"
	PathOutputs     = PathTOI + "/" + Outputs
	PathInputs      = PathTOI + "/" + Inputs
	InputParameters = "parameters"
	InputConfig     = "config"
	InputOCMConfig  = "ocmconfig"
	InputOCMRepo    = "ocmrepo"
)

type Driver interface {
	SetConfig(props map[string]string) error
	Exec(op *Operation) (*OperationResult, error)
}

// Operation describes the data passed into the driver to run an operation
type Operation struct {
	// Action is the action to be performed. It is passed a srgument to the executable
	Action string
	// ComponentVersion is the name of the root component/version to install
	ComponentVersion string
	// Image is the image to invoke
	Image Image
	// Environment contains environment variables that should be injected into the container execution
	Environment map[string]string
	// Files contains files that should be injected into the invocation image.
	Files map[string]accessio.BlobAccess
	// Outputs map of output (sub)paths (e.g. NAME) to the name of the output.
	// Indicates which outputs the driver should return the contents of in the OperationResult.
	Outputs map[string]string
	// Output stream for log messages from the driver
	Out io.Writer
	// Output stream for error messages from the driver
	Err io.Writer
}

// OperationResult is the output of the Driver running an Operation.
type OperationResult struct {
	// Outputs maps from the name of the output to its content.
	Outputs map[string][]byte

	// Error is any errors from executing the operation.
	Error error
}
