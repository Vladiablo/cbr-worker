package collector

import (
	"fmt"
	"time"
)

type Config struct {
	FromDate    time.Time
	ToDate      time.Time
	Timeout     time.Duration
	Concurrency int
}

func (c *Config) Validate() error {
	if !c.ToDate.IsZero() && c.FromDate.IsZero() {
		return fmt.Errorf("toDate can only be used with fromDate")
	}

	if c.ToDate.Before(c.FromDate) {
		return fmt.Errorf("toDate cannot be before fromDate")
	}

	if c.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if c.Concurrency < 0 {
		return fmt.Errorf("concurrency cannot be negative")
	}

	return nil
}
