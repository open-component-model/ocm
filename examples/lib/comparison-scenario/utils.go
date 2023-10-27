// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/utils"
)

func PrintPublicKey(ctx ocm.Context, name string) {
	info := signingattr.Get(ctx)
	key := info.GetPublicKey(name)
	if key == nil {
		fmt.Printf("public key for %s not found\n", name)
	} else {
		buf := bytes.NewBuffer(nil)
		err := rsa.WriteKeyData(key, buf)
		if err != nil {
			fmt.Printf("key error: %s\n", err)
		} else {
			fmt.Printf("public key for %s:\n%s\n", name, buf.String())
		}
	}
}

func PrintSignatures(cv ocm.ComponentVersionAccess) {
	fmt.Printf("signatures:\n")
	for i, s := range cv.GetDescriptor().Signatures {
		fmt.Printf("%2d    name: %s\n", i, s.Name)
		fmt.Printf("      digest:\n")
		fmt.Printf("        algorithm:     %s\n", s.Digest.HashAlgorithm)
		fmt.Printf("        normalization: %s\n", s.Digest.NormalisationAlgorithm)
		fmt.Printf("        value:         %s\n", s.Digest.Value)
		fmt.Printf("      signature:\n")
		fmt.Printf("        algorithm: %s\n", s.Signature.Algorithm)
		fmt.Printf("        mediaType: %s\n", s.Signature.MediaType)
		fmt.Printf("        value:     %s\n", s.Signature.Value)
	}
}

func ListFiles(path string, fss ...vfs.FileSystem) ([]string, error) {
	var result []string
	fs := utils.FileSystem(fss...)
	err := vfs.Walk(fs, path, func(path string, info vfs.FileInfo, err error) error {
		result = append(result, path)
		return nil
	})
	return result, err
}
