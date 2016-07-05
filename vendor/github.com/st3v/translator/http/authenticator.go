package http

import "net/http"

// Authenticator is used to authenticate HTTP requests to API endpoints.
type Authenticator interface {
	// Authenticate a given HTTP request.
	Authenticate(request *http.Request) error
}
