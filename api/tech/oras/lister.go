package oras

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2/registry/remote/auth"
)

type OrasLister struct {
	client    *auth.Client
	ref       string
	plainHTTP bool
}

func (c *OrasLister) List(ctx context.Context) ([]string, error) {
	src, err := createRepository(c.ref, c.client, c.plainHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %q: %w", c.ref, err)
	}

	var result []string
	if err := src.Tags(ctx, "", func(tags []string) error {
		result = append(result, tags...)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return result, nil
}
