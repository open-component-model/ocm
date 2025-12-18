package docker

import (
	"github.com/moby/moby/api/types/container"
)

func NetworkModeOpt(mode string) ConfigurationOption {
	return func(_ *container.Config, h *container.HostConfig) error {
		h.NetworkMode = container.NetworkMode(mode)
		return nil
	}
}

func UsernsModeOpt(mode string) ConfigurationOption {
	return func(_ *container.Config, h *container.HostConfig) error {
		h.UsernsMode = container.UsernsMode(mode)
		return nil
	}
}
