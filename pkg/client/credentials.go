package client

import (
	"net/url"
)

type BasicCredentials struct {
	Username string
	Password string
}

func (c *BasicCredentials) Basic(url *url.URL) (string, string) {
	return c.Username, c.Password
}

func (c *BasicCredentials) RefreshToken(url *url.URL, service string) string {
	return ""
}

func (c *BasicCredentials) SetRefreshToken(url *url.URL, service string, token string) {
}
