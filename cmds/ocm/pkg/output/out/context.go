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

package out

import (
	"fmt"
	"io"
	"os"
)

type Context interface {
	StdOut() io.Writer
	StdErr() io.Writer
	StdIn() io.Reader
}

func Outf(ctx Context, msg string, args ...interface{}) (int, error) {
	if len(args) == 0 {
		return fmt.Fprint(ctx.StdOut(), msg)
	}
	return fmt.Fprintf(ctx.StdOut(), msg, args...)
}

func Out(ctx Context, args ...interface{}) (int, error) {
	return fmt.Fprint(ctx.StdOut(), args...)
}

func Outln(ctx Context, args ...interface{}) (int, error) {
	return fmt.Fprintln(ctx.StdOut(), args...)
}

func Errf(ctx Context, msg string, args ...interface{}) (int, error) {
	if len(args) == 0 {
		return fmt.Fprint(ctx.StdErr(), msg)
	}
	return fmt.Fprintf(ctx.StdErr(), msg, args...)
}

func Err(ctx Context, args ...interface{}) (int, error) {
	return fmt.Fprint(ctx.StdOut(), args...)
}

func Error(ctx Context, msg string, args ...interface{}) (int, error) {
	return Errf(ctx, "Error: "+msg+"\n", args...)
}

func Warning(ctx Context, msg string, args ...interface{}) (int, error) {
	return Errf(ctx, "Warning: "+msg+"\n", args...)
}

////////////////////////////////////////////////////////////////////////////////

type outputContext struct {
	parent Context
	in     io.Reader
	out    io.Writer
	err    io.Writer
}

var DefaultContext = New()

func New() Context {
	return &outputContext{
		in:  os.Stdin,
		out: os.Stdout,
		err: os.Stderr,
	}
}

func NewFor(ctx Context) Context {
	if ctx == nil {
		return DefaultContext
	}
	return &outputContext{
		in:  ctx.StdIn(),
		out: ctx.StdOut(),
		err: ctx.StdErr(),
	}
}

func (o *outputContext) StdOut() io.Writer {
	if o == nil {
		return os.Stdout
	}
	if o.out != nil {
		return o.out
	}
	return o.parent.StdOut()
}

func (o *outputContext) StdErr() io.Writer {
	if o == nil {
		return os.Stderr
	}
	if o.err != nil {
		return o.err
	}
	return o.parent.StdErr()
}

func (o *outputContext) StdIn() io.Reader {
	if o == nil {
		return os.Stdin
	}
	if o.in != nil {
		return o.in
	}
	return o.parent.StdIn()
}

func WithInput(ctx Context, in io.Reader) Context {
	if ctx == nil {
		ctx = DefaultContext
	}
	if in == nil {
		return ctx
	}
	return &outputContext{parent: ctx, in: in}
}

func WithOutput(ctx Context, out io.Writer) Context {
	if ctx == nil {
		ctx = DefaultContext
	}
	if out == nil {
		return ctx
	}
	return &outputContext{parent: ctx, out: out}
}

func WithErrorOutput(ctx Context, err io.Writer) Context {
	if ctx == nil {
		ctx = DefaultContext
	}
	if err == nil {
		return ctx
	}
	return &outputContext{parent: ctx, err: err}
}

func WithStdIO(ctx Context, r io.Reader, o io.Writer, e io.Writer) Context {
	return &outputContext{parent: ctx, in: r, out: o, err: e}
}
