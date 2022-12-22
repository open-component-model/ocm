package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/open-component-model/ocm/pkg/version"
)

func main() {
	if len(os.Args) < 1 {
		log.Fatal("missing argument")
	}

	ver := semver.MustParse(version.ReleaseVersion)

	cmd := os.Args[1]

	switch cmd {
	case "print-version":
		fmt.Println(ver)
	}
}
