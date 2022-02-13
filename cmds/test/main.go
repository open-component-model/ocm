package main

import (
	"fmt"

	"github.com/gardener/ocm/pkg/oci"
	_ "github.com/gardener/ocm/pkg/ocm"
)

func main() {
	x, err := oci.ParseRef("host.de:55/image/test@sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a")
	fmt.Printf("%s\n", err)
	fmt.Printf("host %s\n", x.Host)
	fmt.Printf("repo %s\n", x.Repository)
	fmt.Printf("ref %s\n", x.Reference())
	if x.Tag != nil {
		fmt.Printf("tag %s\n", *x.Tag)
	}
	if x.Digest != nil {
		fmt.Printf("digest %s\n", *x.Digest)
	}

}
