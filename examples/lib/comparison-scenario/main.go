package main

import (
	"fmt"
	"os"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/utils/template"
)

const (
	COMPONENT_NAME    = "acme.org/podinfo"
	COMPONENT_VERSION = "0.1.0"
)

// CFG is the path to the file containing the credentials
var CFG = "config/config.yaml"

func main() {
	arg := 1
	if len(os.Args) > 1 {
		if os.Args[1] == "--config" {
			if len(os.Args) > 2 {
				CFG = os.Args[2]
				arg = 3
			} else {
				fmt.Fprintf(os.Stderr, "error: config file missing\n")
				os.Exit(1)
			}
		}
	}
	cfg, err := helper.ReadConfig(CFG)
	if err == nil {
		cmd := "create"

		if len(os.Args) > arg {
			cmd = os.Args[arg]
		}
		switch cmd {
		case "create":
			err = Create(cfg)
		case "sign":
			err = Sign(cfg)
		case "write":
			err = Write(cfg)
		case "transport":
			err = Transport(cfg)
		case "verify":
			err = Verify(cfg)
		case "download":
			err = Download(cfg)
		case "getref":
			err = GetRef(cfg)
		case "deployscript":
			err = GetDeployScript(cfg)

		case "provider":
			err = Provider(cfg)
		case "consumer":
			err = Consumer(cfg)
		case "localize":
			err = Localize(cfg, template.Values{
				"ui": map[string]interface{}{
					"color": "#1010c0",
				},
			}, "", "")
		case "deploy":
			err = Localize(cfg, template.Values{
				"ui": map[string]interface{}{
					"color": "#1010c0",
				},
			}, "podinfo", "default")

		default:
			err = fmt.Errorf("unknown scenario %q", cmd)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
