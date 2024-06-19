package demo

import (
	"fmt"

	// bind OCM configuration.
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/config"

	"github.com/spf13/cobra"
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
	return c
}

type command struct {
}

func (c *command) Run(cmd *cobra.Command, args []string) error {
	fmt.Printf("demo command called with arguments %v\n", args)
	return nil
}
