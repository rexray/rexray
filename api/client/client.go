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

	// ServerName returns the name of the server to which the client is
	// connected. This is not the same as the host name, rather it's the
	// randomly generated name the server creates for unique identification
	// when the server starts for the first time. This value is updated
	// by every request to the server that returns the server name header
	// as part of its response.
	ServerName string
}
