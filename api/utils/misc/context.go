package misc

import (
	"context"
	"reflect"
)

var printerKey = reflect.TypeOf(printer{})

func WithPrinter(ctx context.Context, printer Printer) context.Context {
	return context.WithValue(ctx, printerKey, AssurePrinter(printer))
}

func GetPrinter(ctx context.Context) Printer {
	p := ctx.Value(printerKey)
	if p == nil {
		return NonePrinter
	}
	return p.(Printer)
}

func AddPrinterGap(ctx context.Context, gap string) context.Context {
	return WithPrinter(ctx, GetPrinter(ctx).AddGap(gap))
}

func IsContextCanceled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
