// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package spi

////////////////////////////////////////////////////////////////////////////////
// Access Type Options

type EntryOptionTarget interface {
	SetFormat(string)
	SetDescription(string)
}

type EntryTypeOption interface {
	ApplyToEntryOptionTarget(EntryOptionTarget)
}

////////////////////////////////////////////////////////////////////////////////

type formatOption struct {
	value string
}

func WithFormatSpec(value string) EntryTypeOption {
	return formatOption{value}
}

func (o formatOption) ApplyToEntryOptionTarget(t EntryOptionTarget) {
	t.SetFormat(o.value)
}

////////////////////////////////////////////////////////////////////////////////

type descriptionOption struct {
	value string
}

func WithDescription(value string) EntryTypeOption {
	return descriptionOption{value}
}

func (o descriptionOption) ApplyToEntryOptionTarget(t EntryOptionTarget) {
	t.SetDescription(o.value)
}
