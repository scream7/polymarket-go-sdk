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

// Client aggregates service clients behind a shared configuration.
type Client struct {
	Config Config

	CLOB   clob.Client
	CLOBWS clobws.Client
	Gamma  gamma.Client
	Data   data.Client
	Bridge bridge.Client
	RTDS   rtds.Client
	CTF    ctf.Client
}

// NewClient creates a new root client with optional overrides.
func NewClient(opts ...Option) *Client {
	c := &Client{Config: DefaultConfig()}
	for _, opt := range opts {
		opt(c)
	}

	if c.CLOB == nil {
		clobTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.CLOB)
		clobTransport.SetUserAgent(c.Config.UserAgent)
		clobTransport.SetUseServerTime(c.Config.UseServerTime)
		c.CLOB = clob.NewClientWithGeoblock(clobTransport, c.Config.BaseURLs.Geoblock)
	}

	if c.Gamma == nil {
		gammaTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Gamma)
		gammaTransport.SetUserAgent(c.Config.UserAgent)
		c.Gamma = gamma.NewClient(gammaTransport)
	}

	if c.Data == nil {
		dataTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Data)
		dataTransport.SetUserAgent(c.Config.UserAgent)
		c.Data = data.NewClient(dataTransport)
	}

	if c.Bridge == nil {
		bridgeTransport := transport.NewClient(c.Config.HTTPClient, c.Config.BaseURLs.Bridge)
		bridgeTransport.SetUserAgent(c.Config.UserAgent)
		c.Bridge = bridge.NewClient(bridgeTransport)
	}

	if c.CTF == nil {
		c.CTF = ctf.NewClient()
	}

	return c
}
