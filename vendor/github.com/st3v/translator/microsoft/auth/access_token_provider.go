package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/st3v/tracerr"
)

// The AccessTokenProvider handles access tokens for Microsoft's API endpoints.
type AccessTokenProvider interface {
	RefreshToken(*accessToken) error
}

type accessTokenProvider struct {
	clientID     string
	clientSecret string
	authURL      string
}

func newAccessTokenProvider(clientID, clientSecret, authURL string) AccessTokenProvider {
	return &accessTokenProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		authURL:      authURL,
	}
}

func (p *accessTokenProvider) RefreshToken(token *accessToken) error {
	values := make(url.Values)
	values.Set("client_id", p.clientID)
	values.Set("client_secret", p.clientSecret)
	values.Set("scope", token.Scope)
	values.Set("grant_type", "client_credentials")

	response, err := http.PostForm(p.authURL, values)
	if err != nil {
		return tracerr.Wrap(err)
	}

	if response.StatusCode != http.StatusOK {
		return tracerr.Errorf("Unexpected Status Code: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return tracerr.Wrap(err)
	}

	if err := json.Unmarshal(body, token); err != nil {
		return tracerr.Wrap(err)
	}

	expiresInSeconds, err := strconv.Atoi(token.ExpiresIn)
	if err != nil {
		return tracerr.Wrap(err)
	}

	token.ExpiresAt = time.Now().Add(time.Duration(expiresInSeconds) * time.Second)

	return nil
}
