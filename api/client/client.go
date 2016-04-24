package client

import (
	"net/http"
)

// Client is the libStorage API client.
type Client struct {

	// HTTP is the underlying HTTP client.
	Client *http.Client

	// Host is the host[:port] of the remote libStorage API.
	Host string

	// LogRequests is a flag indicating whether or not to log HTTP requests.
	LogRequests bool

	// LogResponses is a flag indicating whether or not to log HTTP responses.
	LogResponses bool

	// Headers are headers to send with each HTTP request.
	Headers http.Header
}
