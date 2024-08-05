package mavenaccess

import "ocm.software/ocm/api/tech/maven"

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
