package config

import (
	"encoding/json"
)

type Config struct {
	TransferRepositories TransferRepositories `json:"transferRepositories"`
}

type TransferRepositories struct {
	Types map[string][]string `json:"types,omitempty"`
}

func GetConfig(raw json.RawMessage) (interface{}, error) {
	var cfg Config

	err := json.Unmarshal(raw, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
