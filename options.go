package polymarket

import (
	"go-polymarket-sdk/pkg/bridge"
	"go-polymarket-sdk/pkg/clob"
	"go-polymarket-sdk/pkg/clobws"
	"go-polymarket-sdk/pkg/ctf"
	"go-polymarket-sdk/pkg/data"
	"go-polymarket-sdk/pkg/gamma"
	"go-polymarket-sdk/pkg/rtds"
	"go-polymarket-sdk/pkg/transport"
)

// Option mutates the root client.
type Option func(*Client)

func WithConfig(cfg Config) Option {
	return func(c *Client) {
		c.Config = cfg
	}
}

func WithHTTPClient(doer transport.Doer) Option {
	return func(c *Client) {
		c.Config.HTTPClient = doer
	}
}

func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.Config.UserAgent = userAgent
	}
}

func WithUseServerTime(use bool) Option {
	return func(c *Client) {
		c.Config.UseServerTime = use
	}
}

func WithCLOB(client clob.Client) Option {
	return func(c *Client) {
		c.CLOB = client
	}
}

func WithCLOBWS(client clobws.Client) Option {
	return func(c *Client) {
		c.CLOBWS = client
	}
}

func WithGamma(client gamma.Client) Option {
	return func(c *Client) {
		c.Gamma = client
	}
}

func WithData(client data.Client) Option {
	return func(c *Client) {
		c.Data = client
	}
}

func WithBridge(client bridge.Client) Option {
	return func(c *Client) {
		c.Bridge = client
	}
}

func WithRTDS(client rtds.Client) Option {
	return func(c *Client) {
		c.RTDS = client
	}
}

func WithCTF(client ctf.Client) Option {
	return func(c *Client) {
		c.CTF = client
	}
}
