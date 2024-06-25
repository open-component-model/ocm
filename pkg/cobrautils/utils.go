package cobrautils

import (
	"github.com/spf13/cobra"
)

func Find(cmd *cobra.Command, name string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
