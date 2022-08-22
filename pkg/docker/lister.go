// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package docker

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/log"
	"github.com/pkg/errors"

	"github.com/open-component-model/ocm/pkg/docker/resolve"
)

var ErrObjectNotRequired = errors.New("object not required")

type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type dockerLister struct {
	dockerBase *dockerBase
}

func (r *dockerResolver) Lister(ctx context.Context, ref string) (resolve.Lister, error) {
	base, err := r.resolveDockerBase(ref)
	if err != nil {
		return nil, err
	}
	if base.refspec.Object != "" {
		return nil, ErrObjectNotRequired
	}

	return &dockerLister{
		dockerBase: base,
	}, nil
}

func (r *dockerLister) List(ctx context.Context) ([]string, error) {
	refspec := r.dockerBase.refspec
	base := r.dockerBase
	var (
		firstErr error
		paths    [][]string
		caps     = HostCapabilityPull
	)

	// turns out, we have a valid digest, make a url.
	paths = append(paths, []string{"tags/list"})
	caps |= HostCapabilityResolve

	hosts := base.filterHosts(caps)
	if len(hosts) == 0 {
		return nil, errors.Wrap(errdefs.ErrNotFound, "no list hosts")
	}

	ctx, err := ContextWithRepositoryScope(ctx, refspec, false)
	if err != nil {
		return nil, err
	}

	for _, u := range paths {
		for _, host := range hosts {
			ctx := log.WithLogger(ctx, log.G(ctx).WithField("host", host.Host))

			req := base.request(host, http.MethodGet, u...)
			if err := req.addNamespace(base.refspec.Hostname()); err != nil {
				return nil, err
			}

			req.header["Accept"] = []string{"application/json"}

			log.G(ctx).Debug("listing")
			resp, err := req.doWithRetries(ctx, nil)
			if err != nil {
				if errors.Is(err, ErrInvalidAuthorization) {
					err = errors.Wrapf(err, "pull access denied, repository does not exist or may require authorization")
				}
				// Store the error for referencing later
				if firstErr == nil {
					firstErr = err
				}
				log.G(ctx).WithError(err).Info("trying next host")
				continue // try another host
			}

			if resp.StatusCode > 299 {
				resp.Body.Close()
				if resp.StatusCode == http.StatusNotFound {
					log.G(ctx).Info("trying next host - response was http.StatusNotFound")
					continue
				}
				if resp.StatusCode > 399 {
					// Set firstErr when encountering the first non-404 status code.
					if firstErr == nil {
						firstErr = errors.Errorf("pulling from host %s failed with status code %v: %v", host.Host, u, resp.Status)
					}
					continue // try another host
				}
				return nil, errors.Errorf("taglist from host %s failed with unexpected status code %v: %v", host.Host, u, resp.Status)
			}

			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, err
			}

			tags := &TagList{}

			err = json.Unmarshal(data, tags)
			if err != nil {
				return nil, err
			}
			return tags.Tags, nil
		}
	}

	// If above loop terminates without return, then there was an error.
	// "firstErr" contains the first non-404 error. That is, "firstErr == nil"
	// means that either no registries were given or each registry returned 404.

	if firstErr == nil {
		firstErr = errors.Wrap(errdefs.ErrNotFound, base.refspec.Locator)
	}

	return nil, firstErr
}
