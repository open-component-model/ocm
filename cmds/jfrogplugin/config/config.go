package config

import (
	"encoding/json"
	"fmt"
)

type Config struct{}

// GetConfig returns the config from the raw json message.
// any return is required for the plugin interface.
func GetConfig(raw json.RawMessage) (interface{}, error) {
	var cfg Config

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("could not get config: %w", err)
	}
	return &cfg, nil
}
