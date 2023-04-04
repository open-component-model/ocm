// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"

	"github.com/open-component-model/ocm/pkg/common"
	globalconfig "github.com/open-component-model/ocm/pkg/contexts/config/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/directcreds"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory"
	memorycfg "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory/config"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/toi"
	"github.com/open-component-model/ocm/pkg/utils"
)

type (
	Credentials            = toi.Credentials
	CredentialSpec         = toi.CredentialSpec
	CredentialsRequest     = toi.CredentialsRequest
	CredentialsRequestSpec = toi.CredentialsRequestSpec
)

func ParseCredentialSpecification(data []byte, desc string) (*Credentials, error) {
	spiff := spiffing.New().WithFeatures(features.CONTROL, features.INTERPOLATION)

	templ, err := spiff.Unmarshal(desc, data)
	if err != nil {
		return nil, errors.Newf("invalid credential settings: %s", err)
	}

	cfg, err := spiff.Cascade(templ, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error processing credential settings")
	}
	final, err := spiff.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "credential marshalling")
	}
	var spec Credentials

	err = runtime.DefaultYAMLEncoding.Unmarshal(final, &spec)
	if err != nil {
		return nil, errors.Wrapf(err, "credentials settings")
	}
	return &spec, nil
}

func ParseCredentialRequest(data []byte) (*CredentialsRequest, error) {
	var req CredentialsRequest

	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &req)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse credential request")
	}
	return &req, err
}

func GetCredentials(ctx credentials.Context, spec *Credentials, req map[string]CredentialsRequestSpec, mapping map[string]string) (*globalconfig.Config, error) {
	cfg := config.New()
	mem := memorycfg.New("default")
	memrepo := memory.NewRepositorySpec("default")
	list := errors.ErrListf("providing requested credentials")
	var sub *errors.ErrorList
	for _, n := range utils.StringMapKeys(req) {
		r := req[n]
		list.Add(sub.Result())
		sub = errors.ErrListf("credential request %q", n)
		found, ok := spec.Credentials[n]
		if !ok {
			if !r.Optional {
				sub.Add(errors.ErrNotFound("credential", n))
			}
			continue
		}
		creds, consumer, err := evaluate(ctx, &found)
		if err != nil {
			sub.Add(errors.Wrapf(err, "failed to evaluate"))
			continue
		}
		mapped := n
		if mapping != nil {
			mapped = mapping[n]
		}
		if mapped == "" {
			return nil, errors.Newf("mapping missing credential %q", n)
		}
		err = mem.AddCredentials(mapped, creds)
		if err != nil {
			sub.Add(errors.Wrapf(err, "failed to add credentials"))
			continue
		}
		if len(consumer) != 0 {
			err = cfg.AddConsumer(consumer, credentials.NewCredentialsSpec(mapped, memrepo))
			if err != nil {
				sub.Add(errors.Newf("failed to add consumer %s from config", consumer))
				continue
			}
		}
		if len(r.ConsumerId) != 0 {
			err = cfg.AddConsumer(r.ConsumerId, credentials.NewCredentialsSpec(mapped, memrepo))
			if err != nil {
				sub.Add(errors.Newf("failed to add consumer %s from request", consumer))
				continue
			}
		}
	}
	for _, r := range spec.Forwarded {
		if len(r.ConsumerId) == 0 {
			return nil, errors.ErrInvalid("consumer", r.ConsumerId.String())
		}
		match, _ := ctx.ConsumerIdentityMatchers().Get(r.ConsumerType)
		if match == nil {
			match = credentials.PartialMatch
		}
		src, err := ctx.GetCredentialsForConsumer(r.ConsumerId, match)
		if err != nil || src == nil {
			return nil, errors.ErrNotFoundWrap(err, "consumer", r.ConsumerId.String())
		}
		if src == nil {
			return nil, errors.ErrNotFoundWrap(err, "consumer", r.ConsumerId.String())
		}
		creds, err := src.Credentials(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get credentials for %s", r.ConsumerId.String())
		}
		props := creds.Properties()
		cfg.AddConsumer(r.ConsumerId, directcreds.NewCredentials(props))
	}

	list.Add(sub.Result())
	main := globalconfig.New()
	main.AddConfig(mem)
	main.AddConfig(cfg)
	return main, list.Result()
}

func evaluate(ctx credentials.Context, spec *CredentialSpec) (common.Properties, credentials.ConsumerIdentity, error) {
	var err error
	var props common.Properties
	var src credentials.CredentialsSource
	cnt := 0
	if len(spec.Credentials) > 0 {
		cnt++
		props = spec.Credentials
	}
	if spec.Reference != nil {
		cnt++
		src, err = spec.Reference.Credentials(ctx)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot evaluate credential reference")
		}
	}
	if spec.ConsumerId != nil {
		cnt++
		match, _ := ctx.ConsumerIdentityMatchers().Get(spec.ConsumerType)
		if match == nil {
			match = credentials.PartialMatch
		}
		src, err = ctx.GetCredentialsForConsumer(spec.ConsumerId, match)
		if err != nil {
			return nil, nil, errors.ErrNotFoundWrap(err, "consumer", spec.ConsumerId.String())
		}
	}
	if cnt > 1 {
		return nil, nil, errors.Newf("only one of consumer id or reference or credentials possible")
	}
	if src != nil {
		creds, err := src.Credentials(ctx)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot get credentials for %s", spec.ConsumerId.String())
		}
		props = creds.Properties()
	}

	return props, spec.TargetConsumerId, nil
}
