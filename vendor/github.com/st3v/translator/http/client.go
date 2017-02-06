package http

import (
	"io"
	"net/http"

	"github.com/st3v/tracerr"
)

// Client sends authenticated HTTP requests to API endpoints
type Client interface {
	SendRequest(method, uri string, body io.Reader, contentType string) (*http.Response, error)
}

type client struct {
	client        *http.Client
	authenticator Authenticator
}

// NewClient instantiates a Client and initializes it with the passed Authenticator.
func NewClient(authenticator Authenticator) Client {
	return &client{
		client:        &http.Client{},
		authenticator: authenticator,
	}
}

func (h *client) SendRequest(method, uri string, body io.Reader, contentType string) (*http.Response, error) {
	request, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	request.Header.Add("Content-Type", contentType)

	err = h.authenticator.Authenticate(request)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	response, err := h.client.Do(request)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return response, nil
}
