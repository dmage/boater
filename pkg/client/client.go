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
	flag "github.com/spf13/pflag"
)

type GetManifestOptions struct {
	AcceptKnown        bool
	AcceptSchema1      bool
	AcceptSchema2      bool
	AcceptManifestList bool
	AcceptOCISchema    bool
	AcceptOCIIndex     bool
	MediaTypes         []string
}

func (o *GetManifestOptions) AddToFlagSet(fs *flag.FlagSet) {
	fs.BoolVarP(&o.AcceptKnown, "accept-known", "a", o.AcceptKnown, "accept all known manifest types (as new types may be added in the future, this option does not guarantee backward compatibility)")
	fs.BoolVar(&o.AcceptSchema1, "accept-schema1", o.AcceptSchema1, "accept Schema 1 manifests (application/vnd.docker.distribution.manifest.v1+json)")
	fs.BoolVar(&o.AcceptSchema2, "accept-schema2", o.AcceptSchema2, "accept Schema 2 manifests (application/vnd.docker.distribution.manifest.v2+json)")
	fs.BoolVar(&o.AcceptManifestList, "accept-manifest-list", o.AcceptManifestList, "accept manifest lists (application/vnd.docker.distribution.manifest.list.v2+json)")
	fs.BoolVar(&o.AcceptOCISchema, "accept-ocischema", o.AcceptOCISchema, "accept OCI image manifests (application/vnd.oci.image.manifest.v1+json)")
	fs.BoolVar(&o.AcceptOCIIndex, "accept-oci-index", o.AcceptOCIIndex, "accept OCI image index (application/vnd.oci.image.index.v1+json)")
	fs.StringArrayVarP(&o.MediaTypes, "accept", "t", o.MediaTypes, "accept manifests with a custom media type")
}

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

func (c *Client) GetManifest(name string, opts GetManifestOptions) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.URL("/v2/%s/manifests/%s", c.Scope(), name), nil)
	if err != nil {
		return nil, err
	}
	if opts.AcceptKnown || opts.AcceptSchema1 {
		req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v1+json")
	}
	if opts.AcceptKnown || opts.AcceptSchema2 {
		req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	}
	if opts.AcceptKnown || opts.AcceptManifestList {
		req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	}
	if opts.AcceptKnown || opts.AcceptOCISchema {
		req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")
	}
	if opts.AcceptKnown || opts.AcceptOCIIndex {
		req.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")
	}
	for _, mediatype := range opts.MediaTypes {
		req.Header.Add("Accept", mediatype)
	}
	return c.Do(req)
}
