package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
)

type aggregatedError []error

func (e aggregatedError) Error() string {
	s := make([]string, len(e))
	for i, err := range e {
		s[i] = err.Error()
	}
	return strings.Join(s, "; ")
}

type connectionType int

const (
	httpsConnection connectionType = iota
	httpConnection
)

func (c connectionType) Scheme() string {
	if c == httpConnection {
		return "http"
	}
	return "https"
}

type Client struct {
	named      reference.Named
	insecure   bool
	connection connectionType
	transport  http.RoundTripper
	httpClient *http.Client
}

func URL(scheme string, host string, format string, a ...interface{}) string {
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
		named:      named,
		insecure:   insecure,
		connection: httpsConnection,
		transport:  transport,
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
	return URL(c.connection.Scheme(), reference.Domain(c.named), format, a...)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) auth(creds auth.CredentialStore, scope string, actions ...string) error {
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

func (c *Client) Auth(creds auth.CredentialStore, scope string, actions ...string) error {
	connectionTypes := []connectionType{httpsConnection}
	if c.insecure {
		connectionTypes = append(connectionTypes, httpConnection)
	}

	var errs []error
	for _, connection := range connectionTypes {
		c.connection = connection
		err := c.auth(creds, scope, actions...)
		if err == nil {
			return nil
		}
		errs = append(errs, err)
	}
	return aggregatedError(errs)
}
