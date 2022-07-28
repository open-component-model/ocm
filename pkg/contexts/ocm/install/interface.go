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
	"encoding/json"
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

const (
	PathOCM         = "/ocm"
	PathExec        = PathOCM + "/run"
	PathOutputs     = PathOCM + "/outputs"
	PathInputs      = PathOCM + "/inputs"
	InputParameters = "parameters"
	InputConfig     = "config"
	InputOCMRepo    = "ocmrepo"
)

const InstallerSpecificationMimeType = "application/vnd.ocm.gardener.cloud.installer.v1+yaml"

type Driver interface {
	SetConfig(props map[string]string) error
	Exec(op *Operation) (*OperationResult, error)
}

type Executor struct {
	Actions          []string                 `json:"actions,omitempty"`
	ImageResourceRef metav1.ResourceReference `json:"imageResourceRef"`
	Image            *Image                   `json:"image,omitempty"`
	Config           json.RawMessage          `json:"config,omitempty"`
	Outputs          map[string]string        `json:"outputs,omitempty"`
}

type Specification struct {
	Template  json.RawMessage            `json:"configTemplate"`
	Libraries []metav1.ResourceReference `json:"templateLibraries"`
	Scheme    json.RawMessage            `json:"configScheme"`
	Executors []Executor                 `json:"executors"`
}

type Image struct {
	Ref    string `json:"image"`
	Digest string `json:"digest"`
}

// Operation describes the data passed into the driver to run an operation
type Operation struct {
	// Action is the action to be performed. It is passed a srgument to the executable
	Action string
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
