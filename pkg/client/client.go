package client

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
)

type Client struct {
	named      reference.Named
	insecure   bool
	transport  http.RoundTripper
	httpClient *http.Client
}

func URL(insecure bool, host string, format string, a ...interface{}) string {
	scheme := "https"
	if insecure {
		scheme = "http"
	}

	if host == "docker.io" {
		host = "index.docker.io"
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   fmt.Sprintf(format, a...),
	}
	return u.String()

}

func New(ref string, insecure bool, transport http.RoundTripper) (*Client, error) {
	named, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return nil, err
	}

	return &Client{
		named:     named,
		insecure:  insecure,
		transport: transport,
		httpClient: &http.Client{
			Transport: transport,
		},
	}, nil
}

func (c *Client) Named() reference.Named {
	return c.named
}

func (c *Client) Scope() string {
	return reference.Path(c.named)
}

func (c *Client) URL(format string, a ...interface{}) string {
	return URL(c.insecure, reference.Domain(c.named), format, a...)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) Auth(creds auth.CredentialStore, scope string, actions ...string) error {
	resp, err := c.httpClient.Get(c.URL("/v2/"))
	if err != nil {
		return fmt.Errorf("get challenges from /v2/: %s", err)
	}
	defer resp.Body.Close()

	manager := challenge.NewSimpleManager()
	if err := manager.AddResponse(resp); err != nil {
		return fmt.Errorf("add response to challenge manager: %s", err)
	}

	authorizer := auth.NewAuthorizer(
		manager,
		auth.NewTokenHandler(c.transport, creds, scope, actions...),
	)

	rt := transport.NewTransport(c.transport, authorizer)
	c.httpClient.Transport = rt
	return nil
}
