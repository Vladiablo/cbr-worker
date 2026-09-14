package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type CollectorArgs struct {
	databaseUrl string

	fromDate time.Time
	toDate   time.Time

	timeout time.Duration

	concurrency int
}

type DateFlag struct {
	time.Time
}

var _ flag.Value = (*DateFlag)(nil)

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

func parseEnvDate(name string) (time.Time, error) {
	if strDate := os.Getenv(name); strDate != "" {
		date, err := time.Parse(time.DateOnly, strDate)
		if err != nil {
			return time.Time{}, fmt.Errorf("%v has invalid format: %w", name, err)
		}

		return date, nil
	}

	return time.Time{}, nil
}

func (a *CollectorArgs) parseEnv() error {
	databaseUrl := os.Getenv("DATABASE_URL")

	fromDate, err := parseEnvDate("COLLECTOR_FROM_DATE")
	if err != nil {
		return err
	}

	toDate, err := parseEnvDate("COLLECTOR_TO_DATE")
	if err != nil {
		return err
	}

	var timeout time.Duration
	strTimeout := os.Getenv("COLLECTOR_TIMEOUT")
	if strTimeout != "" {
		timeout, err = time.ParseDuration(strTimeout)
		if err != nil {
			return fmt.Errorf("COLLECTOR_TIMEOUT has invalid format: %w", err)
		}
	}

	var concurrency int64
	strConcurrency := os.Getenv("COLLECTOR_CONCURRENCY")
	if strConcurrency != "" {
		concurrency, err = strconv.ParseInt(strConcurrency, 10, 0)
		if err != nil {
			return fmt.Errorf("COLLECTOR_CONCURRENCY has invalid format: %w", err)
		}
	}

	a.applyArgs(&CollectorArgs{
		databaseUrl: databaseUrl,
		fromDate:    fromDate,
		toDate:      toDate,
		timeout:     timeout,
		concurrency: int(concurrency),
	})

	return nil
}

func (a *CollectorArgs) parseCmdLineArgs() error {
	fs := flag.NewFlagSet("collector", flag.ContinueOnError)

	var databaseUrl string
	fs.StringVar(&databaseUrl, "database-url", "", "")

	var fromDate DateFlag
	fs.Var(&fromDate, "from-date", "")

	var toDate DateFlag
	fs.Var(&toDate, "to-date", "")

	var timeout time.Duration
	fs.DurationVar(&timeout, "timeout", 0, "")

	var concurrency int
	fs.IntVar(&concurrency, "concurrency", 0, "")

	err := fs.Parse(os.Args[1:])
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		return fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	a.applyArgs(&CollectorArgs{
		databaseUrl: databaseUrl,
		fromDate:    fromDate.Time,
		toDate:      toDate.Time,
		timeout:     timeout,
		concurrency: concurrency,
	})

	return nil
}

func (a *CollectorArgs) applyArgs(other *CollectorArgs) {
	if a.databaseUrl == "" {
		a.databaseUrl = other.databaseUrl
	}

	if a.fromDate.IsZero() {
		a.fromDate = other.fromDate
	}

	if a.toDate.IsZero() {
		a.toDate = other.toDate
	}

	if a.timeout == 0 {
		a.timeout = other.timeout
	}

	if a.concurrency == 0 {
		a.concurrency = other.concurrency
	}
}

func parseArgs() (*CollectorArgs, error) {
	args := &CollectorArgs{}

	if err := args.parseCmdLineArgs(); err != nil {
		return nil, err
	}

	if err := args.parseEnv(); err != nil {
		return nil, err
	}

	return args, nil
}
