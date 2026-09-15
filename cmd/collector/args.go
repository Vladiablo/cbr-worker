package main

import (
	"encoding"
	"fmt"
	"time"

	"github.com/alexflint/go-arg"
)

type CollectorArgs struct {
	DatabaseUrl string `arg:"--database,env:DATABASE_URL"`

	FromDate DateFlag `arg:"--from-date,env:FROM_DATE"`
	ToDate   DateFlag `arg:"--to-date,env:TO_DATE"`

	Timeout time.Duration `arg:"--timeout,env:TIMEOUT"`

	Concurrency int `arg:"--concurrency,env:CONCURRENCY"`
}
type DateFlag struct {
	time.Time
}

var _ encoding.TextUnmarshaler = (*DateFlag)(nil)

func (d *DateFlag) UnmarshalText(text []byte) error {
	dt, err := time.Parse(time.DateOnly, string(text))
	if err != nil {
		return fmt.Errorf("invalid date format %q: %w", text, err)
	}

	*d = DateFlag{Time: dt}

	return nil
}

func (d DateFlag) MarshalText() ([]byte, error) {
	res := d.Format(time.DateOnly)

	return []byte(res), nil
}

func ParseArgs() (*CollectorArgs, error) {
	var args CollectorArgs

	if err := arg.Parse(&args); err != nil {
		return nil, fmt.Errorf("parse args: %w", err)
	}

	return &args, nil
}
