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

package printer

import (
	"context"
)

var key = "context printer"

func WithPrinter(ctx context.Context, p Printer) context.Context {
	if ctx==nil {
		ctx=context.Background()
	}
	return context.WithValue(ctx, &key, p)
}

func For(ctx context.Context) Printer {
	if ctx==nil {
		return nil
	}
	p := ctx.Value(&key)
	if p==nil {
		return nil
	}
	return p.(Printer)
}

func WithGap(ctx context.Context, gap string) context.Context {
	p := For(ctx)
	if p==nil {
		return ctx
	}
	return WithPrinter(ctx, p.AddGap(gap))
}

func Printf(ctx context.Context, msg string, args...interface{}) {
	p := For(ctx)
	if p==nil {
		return
	}
	if len(args)==0 {
		p.Printf("%s", msg)
	} else {
		p.Printf(msg, args...)
	}
}

func Warnf(ctx context.Context, msg string, args...interface{}) {
	p := For(ctx)
	if p==nil {
		return
	}
	if len(args)==0 {
		p.Warnf("%s", msg)
	} else {
		p.Warnf(msg, args...)
	}
}

func Errorf(ctx context.Context, msg string, args...interface{}) {
	p := For(ctx)
	if p==nil {
		return
	}
	if len(args)==0 {
		p.Errorf("%s", msg)
	} else {
		p.Errorf(msg, args...)
	}
}
