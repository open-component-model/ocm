package artifactset

import "fmt"

type GetArtifactError struct {
	Original error
	Ref      string
}

func (e GetArtifactError) Error() string {
	message := fmt.Sprintf("failed to get artifact '%s'", e.Ref)

	if e.Original != nil {
		message = fmt.Sprintf("%s: %s", message, e.Original.Error())
	}

	return message
}

func (e GetArtifactError) Unwrap() error {
	return e.Original
}
