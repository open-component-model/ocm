// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
)

const (
	KIND_PLUGIN       = internal.KIND_PLUGIN
	KIND_UPLOADER     = internal.KIND_UPLOADER
	KIND_ACCESSMETHOD = internal.KIND_ACCESSMETHOD
)

var TAG = internal.TAG

type (
	Descriptor           = internal.Descriptor
	AccessSpecInfo       = internal.AccessSpecInfo
	UploadTargetSpecInfo = internal.UploadTargetSpecInfo
)
