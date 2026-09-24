package client

import (
	"fmt"
	"time"
)

type Config struct {
	BaseUrl string

	MaxRetries               int
	RetryInitialInterval     time.Duration
	RetryMaxInterval         time.Duration
	RetryRandomizationFactor float64
	RetryMultiplier          float64
}

func DefaultConfig() *Config {
	return &Config{
		BaseUrl: "http://www.cbr.ru",

		MaxRetries:               0,
		RetryInitialInterval:     500 * time.Millisecond,
		RetryMaxInterval:         60 * time.Second,
		RetryRandomizationFactor: 0.5,
		RetryMultiplier:          1.5,
	}
}

func (c *Config) Validate() error {
	if c.BaseUrl == "" {
		return fmt.Errorf("base url cannot be empty")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if c.RetryInitialInterval < 0 {
		return fmt.Errorf("retry initial interval cannot be negative")
	}

	if c.RetryMaxInterval < c.RetryInitialInterval {
		return fmt.Errorf("retry max interval should be greater than initial interval")
	}

	if c.RetryRandomizationFactor < 0 {
		return fmt.Errorf("retry randomization factor cannot be negative")
	}

	if c.RetryMultiplier < 1.0 {
		return fmt.Errorf("retry multiplier should be greater than one")
	}

	return nil
}

func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}

	if other.BaseUrl != "" {
		c.BaseUrl = other.BaseUrl
	}

	if other.MaxRetries != 0 {
		c.MaxRetries = other.MaxRetries
	}
	if other.RetryInitialInterval != 0 {
		c.RetryInitialInterval = other.RetryInitialInterval
	}
	if other.RetryMaxInterval != 0 {
		c.RetryMaxInterval = other.RetryMaxInterval
	}
	if other.RetryRandomizationFactor != 0 {
		c.RetryRandomizationFactor = other.RetryRandomizationFactor
	}
	if other.RetryMultiplier != 0 {
		c.RetryMultiplier = other.RetryMultiplier
	}
}
