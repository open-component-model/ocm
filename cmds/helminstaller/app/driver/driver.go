package driver

import (
	"io"

	"github.com/mandelsoft/logging"
)

type Config struct {
	ChartPath       string
	Release         string
	Namespace       string
	CreateNamespace bool
	Values          []byte
	Kubeconfig      []byte
	Output          io.Writer
	Debug           logging.Logger
}

type Driver interface {
	Install(*Config) error
	Uninstall(*Config) error
}
