package http

import "net/http"

// NewAuthenticatedClient returns an HTTP client with a mocked-out authenticator.
func NewAuthenticatedClient() Client {
	authenticator := newMockAuthenticator(func(request *http.Request) error {
		request.Header.Set("Authorization", "fake-authorization")
		return nil
	})

	return NewClient(authenticator)
}

func newMockAuthenticator(authenticate func(request *http.Request) error) *mockAuthenticator {
	return &mockAuthenticator{
		authenticate: authenticate,
	}
}

type mockAuthenticator struct {
	authenticate func(request *http.Request) error
}

func (a *mockAuthenticator) Authenticate(request *http.Request) error {
	if a.authenticate != nil {
		return a.authenticate(request)
	}
	return nil
}
