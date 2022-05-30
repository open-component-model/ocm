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

package common

import (
	"fmt"
	"io"
	"strings"
)

type Printer interface {
	io.Writer
	Printf(msg string, args ...interface{}) (int, error)

	AddGap(gap string) Printer
}

type printerState struct {
	pending bool
}

type printer struct {
	writer io.Writer
	gap    string
	state  *printerState
}

func NewPrinter(writer io.Writer) Printer {
	return &printer{writer: writer, state: &printerState{true}}
}

func (p *printer) AddGap(gap string) Printer {
	return &printer{
		writer: p.writer,
		gap:    p.gap + gap,
		state:  p.state,
	}
}

func (p *printer) Write(data []byte) (int, error) {
	if p.writer == nil {
		return 0, nil
	}
	s := strings.ReplaceAll(string(data), "\n", "\n"+p.gap)
	if strings.HasSuffix(s, "\n"+p.gap) {
		p.state.pending = true
		s = s[:len(s)-len(p.gap)]
	}
	return p.writer.Write([]byte(s))
}

func (p *printer) printf(msg string, args ...interface{}) (int, error) {
	if p.writer == nil {
		return 0, nil
	}
	if p.gap == "" {
		return fmt.Fprintf(p.writer, msg, args...)
	}
	if p.state.pending {
		msg = p.gap + msg
	}
	data := fmt.Sprintf(msg, args...)
	p.state.pending = false
	return p.Write([]byte(data))
}

func (p *printer) Printf(msg string, args ...interface{}) (int, error) {
	return p.printf(msg, args...)
}
