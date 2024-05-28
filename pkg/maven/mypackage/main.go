// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/open-component-model/ocm/pkg/maven/mypackage/mysubpackage"
)

func main() {
	packageName := mysubpackage.GetPackageName(mysubpackage.SampleFunction)
	fmt.Println("Package name:", packageName)
}
