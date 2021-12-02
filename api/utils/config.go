package utils

import "fmt"

// // Configuration is an interface for the Config struct in main.go.
// type Config interface {
// 	GetHost()
// 	GetPort()
// 	GetAddress()
// 	GetListenPort()
// }
type Config struct {
	Host string
	Port int
}

func (c *Config) GetHost() string {
	return c.Host
}

func (c *Config) GetPort() int {
	return c.Port
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) GetListenPort() string {
	return fmt.Sprintf(":%d", c.Port)
}
