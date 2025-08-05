package config

import (
	cfgcpi "ocm.software/ocm/api/config/cpi"
)

// GetConfigured applies config objects of a config context
// to a configuration struct of type T.
// A pointer to the configured struct is returned.
// Attention: T must be a struct type.
func GetConfigured[T any](ctxp ContextProvider) (*T, error) {
	var c T
	err := cfgcpi.NewUpdater(ctxp.ConfigContext(), &c).Update()
	if err != nil {
		return nil, err
	}
	return &c, nil
}
