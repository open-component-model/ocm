package internal

import (
	"github.com/mandelsoft/goutils/errors"
)

const (
	KIND_CREDENTIALS = "credentials"
	KIND_CONSUMER    = "consumer"
	KIND_REPOSITORY  = "repository"
)

func ErrUnknownCredentials(name string) error {
	return errors.ErrUnknown(KIND_CREDENTIALS, name)
}

func ErrUnknownConsumer(name string) error {
	return errors.ErrUnknown(KIND_CONSUMER, name)
}

func ErrUnknownRepository(kind, name string) error {
	return errors.ErrUnknown(KIND_REPOSITORY, name, kind)
}
