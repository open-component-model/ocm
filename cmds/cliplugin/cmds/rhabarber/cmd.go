package rhabarber

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mandelsoft/goutils/errors"
	"github.com/open-component-model/ocm/pkg/logging"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &command{}
	c := &cobra.Command{
		Use:   "rhabarber <options>",
		Short: "determine whether we are in rhubarb season",
		Long:  "The rhubarb season is between march and april.",
		RunE:  cmd.Run,
	}

	c.Flags().StringVarP(&cmd.date, "date", "d", "", "the date to ask for (MM/DD)")
	return c
}

type command struct {
	date string
}

var months = map[string]int{
	"jan": 1,
	"feb": 2,
	"mär": 3, "mar": 3,
	"apr": 4,
	"mai": 5, "may": 5,
	"jun": 6,
	"jul": 7,
	"aug": 8,
	"sep": 9,
	"okt": 10, "oct": 10,
	"nov": 11,
	"dez": 12, "dec": 12,
}

func (c *command) Run(cmd *cobra.Command, args []string) error {
	d := time.Now()
	if c.date != "" {
		parts := strings.Split(c.date, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid date, expected MM/DD")
		}
		month, err := strconv.Atoi(parts[0])
		if err != nil {
			month = months[strings.ToLower(parts[0])]
			if month == 0 {
				return errors.Wrapf(err, "invalid month")
			}
		}
		day, err := strconv.Atoi(parts[1])
		if err != nil {
			return errors.Wrapf(err, "invalid day")
		}
		logging.Context().Logger().Debug("testing rhabarb season", "date", d.String())
		d = time.Date(d.Year(), time.Month(month), day, 0, 0, 0, 0, time.Local)
	}

	if d.Month() >= time.March && d.Month() <= time.April {
		fmt.Printf("Yeah, it's rhabarb season - happy rhabarbing!")
	} else {
		fmt.Printf("Sorry, but you have to stay hungry.")
	}
	return nil
}
