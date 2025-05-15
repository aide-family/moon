package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Client interface {
	Clone() Client
	WithHeader(header http.Header) Client
	WithHeaderKV(key, value string) Client
	WithBasicAuth(username, password string) Client
	WithContext(ctx context.Context) Client
	WithTLSClientConfig(cert, key string) Client
	WithRootCA(caPath string) Client
	WithServerName(serverName string) Client
	Do(req *http.Request) (*http.Response, error)
	Get(api string, params ...url.Values) (*http.Response, error)
	Post(api string, body []byte) (*http.Response, error)
	PostForm(api string, data url.Values) (*http.Response, error)
	PostJson(api string, body []byte) (*http.Response, error)
}

var (
	httpClient     = http.DefaultClient
	httpClientOnce sync.Once
)

func SetHttpClient(cli *http.Client) {
	httpClientOnce.Do(func() {
		httpClient = cli
	})
}

func GetHttpClient() *http.Client {
	return httpClient
}

func NewClient(opts ...Option) Client {
	cli := &client{
		cli:    httpClient,
		header: make(http.Header),
		ctx:    context.Background(),
	}
	for _, opt := range opts {
		opt(cli)
	}
	return cli
}

type client struct {
	cli                *http.Client
	header             http.Header
	username, password string
	ctx                context.Context
	tlsConfig          *tls.Config
}

type Option func(*client)

func WithContext(ctx context.Context) Option {
	return func(c *client) {
		c.ctx = context.WithoutCancel(ctx)
	}
}

func WithHeader(header http.Header) Option {
	return func(c *client) {
		for k, v := range header {
			c.header.Set(k, v[0])
		}
	}
}

func WithBasicAuth(username, password string) Option {
	return func(c *client) {
		c.username = username
		c.password = password
	}
}

func WithClient(cli *http.Client) Option {
	return func(c *client) {
		c.cli = cli
	}
}

func (c *client) Clone() Client {
	return &client{
		cli:       c.cli,
		header:    c.header.Clone(),
		username:  c.username,
		password:  c.password,
		ctx:       c.ctx,
		tlsConfig: c.tlsConfig,
	}
}

func (c *client) WithContext(ctx context.Context) Client {
	clone := c.Clone().(*client)
	WithContext(ctx)(clone)
	return clone
}

func (c *client) WithHeader(header http.Header) Client {
	clone := c.Clone().(*client)
	WithHeader(header)(clone)
	return clone
}

func (c *client) WithHeaderKV(key, value string) Client {
	clone := c.Clone().(*client)
	c.header.Set(key, value)
	return clone
}

func (c *client) WithBasicAuth(username, password string) Client {
	clone := c.Clone().(*client)
	WithBasicAuth(username, password)(clone)
	return clone
}

func (c *client) WithTLSClientConfig(cert, key string) Client {
	clone := c.Clone().(*client)
	if clone.tlsConfig == nil {
		clone.tlsConfig = &tls.Config{}
	}
	if cert != "" && key != "" {
		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return clone
		}
		clone.tlsConfig.Certificates = []tls.Certificate{certificate}
	}
	if clone.cli.Transport == nil {
		clone.cli.Transport = &http.Transport{}
	}
	clone.cli.Transport.(*http.Transport).TLSClientConfig = clone.tlsConfig
	return clone
}

func (c *client) WithRootCA(caPath string) Client {
	clone := c.Clone().(*client)
	if clone.tlsConfig == nil {
		clone.tlsConfig = &tls.Config{}
	}
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return clone
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	clone.tlsConfig.RootCAs = caCertPool
	if clone.cli.Transport == nil {
		clone.cli.Transport = &http.Transport{}
	}
	clone.cli.Transport.(*http.Transport).TLSClientConfig = clone.tlsConfig
	return clone
}

func (c *client) WithServerName(serverName string) Client {
	clone := c.Clone().(*client)
	if clone.tlsConfig == nil {
		clone.tlsConfig = &tls.Config{}
	}
	clone.tlsConfig.ServerName = serverName
	if clone.cli.Transport == nil {
		clone.cli.Transport = &http.Transport{}
	}
	clone.cli.Transport.(*http.Transport).TLSClientConfig = clone.tlsConfig
	return clone
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	if c.ctx != nil {
		req = req.WithContext(c.ctx)
	}
	for k, h := range c.header {
		req.Header.Set(k, h[0])
	}
	return c.cli.Do(req)
}

func (c *client) Get(api string, params ...url.Values) (*http.Response, error) {
	urlParams := url.Values{}
	for _, param := range params {
		for k, v := range param {
			urlParams.Set(k, v[0])
		}
	}
	api = api + "?" + urlParams.Encode()
	req, err := http.NewRequest(http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *client) Post(api string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, api, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *client) PostForm(api string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, api, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}

func (c *client) PostJson(api string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, api, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Do(req)
}
