package common

import (
	"fmt"

	"ocm.software/ocm/api/utils/out"
)

func Outf(ctx out.Context, dryRun bool, msg string, args ...any) (int, error) {
	if dryRun {
		return -1, nil
	}

	if len(args) == 0 {
		return fmt.Fprint(ctx.StdOut(), msg)
	}

	return fmt.Fprintf(ctx.StdOut(), msg, args...)
}
