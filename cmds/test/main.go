package main

import (
	"fmt"
	"os"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	_ "github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
)

func CheckErr(err error, msg string, args ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s: %s\n ", fmt.Sprintf(msg, args), err)
		os.Exit(1)
	}
}

func main() {
	data, err := os.ReadFile("component-descriptor.yaml")
	CheckErr(err, "read")
	cd, err := compdesc.Decode(data)
	CheckErr(err, "decode")

	raw, err := cd.RepositoryContexts[0].GetRaw()
	CheckErr(err, "raw ctx")
	fmt.Printf("ctx: %s\n", string(raw))
	_ = cd
	data, err = compdesc.Encode(cd, v2.SchemaVersion, compdesc.DefaultYAMLCodec)
	CheckErr(err, "marshal")
	fmt.Printf("%s\n", string(data))
}
