package polymarket

import (
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/bridge"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/ctf"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/data"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/gamma"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/rtds"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
)

// Client aggregates service clients behind a shared configuration.
type Client struct {
	Config Config

	CLOB   clob.Client
	CLOBWS ws.Client
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

	if c.CLOBWS == nil {
		wsURL := c.Config.BaseURLs.CLOBWS
		if wsURL == "" {
			wsURL = ws.ProdBaseURL
		}
		// Note: Root client initialization of WS might not have signer/apiKey yet.
		// These will be set when WithAuth is called on the root client or CLOB client.
		c.CLOBWS, _ = ws.NewClient(wsURL, nil, nil)
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

// WithAuth returns a new client with auth credentials applied to all sub-clients.
func (c *Client) WithAuth(signer auth.Signer, apiKey *auth.APIKey) *Client {
	if c.CLOB != nil {
		c.CLOB = c.CLOB.WithAuth(signer, apiKey)
	}
	return c
}
