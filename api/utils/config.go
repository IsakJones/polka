package utils

import "fmt"

type Config struct {
	Host string
	Port int
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) GetListenPort() string {
	return fmt.Sprintf(":%d", c.Port)
}
