package auth

import "time"

type accessToken struct {
	Token     string `json:"access_token"`
	Type      string `json:"token_type"`
	Scope     string `json:"scope"`
	ExpiresIn string `json:"expires_in"`
	ExpiresAt time.Time
}

func newAccessToken(scope string) *accessToken {
	return &accessToken{
		Scope: scope,
	}
}

func (t *accessToken) expired() bool {
	// be conservative and expire 10 seconds early
	return t.ExpiresAt.Before(time.Now().Add(time.Second * 10))
}
