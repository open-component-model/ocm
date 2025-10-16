package demo

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	// bind OCM configuration.
	_ "ocm.software/ocm/api/ocm/plugin/ppi/config"
)

const Name = "demo"

func New() *cobra.Command {
	cmd := &command{}
	c := &cobra.Command{
		Use:   Name + " <options>",
		Short: "a demo command",
		Long:  "a demo command in a provided command group",
		RunE:  cmd.Run,
	}

	c.Flags().StringVarP(&cmd.version, "version", "", "", "some overloaded option")
	return c
}

type command struct {
	version string
}

func (c *command) Run(cmd *cobra.Command, args []string) error {
	fmt.Printf("demo command called with arguments %v (and version option %s)\n", args, c.version)
	if strings.HasPrefix(c.version, "error") {
		msg := strings.TrimSpace(c.version[5:])
		if len(msg) != 0 {
			fmt.Fprintf(os.Stderr, "this is an error my friend\n")
		}
		return fmt.Errorf("demo error")
	}
	return nil
}
