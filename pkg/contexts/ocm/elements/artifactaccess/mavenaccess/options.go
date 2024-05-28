package mavenaccess

import "github.com/open-component-model/ocm/pkg/maven"

type (
	Options = maven.Coordinates
	Option  = maven.CoordinateOption
)

type WithClassifier = maven.WithClassifier

func WithOptionalClassifier(c *string) Option {
	return maven.WithOptionalClassifier(c)
}

type WithExtension = maven.WithExtension

func WithOptionalExtension(e *string) Option {
	return maven.WithOptionalExtension(e)
}
