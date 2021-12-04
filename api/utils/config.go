package utils

// Configuration is an interface for the Config struct in main.go.
type Config interface {
	GetHost() string
	GetPort() int
	GetAddress() string
	GetListenPort() string
}
