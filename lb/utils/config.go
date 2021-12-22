package utils

import "net/url"

// Configuration is an interface for the Config struct in main.go.
type Config interface {
	GetAddress() *url.URL
	GetListenPort() string
}
