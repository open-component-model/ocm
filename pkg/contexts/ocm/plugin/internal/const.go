// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/errors"
)

const (
	KIND_PLUGIN       = "plugin"
	KIND_DOWNLOADER   = "downloader"
	KIND_UPLOADER     = "uploader"
	KIND_ACCESSMETHOD = errors.KIND_ACCESSMETHOD
	KIND_ACTION       = action.KIND_ACTION
)

var TAG = logging.NewTag("plugins")
