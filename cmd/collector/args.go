package main

import (
	"flag"
	"fmt"
	"time"
)

type CollectorArgs struct {
	from DateFlag
	to   DateFlag
}

type DateFlag struct {
	time.Time
}

func (d *DateFlag) String() string {
	return d.Format(time.DateOnly)
}

func (d *DateFlag) Set(s string) error {
	date, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	d.Time = date

	return nil
}

func (c *CollectorArgs) Validate() error {
	if !c.to.IsZero() && c.from.IsZero() {
		return fmt.Errorf("-to should be used only with -from")
	}

	if c.to.Before(c.from.Time) {
		return fmt.Errorf("-to should be greater than -from")
	}

	return nil
}

func parseArgs() *CollectorArgs {
	args := &CollectorArgs{}

	flag.Var(&args.from, "from", "")
	flag.Var(&args.to, "to", "")

	flag.Parse()

	return args
}
