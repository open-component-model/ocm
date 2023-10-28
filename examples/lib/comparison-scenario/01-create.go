// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements"
	exthelm "github.com/open-component-model/ocm/pkg/contexts/ocm/elements/externalartifacts/helm"
	extociartifact "github.com/open-component-model/ocm/pkg/contexts/ocm/elements/externalartifacts/ociartifact"
	localfile "github.com/open-component-model/ocm/pkg/contexts/ocm/elements/localartifacts/file"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
)

const (
	PODINFO_IMAGE  = "ghcr.io/stefanprodan/podinfo:6.5.2"
	HELMCHART_REPO = "oci://ghcr.io/stefanprodan/charts"
	HELMCHART_NAME = "podinfo:6.5.2"
)

const (
	RSC_IMAGE     = "podinfo-image"
	RSC_HELMCHART = "helmchart"
	RSC_DEPLOY    = "deplyscript"
)
const DEPLOY_SCRIPT_TYPE = "helmDeployScript"

func Create(cfg *helper.Config) error {
	ctx := ocm.DefaultContext()

	cv, err := CreateComponentVersion(ctx)
	if err != nil {
		return err
	}
	defer cv.Close()
	return nil
}

// CreateComponentVersion creates the scenario component version with
// three resources: the podinfo image, the helm chart and a locally
// found deploy script.
func CreateComponentVersion(ctx ocm.Context) (ocm.ComponentVersionAccess, error) {
	fmt.Printf("*** composing component version %s:%s\n", COMPONENT_NAME, COMPONENT_VERSION)

	cv := composition.NewComponentVersion(ctx, COMPONENT_NAME, COMPONENT_VERSION)

	cv.SetProvider(&metav1.Provider{
		Name: "acme.org",
	})

	// podinfo image as external resource reference
	image_meta, err := elements.ResourceMeta(RSC_IMAGE, resourcetypes.OCI_IMAGE)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create resource meta for podinfo-image")
	}
	image_res := extociartifact.ResourceAccess(ctx, image_meta, PODINFO_IMAGE)
	err = cv.SetResourceAccess(image_res)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot add resource podinfo-image")
	}

	// helm chart as external resource reference
	helm_meta, err := elements.ResourceMeta(RSC_HELMCHART, resourcetypes.HELM_CHART)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create resource meta for helmchart")
	}
	helm_res := exthelm.ResourceAccess(ctx, helm_meta, HELMCHART_NAME, HELMCHART_REPO)
	err = cv.SetResourceAccess(helm_res)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot add resource helmchart")
	}

	// deploy script found in filesystem
	script_meta, err := elements.ResourceMeta(RSC_DEPLOY, DEPLOY_SCRIPT_TYPE)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create resource meta for podinfo-image")
	}
	script_res := localfile.ResourceAccess(ctx, mime.MIME_YAML, script_meta, "resources/deployscript")

	err = cv.SetResourceAccess(script_res)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot add resource helmchart")
	}

	return cv, nil
}
