package main

import (
	"fmt"
	"os"

	"github.com/gardener/ocm/cmds/test/x"
	_ "github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
)

func CheckErr(err error, msg string, args ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s: %s\n ", fmt.Sprintf(msg, args), err)
		os.Exit(1)
	}
}

func C() (err error) {
	defer compdesc.CatchConversionError(&err)
	C1()
	return
}

func C1() {
	compdesc.ThrowConversionError(fmt.Errorf("occured"))
}

func main() {
	x.Vtest()
	data, err := os.ReadFile("component-descriptor.yaml")
	CheckErr(err, "read")
	cd, err := compdesc.Decode(data)
	CheckErr(err, "decode")

	raw, err := cd.RepositoryContexts[0].GetRaw()
	CheckErr(err, "raw ctx")
	fmt.Printf("ctx: %s\n", string(raw))
	_ = cd
	data, err = compdesc.Encode(cd)
	CheckErr(err, "marshal")
	fmt.Printf("%s\n", string(data))

	err = C()
	fmt.Printf("catched error %s\n", err)

	x.DoReflect()
}
