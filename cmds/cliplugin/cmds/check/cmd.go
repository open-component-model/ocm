package check

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	"github.com/spf13/cobra"
	"ocm.software/ocm/api/ocm"
	// bind OCM configuration.
	_ "ocm.software/ocm/api/ocm/plugin/ppi/config"
)

const Name = "check"

var log = logging.DynamicLogger(logging.DefaultContext(), logging.NewRealm("cliplugin/rhabarber"))

func New() *cobra.Command {
	cmd := &command{}
	c := &cobra.Command{
		Use:   Name + " <options>",
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
	season := Season{
		Start: "mar/1",
		End:   "apr/30",
	}

	ctx := ocm.FromContext(cmd.Context())
	ctx.ConfigContext().ApplyTo(0, &season)

	start, err := ParseDate(season.Start)
	if err != nil {
		return errors.Wrapf(err, "invalid season start")
	}

	end, err := ParseDate(season.End)
	if err != nil {
		return errors.Wrapf(err, "invalid season end")
	}
	end = end.Add(time.Hour * 24)

	d := time.Now()
	if c.date != "" {
		d, err = ParseDate(c.date)
		if err != nil {
			return err
		}
	}

	log.Debug("testing rhabarb season", "date", d.String())
	if d.After(start) && d.Before(end) {
		fmt.Printf("Yeah, it's rhabarb season - happy rhabarbing!\n")
	} else {
		fmt.Printf("Sorry, but you have to stay hungry.\n")
	}
	return nil
}

func ParseDate(s string) (time.Time, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid date, expected MM/DD")
	}
	month, err := strconv.Atoi(parts[0])
	if err != nil {
		month = months[strings.ToLower(parts[0])]
		if month == 0 {
			return time.Time{}, errors.Wrapf(err, "invalid month")
		}
	}
	day, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "invalid day")
	}

	return time.Date(time.Now().Year(), time.Month(month), day, 0, 0, 0, 0, time.Local), nil //nolint:gosmopolitan // yes
}
