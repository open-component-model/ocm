package support

import "fmt"

type UpdateComponentVersionContainerError struct {
	Name    string
	Version string
	Type    string

	Original error
}

func (e UpdateComponentVersionContainerError) Error() string {
	message := fmt.Sprintf(
		"unable to update '%s:%s' base component container",
		e.Name,
		e.Version,
	)

	if e.Original != nil {
		message = fmt.Sprintf("%s: %s", message, e.Original.Error())
	}

	return message
}

func (e UpdateComponentVersionContainerError) Unwrap() error {
	return e.Original
}

type AccessCheckError struct {
	Name    string
	Version string
	Type    string

	Original error
}

func (e AccessCheckError) Error() string {
	message := fmt.Sprintf(
		"failed access spec check on '%s:%s' with type '%s'",
		e.Name,
		e.Version,
		e.Type,
	)

	if e.Original != nil {
		message = fmt.Sprintf("%s: %s", message, e.Original.Error())
	}

	return message
}

func (e AccessCheckError) Unwrap() error {
	return e.Original
}
