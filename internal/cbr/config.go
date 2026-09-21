package cbr

import "time"

// TODO: Write client config
type Config struct {
	BaseUrl string

	Timeout time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		BaseUrl: "http://www.cbr.ru",

		Timeout: time.Second * 10,
	}
}
