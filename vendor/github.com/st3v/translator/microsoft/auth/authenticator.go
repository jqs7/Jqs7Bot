package auth

import (
	"net/http"

	"github.com/st3v/tracerr"
	_http "github.com/st3v/translator/http"
)

type authenticator struct {
	accessTokenProvider AccessTokenProvider
	accessTokenChan     chan *accessToken
}

// NewAuthenticator returns an autheticator for Microsoft API endpoints.
func NewAuthenticator(clientID, clientSecret, scope, authURL string) _http.Authenticator {
	// make buffered accessToken channel and pre-fill it with an expired token
	tokenChan := make(chan *accessToken, 1)
	tokenChan <- newAccessToken(scope)

	// return new authenticator that uses the above accessToken channel
	return &authenticator{
		accessTokenProvider: newAccessTokenProvider(clientID, clientSecret, authURL),
		accessTokenChan:     tokenChan,
	}
}

func (a *authenticator) Authenticate(request *http.Request) error {
	authToken, err := a.authToken()
	if err != nil {
		return tracerr.Wrap(err)
	}

	request.Header.Add("Authorization", authToken)
	return nil
}

func (a *authenticator) authToken() (string, error) {
	// grab the token
	accessToken := <-a.accessTokenChan

	// make sure it's valid, otherwise request a new one
	if accessToken == nil || accessToken.expired() {
		err := a.accessTokenProvider.RefreshToken(accessToken)
		if err != nil || accessToken == nil {
			a.accessTokenChan <- nil
			return "", tracerr.Wrap(err)
		}
	}

	// put the token back on the channel
	a.accessTokenChan <- accessToken

	// return authToken
	return "Bearer " + accessToken.Token, nil
}
