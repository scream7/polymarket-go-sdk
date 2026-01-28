package clob

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/heartbeat"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/rfq"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/ws"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/transport"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// clientImpl implements the Client interface.
type clientImpl struct {
	httpClient     *transport.Client
	signer         auth.Signer
	apiKey         *auth.APIKey
	builderCfg     *auth.BuilderConfig
	cache          *clientCache
	geoblockHost   string
	geoblockClient *transport.Client
	rfq            rfq.Client
	ws             ws.Client
	heartbeat      heartbeat.Client
}

type clientCache struct {
	mu        sync.RWMutex
	tickSizes map[string]string
	feeRates  map[string]int64
	negRisk   map[string]bool
}

func newClientCache() *clientCache {
	return &clientCache{
		tickSizes: make(map[string]string),
		feeRates:  make(map[string]int64),
		negRisk:   make(map[string]bool),
	}
}

// NewClient creates a new CLOB client.
func NewClient(httpClient *transport.Client) Client {
	return NewClientWithGeoblock(httpClient, "")
}

// OfficialSignerURL is the endpoint for the official SDK builder attribution signer.
// Users can override this by providing their own Builder Config.
const OfficialSignerURL = "https://api.your-domain.com/v1/sign-builder"

// NewClientWithGeoblock creates a new CLOB client with an explicit geoblock host.
func NewClientWithGeoblock(httpClient *transport.Client, geoblockHost string) Client {
	if geoblockHost == "" {
		geoblockHost = DefaultGeoblockHost
	}

	// Default Builder Config (Attribution to SDK Maintainer)
	// This acts as a fallback. If the user provides their own config via WithBuilderConfig,
	// this will be overwritten in the returned client or subsequent calls.
	defaultBuilderCfg := &auth.BuilderConfig{
		Remote: &auth.BuilderRemoteConfig{
			Host: OfficialSignerURL,
		},
	}
	
	// Apply to transport immediately
	if httpClient != nil {
		httpClient.SetBuilderConfig(defaultBuilderCfg)
	}

	c := &clientImpl{
		httpClient:     httpClient,
		cache:          newClientCache(),
		geoblockHost:   geoblockHost,
		geoblockClient: nil,
		builderCfg:     defaultBuilderCfg, // Set default
		rfq:            rfq.NewClient(httpClient),
		heartbeat:      heartbeat.NewClient(httpClient),
	}
	if httpClient != nil {
		c.geoblockClient = httpClient.CloneWithBaseURL(geoblockHost)
	}
	return c
}

func (c *clientImpl) RFQ() rfq.Client {
	return c.rfq
}

func (c *clientImpl) WS() ws.Client {
	return c.ws
}

func (c *clientImpl) Heartbeat() heartbeat.Client {
	return c.heartbeat
}

// WithAuth returns a new client with authentication capabilities.
func (c *clientImpl) WithAuth(signer auth.Signer, apiKey *auth.APIKey) Client {
	// Update transport with auth credentials for L2 headers
	c.httpClient.SetAuth(signer, apiKey)

	return &clientImpl{
		httpClient:     c.httpClient,
		signer:         signer,
		apiKey:         apiKey,
		builderCfg:     c.builderCfg,
		cache:          c.cache,
		geoblockHost:   c.geoblockHost,
		geoblockClient: c.geoblockClient,
		rfq:            c.rfq,
		ws:             c.ws,
		heartbeat:      c.heartbeat,
	}
}
// WithBuilderConfig returns a new client configured with builder authentication.
func (c *clientImpl) WithBuilderConfig(config *auth.BuilderConfig) Client {
	c.httpClient.SetBuilderConfig(config)

	return &clientImpl{
		httpClient:     c.httpClient,
		signer:         c.signer,
		apiKey:         c.apiKey,
		builderCfg:     config,
		cache:          c.cache,
		geoblockHost:   c.geoblockHost,
		geoblockClient: c.geoblockClient,
		rfq:            c.rfq,
		ws:             c.ws,
	heartbeat:      c.heartbeat,
	}
}

// WithUseServerTime enables or disables server-time signing for this client.
func (c *clientImpl) WithUseServerTime(use bool) Client {
	c.httpClient.SetUseServerTime(use)
	return &clientImpl{
		httpClient:     c.httpClient,
		signer:         c.signer,
		apiKey:         c.apiKey,
		builderCfg:     c.builderCfg,
		cache:          c.cache,
		geoblockHost:   c.geoblockHost,
		geoblockClient: c.geoblockClient,
		rfq:            c.rfq,
		ws:             c.ws,
	heartbeat:      c.heartbeat,
	}
}

// WithGeoblockHost sets the geoblock host and returns a new client.
func (c *clientImpl) WithGeoblockHost(host string) Client {
	if host == "" {
		host = DefaultGeoblockHost
	}
	var geoblockClient *transport.Client
	if c.httpClient != nil {
		geoblockClient = c.httpClient.CloneWithBaseURL(host)
	}
	return &clientImpl{
		httpClient:     c.httpClient,
		signer:         c.signer,
		apiKey:         c.apiKey,
		builderCfg:     c.builderCfg,
		cache:          c.cache,
		geoblockHost:   host,
		geoblockClient: geoblockClient,
		rfq:            c.rfq,
		ws:             c.ws,
	heartbeat:      c.heartbeat,
	}
}

// WithWS sets the WebSocket client and returns a new client.
func (c *clientImpl) WithWS(ws ws.Client) Client {
	return &clientImpl{
		httpClient:     c.httpClient,
		signer:         c.signer,
		apiKey:         c.apiKey,
		builderCfg:     c.builderCfg,
		cache:          c.cache,
		geoblockHost:   c.geoblockHost,
		geoblockClient: c.geoblockClient,
		rfq:            c.rfq,
		ws:             ws,
	heartbeat:      c.heartbeat,
	}
}

func (c *clientImpl) Health(ctx context.Context) (string, error) {
	var resp struct {
		Status string `json:"status"`
	}
	err := c.httpClient.Get(ctx, "/time", nil, &resp)
	if err != nil {
		return "DOWN", err
	}
	return "UP", nil
}

func (c *clientImpl) Time(ctx context.Context) (clobtypes.TimeResponse, error) {
	var ts int64
	err := c.httpClient.Get(ctx, "/time", nil, &ts)
	if err != nil {
		return clobtypes.TimeResponse{}, err
	}
	return clobtypes.TimeResponse{Timestamp: ts}, nil
}

func (c *clientImpl) Geoblock(ctx context.Context) (clobtypes.GeoblockResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	geo := c.geoblockClient
	if geo == nil {
		host := c.geoblockHost
		if host == "" {
			host = DefaultGeoblockHost
		}
		geo = transport.NewClient(nil, host)
	}
	var resp clobtypes.GeoblockResponse
	err := geo.Get(ctx, "/api/geoblock", nil, &resp)
	return resp, err
}

func (c *clientImpl) Markets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Cursor != "" {
			q.Set("cursor", req.Cursor)
		}
		if req.Active != nil {
			q.Set("active", strconv.FormatBool(*req.Active))
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
	}

	var resp clobtypes.MarketsResponse
	err := c.httpClient.Get(ctx, "/markets", q, &resp)
	return resp, err
}

func (c *clientImpl) Market(ctx context.Context, id string) (clobtypes.MarketResponse, error) {
	var resp clobtypes.MarketResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/markets/%s", id), nil, &resp)
	return resp, err
}

func (c *clientImpl) SimplifiedMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error) {
	q := url.Values{}
	// similar param handling...
	var resp clobtypes.MarketsResponse
	err := c.httpClient.Get(ctx, "/simplified-markets", q, &resp)
	return resp, err
}

func (c *clientImpl) SamplingMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error) {
	var resp clobtypes.MarketsResponse
	err := c.httpClient.Get(ctx, "/sampling-markets", nil, &resp)
	return resp, err
}

func (c *clientImpl) SamplingSimplifiedMarkets(ctx context.Context, req *clobtypes.MarketsRequest) (clobtypes.MarketsResponse, error) {
	var resp clobtypes.MarketsResponse
	err := c.httpClient.Get(ctx, "/sampling-simplified-markets", nil, &resp)
	return resp, err
}

func (c *clientImpl) OrderBook(ctx context.Context, req *clobtypes.BookRequest) (clobtypes.OrderBookResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
		if req.Side != "" {
			q.Set("side", req.Side)
		}
	}
	var resp clobtypes.OrderBookResponse
	err := c.httpClient.Get(ctx, "/book", q, &resp)
	return resp, err
}

func (c *clientImpl) OrderBooks(ctx context.Context, req *clobtypes.BooksRequest) (clobtypes.OrderBooksResponse, error) {
	var resp clobtypes.OrderBooksResponse
	var body []map[string]string
	if req != nil {
		body = make([]map[string]string, 0, len(req.TokenIDs))
		for _, id := range req.TokenIDs {
			body = append(body, map[string]string{"token_id": id})
		}
	}
	err := c.httpClient.Post(ctx, "/books", body, &resp)
	return resp, err
}

func (c *clientImpl) Midpoint(ctx context.Context, req *clobtypes.MidpointRequest) (clobtypes.MidpointResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	var resp clobtypes.MidpointResponse
	err := c.httpClient.Get(ctx, "/midpoint", q, &resp)
	return resp, err
}

func (c *clientImpl) Midpoints(ctx context.Context, req *clobtypes.MidpointsRequest) (clobtypes.MidpointsResponse, error) {
	var resp clobtypes.MidpointsResponse
	var body []map[string]string
	if req != nil {
		body = make([]map[string]string, 0, len(req.TokenIDs))
		for _, id := range req.TokenIDs {
			body = append(body, map[string]string{"token_id": id})
		}
	}
	err := c.httpClient.Post(ctx, "/midpoints", body, &resp)
	return resp, err
}

func (c *clientImpl) Price(ctx context.Context, req *clobtypes.PriceRequest) (clobtypes.PriceResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
		if req.Side != "" {
			q.Set("side", req.Side)
		}
	}
	var resp clobtypes.PriceResponse
	err := c.httpClient.Get(ctx, "/price", q, &resp)
	return resp, err
}

func (c *clientImpl) Prices(ctx context.Context, req *clobtypes.PricesRequest) (clobtypes.PricesResponse, error) {
	var resp clobtypes.PricesResponse
	var body []map[string]string
	if req != nil {
		body = make([]map[string]string, 0, len(req.TokenIDs))
		for _, id := range req.TokenIDs {
			entry := map[string]string{"token_id": id}
			if req.Side != "" {
				entry["side"] = req.Side
			}
			body = append(body, entry)
		}
	}
	err := c.httpClient.Post(ctx, "/prices", body, &resp)
	return resp, err
}

func (c *clientImpl) AllPrices(ctx context.Context) (clobtypes.PricesResponse, error) {
	var resp clobtypes.PricesResponse
	err := c.httpClient.Get(ctx, "/prices", nil, &resp)
	return resp, err
}

func (c *clientImpl) Spread(ctx context.Context, req *clobtypes.SpreadRequest) (clobtypes.SpreadResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	var resp clobtypes.SpreadResponse
	err := c.httpClient.Get(ctx, "/spread", q, &resp)
	return resp, err
}

func (c *clientImpl) Spreads(ctx context.Context, req *clobtypes.SpreadsRequest) (clobtypes.SpreadsResponse, error) {
	var resp clobtypes.SpreadsResponse
	var body []map[string]string
	if req != nil {
		body = make([]map[string]string, 0, len(req.TokenIDs))
		for _, id := range req.TokenIDs {
			body = append(body, map[string]string{"token_id": id})
		}
	}
	err := c.httpClient.Post(ctx, "/spreads", body, &resp)
	return resp, err
}

func (c *clientImpl) LastTradePrice(ctx context.Context, req *clobtypes.LastTradePriceRequest) (clobtypes.LastTradePriceResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	var resp clobtypes.LastTradePriceResponse
	err := c.httpClient.Get(ctx, "/last-trade-price", q, &resp)
	return resp, err
}

func (c *clientImpl) LastTradesPrices(ctx context.Context, req *clobtypes.LastTradesPricesRequest) (clobtypes.LastTradesPricesResponse, error) {
	var resp clobtypes.LastTradesPricesResponse
	var body []map[string]string
	if req != nil {
		body = make([]map[string]string, 0, len(req.TokenIDs))
		for _, id := range req.TokenIDs {
			body = append(body, map[string]string{"token_id": id})
		}
	}
	err := c.httpClient.Post(ctx, "/last-trades-prices", body, &resp)
	return resp, err
}

func (c *clientImpl) TickSize(ctx context.Context, req *clobtypes.TickSizeRequest) (clobtypes.TickSizeResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.tickSizes[req.TokenID]; ok && cached != "" {
			c.cache.mu.RUnlock()
			return clobtypes.TickSizeResponse{MinimumTickSize: cached}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp clobtypes.TickSizeResponse
	err := c.httpClient.Get(ctx, "/tick-size", q, &resp)
	if err == nil && req != nil && req.TokenID != "" && c.cache != nil {
		tickSize := resp.MinimumTickSize
		if tickSize == "" {
			tickSize = resp.TickSize
		}
		if tickSize != "" {
			c.cache.mu.Lock()
			c.cache.tickSizes[req.TokenID] = tickSize
			c.cache.mu.Unlock()
		}
	}
	return resp, err
}

func (c *clientImpl) NegRisk(ctx context.Context, req *clobtypes.NegRiskRequest) (clobtypes.NegRiskResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.negRisk[req.TokenID]; ok {
			c.cache.mu.RUnlock()
			return clobtypes.NegRiskResponse{NegRisk: cached}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp clobtypes.NegRiskResponse
	err := c.httpClient.Get(ctx, "/neg-risk", q, &resp)
	if err == nil && req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.Lock()
		c.cache.negRisk[req.TokenID] = resp.NegRisk
		c.cache.mu.Unlock()
	}
	return resp, err
}

func (c *clientImpl) FeeRate(ctx context.Context, req *clobtypes.FeeRateRequest) (clobtypes.FeeRateResponse, error) {
	q := url.Values{}
	if req != nil && req.TokenID != "" {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.feeRates[req.TokenID]; ok {
			c.cache.mu.RUnlock()
			return clobtypes.FeeRateResponse{BaseFee: int(cached)}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp clobtypes.FeeRateResponse
	err := c.httpClient.Get(ctx, "/fee-rate", q, &resp)
	if err == nil && req != nil && req.TokenID != "" && c.cache != nil {
		fee := int64(resp.BaseFee)
		if fee == 0 && resp.FeeRate != "" {
			if parsed, parseErr := strconv.ParseInt(resp.FeeRate, 10, 64); parseErr == nil {
				fee = parsed
			}
		}
		if fee > 0 {
			c.cache.mu.Lock()
			c.cache.feeRates[req.TokenID] = fee
			c.cache.mu.Unlock()
		}
	}
	return resp, err
}

func (c *clientImpl) InvalidateCaches() {
	if c.cache == nil {
		return
	}
	c.cache.mu.Lock()
	c.cache.tickSizes = make(map[string]string)
	c.cache.feeRates = make(map[string]int64)
	c.cache.negRisk = make(map[string]bool)
	c.cache.mu.Unlock()
}

func (c *clientImpl) SetTickSize(tokenID, tickSize string) {
	if c.cache == nil || tokenID == "" || tickSize == "" {
		return
	}
	c.cache.mu.Lock()
	c.cache.tickSizes[tokenID] = tickSize
	c.cache.mu.Unlock()
}

func (c *clientImpl) SetNegRisk(tokenID string, negRisk bool) {
	if c.cache == nil || tokenID == "" {
		return
	}
	c.cache.mu.Lock()
	c.cache.negRisk[tokenID] = negRisk
	c.cache.mu.Unlock()
}

func (c *clientImpl) SetFeeRateBps(tokenID string, feeRateBps int64) {
	if c.cache == nil || tokenID == "" || feeRateBps <= 0 {
		return
	}
	c.cache.mu.Lock()
	c.cache.feeRates[tokenID] = feeRateBps
	c.cache.mu.Unlock()
}

func (c *clientImpl) PricesHistory(ctx context.Context, req *clobtypes.PricesHistoryRequest) (clobtypes.PricesHistoryResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
		if req.StartTs > 0 {
			q.Set("start_ts", strconv.FormatInt(req.StartTs, 10))
		}
		if req.EndTs > 0 {
			q.Set("end_ts", strconv.FormatInt(req.EndTs, 10))
		}
		if req.Resolution != "" {
			q.Set("resolution", req.Resolution)
		}
	}
	var resp clobtypes.PricesHistoryResponse
	err := c.httpClient.Get(ctx, "/prices-history", q, &resp)
	return resp, err
}

// CreateOrder builds and signs an order, then posts it to the CLOB.
// This is a higher-level helper that combines signing and posting.
func (c *clientImpl) CreateOrder(ctx context.Context, order *clobtypes.Order) (clobtypes.OrderResponse, error) {
	return c.CreateOrderWithOptions(ctx, order, nil)
}

func (c *clientImpl) CreateOrderWithOptions(ctx context.Context, order *clobtypes.Order, opts *clobtypes.OrderOptions) (clobtypes.OrderResponse, error) {
	signed, err := c.signOrder(order)
	if err != nil {
		return clobtypes.OrderResponse{}, err
	}
	if opts != nil {
		signed.OrderType = opts.OrderType
		signed.PostOnly = opts.PostOnly
		signed.DeferExec = opts.DeferExec
	}
	return c.PostOrder(ctx, signed)
}

func (c *clientImpl) CreateOrderFromSignable(ctx context.Context, order *clobtypes.SignableOrder) (clobtypes.OrderResponse, error) {
	if order == nil || order.Order == nil {
		return clobtypes.OrderResponse{}, fmt.Errorf("order is required")
	}
	opts := &clobtypes.OrderOptions{
		OrderType: order.OrderType,
		PostOnly:  order.PostOnly,
	}
	return c.CreateOrderWithOptions(ctx, order.Order, opts)
}

func (c *clientImpl) signOrder(order *clobtypes.Order) (*clobtypes.SignedOrder, error) {
	return signOrderWithCreds(c.signer, c.apiKey, order)
}

// SignOrder builds an EIP-712 signature for the given order without posting it.
func SignOrder(signer auth.Signer, apiKey *auth.APIKey, order *clobtypes.Order) (*clobtypes.SignedOrder, error) {
	return signOrderWithCreds(signer, apiKey, order)
}

func signOrderWithCreds(signer auth.Signer, apiKey *auth.APIKey, order *clobtypes.Order) (*clobtypes.SignedOrder, error) {
	if signer == nil {
		return nil, auth.ErrMissingSigner
	}
	if apiKey == nil {
		return nil, auth.ErrMissingCreds
	}
	if order == nil {
		return nil, fmt.Errorf("order is required")
	}

	domain := &apitypes.TypedDataDomain{
		Name:              "Polymarket CTF Exchange",
		Version:           "1",
		ChainId:           (*math.HexOrDecimal256)(signer.ChainID()),
		VerifyingContract: "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E", // Exchange Contract Address (Mainnet)
	}

	typesDef := apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"clobtypes.Order": {
			{Name: "salt", Type: "uint256"},
			{Name: "maker", Type: "address"},
			{Name: "signer", Type: "address"},
			{Name: "taker", Type: "address"},
			{Name: "tokenId", Type: "uint256"},
			{Name: "makerAmount", Type: "uint256"},
			{Name: "takerAmount", Type: "uint256"},
			{Name: "expiration", Type: "uint256"},
			{Name: "nonce", Type: "uint256"},
			{Name: "feeRateBps", Type: "uint256"},
			{Name: "side", Type: "uint8"},
			{Name: "signatureType", Type: "uint8"},
		},
	}

	sideInt := 0
	if strings.ToUpper(order.Side) == "SELL" {
		sideInt = 1
	}

	if order.Salt.Int == nil || order.Salt.Int.Sign() == 0 {
		salt, err := generateSalt()
		if err != nil {
			return nil, err
		}
		order.Salt = types.U256{Int: salt}
	}

	sigType := 0
	if order.SignatureType != nil {
		sigType = *order.SignatureType
	}

	expiration := big.NewInt(0)
	if order.Expiration.Int != nil {
		expiration = order.Expiration.Int
	}

	message := apitypes.TypedDataMessage{
		"salt":          (*math.HexOrDecimal256)(order.Salt.Int),
		"maker":         order.Maker.String(),
		"signer":        signer.Address().String(),
		"taker":         order.Taker.String(),
		"tokenId":       (*math.HexOrDecimal256)(order.TokenID.Int),
		"makerAmount":   (*math.HexOrDecimal256)(order.MakerAmount.BigInt()),
		"takerAmount":   (*math.HexOrDecimal256)(order.TakerAmount.BigInt()),
		"expiration":    (*math.HexOrDecimal256)(expiration),
		"nonce":         (*math.HexOrDecimal256)(order.Nonce.Int),
		"feeRateBps":    (*math.HexOrDecimal256)(order.FeeRateBps.BigInt()),
		"side":          (*math.HexOrDecimal256)(big.NewInt(int64(sideInt))),
		"signatureType": (*math.HexOrDecimal256)(big.NewInt(int64(sigType))),
	}

	sig, err := signer.SignTypedData(domain, typesDef, message, "clobtypes.Order")
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	owner := apiKey.Key
	if owner == "" {
		owner = signer.Address().String()
	}

	return &clobtypes.SignedOrder{
		Order:     *order,
		Signature: hexutil.Encode(sig),
		Owner:     owner,
	}, nil
}

func (c *clientImpl) PostOrder(ctx context.Context, req *clobtypes.SignedOrder) (clobtypes.OrderResponse, error) {
	var resp clobtypes.OrderResponse
	payload, err := buildOrderPayload(req)
	if err != nil {
		return resp, err
	}
	err = c.httpClient.Post(ctx, "/order", payload, &resp)
	return resp, err
}

func (c *clientImpl) PostOrders(ctx context.Context, req *clobtypes.SignedOrders) (clobtypes.PostOrdersResponse, error) {
	var resp clobtypes.PostOrdersResponse
	payload, err := buildOrdersPayload(req)
	if err != nil {
		return resp, err
	}
	err = c.httpClient.Post(ctx, "/orders", payload, &resp)
	return resp, err
}

func (c *clientImpl) CancelOrder(ctx context.Context, req *clobtypes.CancelOrderRequest) (clobtypes.CancelResponse, error) {
	var resp clobtypes.CancelResponse
	err := c.httpClient.Delete(ctx, "/order", req, &resp)
	return resp, err
}

func (c *clientImpl) CancelOrders(ctx context.Context, req *clobtypes.CancelOrdersRequest) (clobtypes.CancelResponse, error) {
	var resp clobtypes.CancelResponse
	err := c.httpClient.Delete(ctx, "/orders", req, &resp)
	return resp, err
}

func (c *clientImpl) CancelAll(ctx context.Context) (clobtypes.CancelAllResponse, error) {
	var resp clobtypes.CancelAllResponse
	err := c.httpClient.Delete(ctx, "/cancel-all", nil, &resp)
	return resp, err
}

func (c *clientImpl) CancelMarketOrders(ctx context.Context, req *clobtypes.CancelMarketOrdersRequest) (clobtypes.CancelMarketOrdersResponse, error) {
	var resp clobtypes.CancelMarketOrdersResponse
	err := c.httpClient.Delete(ctx, "/cancel-market-orders", req, &resp)
	return resp, err
}

func (c *clientImpl) Order(ctx context.Context, id string) (clobtypes.OrderResponse, error) {
	var resp clobtypes.OrderResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/data/order/%s", id), nil, &resp)
	return resp, err
}

func (c *clientImpl) Orders(ctx context.Context, req *clobtypes.OrdersRequest) (clobtypes.OrdersResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.ID != "" {
			q.Set("id", req.ID)
		}
		if req.Market != "" {
			q.Set("market", req.Market)
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		nextCursor := req.NextCursor
		if nextCursor == "" {
			nextCursor = req.Cursor
		}
		if nextCursor != "" {
			q.Set("next_cursor", nextCursor)
		}
	}
	var resp clobtypes.OrdersResponse
	err := c.httpClient.Get(ctx, "/data/orders", q, &resp)
	return resp, err
}

func (c *clientImpl) Trades(ctx context.Context, req *clobtypes.TradesRequest) (clobtypes.TradesResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.ID != "" {
			q.Set("id", req.ID)
		}
		if req.Taker != "" {
			q.Set("taker", req.Taker)
		}
		if req.Maker != "" {
			q.Set("maker", req.Maker)
		}
		if req.Market != "" {
			q.Set("market", req.Market)
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
		if req.Before > 0 {
			q.Set("before", strconv.FormatInt(req.Before, 10))
		}
		if req.After > 0 {
			q.Set("after", strconv.FormatInt(req.After, 10))
		}
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		nextCursor := req.NextCursor
		if nextCursor == "" {
			nextCursor = req.Cursor
		}
		if nextCursor != "" {
			q.Set("next_cursor", nextCursor)
		}
	}
	var resp clobtypes.TradesResponse
	err := c.httpClient.Get(ctx, "/data/trades", q, &resp)
	return resp, err
}

func (c *clientImpl) OrdersAll(ctx context.Context, req *clobtypes.OrdersRequest) ([]clobtypes.OrderResponse, error) {
	var results []clobtypes.OrderResponse
	cursor := clobtypes.InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = clobtypes.InitialCursor
	}

	for cursor != clobtypes.EndCursor {
		nextReq := clobtypes.OrdersRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.NextCursor = cursor

		resp, err := c.Orders(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Data...)

		if resp.NextCursor == "" || resp.NextCursor == cursor {
			break
		}
		cursor = resp.NextCursor
	}

	return results, nil
}

func (c *clientImpl) TradesAll(ctx context.Context, req *clobtypes.TradesRequest) ([]clobtypes.Trade, error) {
	var results []clobtypes.Trade
	cursor := clobtypes.InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = clobtypes.InitialCursor
	}

	for cursor != clobtypes.EndCursor {
		nextReq := clobtypes.TradesRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.NextCursor = cursor

		resp, err := c.Trades(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Data...)

		if resp.NextCursor == "" || resp.NextCursor == cursor {
			break
		}
		cursor = resp.NextCursor
	}

	return results, nil
}

func (c *clientImpl) BuilderTradesAll(ctx context.Context, req *clobtypes.BuilderTradesRequest) ([]clobtypes.Trade, error) {
	var results []clobtypes.Trade
	cursor := clobtypes.InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = clobtypes.InitialCursor
	}

	for cursor != clobtypes.EndCursor {
		nextReq := clobtypes.BuilderTradesRequest{}
		if req != nil {
			nextReq = *req
		}
		nextReq.NextCursor = cursor

		resp, err := c.BuilderTrades(ctx, &nextReq)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Data...)

		if resp.NextCursor == "" || resp.NextCursor == cursor {
			break
		}
		cursor = resp.NextCursor
	}

	return results, nil
}
func (c *clientImpl) OrderScoring(ctx context.Context, req *clobtypes.OrderScoringRequest) (clobtypes.OrderScoringResponse, error) {
	q := url.Values{}
	if req != nil && req.ID != "" {
		q.Set("order_id", req.ID)
	}
	var resp clobtypes.OrderScoringResponse
	err := c.httpClient.Get(ctx, "/order-scoring", q, &resp)
	return resp, err
}
func (c *clientImpl) OrdersScoring(ctx context.Context, req *clobtypes.OrdersScoringRequest) (clobtypes.OrdersScoringResponse, error) {
	var resp clobtypes.OrdersScoringResponse
	var body []string
	if req != nil {
		body = req.IDs
	}
	err := c.httpClient.Post(ctx, "/orders-scoring", body, &resp)
	return resp, err
}
func (c *clientImpl) BalanceAllowance(ctx context.Context, req *clobtypes.BalanceAllowanceRequest) (clobtypes.BalanceAllowanceResponse, error) {
	q := url.Values{}
	if req != nil && req.Asset != "" {
		q.Set("asset", req.Asset)
	}
	var resp clobtypes.BalanceAllowanceResponse
	err := c.httpClient.Get(ctx, "/balance-allowance", q, &resp)
	return resp, err
}

func (c *clientImpl) UpdateBalanceAllowance(ctx context.Context, req *clobtypes.BalanceAllowanceUpdateRequest) (clobtypes.BalanceAllowanceResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Asset != "" {
			q.Set("asset", req.Asset)
		}
		if req.Amount != "" {
			q.Set("amount", req.Amount)
		}
	}
	var resp clobtypes.BalanceAllowanceResponse
	err := c.httpClient.Get(ctx, "/balance-allowance/update", q, &resp)
	return resp, err
}

func (c *clientImpl) Notifications(ctx context.Context, req *clobtypes.NotificationsRequest) (clobtypes.NotificationsResponse, error) {
	q := url.Values{}
	if req != nil && req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	var resp clobtypes.NotificationsResponse
	err := c.httpClient.Get(ctx, "/notifications", q, &resp)
	return resp, err
}

func (c *clientImpl) DropNotifications(ctx context.Context, req *clobtypes.DropNotificationsRequest) (clobtypes.DropNotificationsResponse, error) {
	q := url.Values{}
	if req != nil && req.ID != "" {
		q.Set("id", req.ID)
	}
	var resp clobtypes.DropNotificationsResponse
	var err error
	if len(q) > 0 {
		err = c.httpClient.Call(ctx, "DELETE", "/notifications", q, nil, &resp, nil)
	} else {
		err = c.httpClient.Delete(ctx, "/notifications", nil, &resp)
	}
	return resp, err
}

func (c *clientImpl) UserEarnings(ctx context.Context, req *clobtypes.UserEarningsRequest) (clobtypes.UserEarningsResponse, error) {
	q := url.Values{}
	if req != nil && req.Asset != "" {
		q.Set("asset", req.Asset)
	}
	var resp clobtypes.UserEarningsResponse
	err := c.httpClient.Get(ctx, "/rewards/user", q, &resp)
	return resp, err
}

func (c *clientImpl) UserTotalEarnings(ctx context.Context, req *clobtypes.UserTotalEarningsRequest) (clobtypes.UserTotalEarningsResponse, error) {
	q := url.Values{}
	if req != nil && req.Asset != "" {
		q.Set("asset", req.Asset)
	}
	var resp clobtypes.UserTotalEarningsResponse
	err := c.httpClient.Get(ctx, "/rewards/user/total", q, &resp)
	return resp, err
}

func (c *clientImpl) UserRewardPercentages(ctx context.Context, req *clobtypes.UserRewardPercentagesRequest) (clobtypes.UserRewardPercentagesResponse, error) {
	var resp clobtypes.UserRewardPercentagesResponse
	err := c.httpClient.Get(ctx, "/rewards/user/percentages", nil, &resp)
	return resp, err
}

func (c *clientImpl) RewardsMarketsCurrent(ctx context.Context) (clobtypes.RewardsMarketsResponse, error) {
	var resp clobtypes.RewardsMarketsResponse
	err := c.httpClient.Get(ctx, "/rewards/markets/current", nil, &resp)
	return resp, err
}

func (c *clientImpl) RewardsMarkets(ctx context.Context, id string) (clobtypes.RewardsMarketResponse, error) {
	var resp clobtypes.RewardsMarketResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/rewards/markets/%s", id), nil, &resp)
	return resp, err
}

func (c *clientImpl) UserRewardsByMarket(ctx context.Context, req *clobtypes.UserRewardsByMarketRequest) (clobtypes.UserRewardsByMarketResponse, error) {
	q := url.Values{}
	if req != nil && req.MarketID != "" {
		q.Set("market_id", req.MarketID)
	}
	var resp clobtypes.UserRewardsByMarketResponse
	err := c.httpClient.Get(ctx, "/rewards/user/markets", q, &resp)
	return resp, err
}

func (c *clientImpl) MarketTradesEvents(ctx context.Context, id string) (clobtypes.MarketTradesEventsResponse, error) {
	var resp clobtypes.MarketTradesEventsResponse
	err := c.httpClient.Get(ctx, "/v1/market-trades-events/"+id, nil, &resp)
	return resp, err
}

func (c *clientImpl) CreateAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	if c.signer == nil {
		return clobtypes.APIKeyResponse{}, auth.ErrMissingSigner
	}

	headersRaw, err := auth.BuildL1Headers(c.signer, 0, 0)
	if err != nil {
		return clobtypes.APIKeyResponse{}, err
	}

	headers := map[string]string{
		auth.HeaderPolyAddress:   headersRaw.Get(auth.HeaderPolyAddress),
		auth.HeaderPolyTimestamp: headersRaw.Get(auth.HeaderPolyTimestamp),
		auth.HeaderPolyNonce:     headersRaw.Get(auth.HeaderPolyNonce),
		auth.HeaderPolySignature: headersRaw.Get(auth.HeaderPolySignature),
	}

	var resp clobtypes.APIKeyResponse
	// Note: We use CallWithHeaders to inject L1 headers.
	// clobtypes.CreateAPIKey uses POST /auth/api-key
	err = c.httpClient.CallWithHeaders(ctx, "POST", "/auth/api-key", nil, nil, &resp, headers)
	return resp, err
}

func (c *clientImpl) ListAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error) {
	var resp clobtypes.APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/api-keys", nil, &resp)
	return resp, err
}

func (c *clientImpl) DeleteAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	q := url.Values{}
	if id != "" {
		q.Set("api_key", id)
	}
	if len(q) > 0 {
		err := c.httpClient.Call(ctx, "DELETE", "/auth/api-key", q, nil, &resp, nil)
		return resp, err
	}
	err := c.httpClient.Delete(ctx, "/auth/api-key", nil, &resp)
	return resp, err
}

func (c *clientImpl) DeriveAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	headersRaw, err := auth.BuildL1Headers(c.signer, 0, 0)
	if err != nil {
		return clobtypes.APIKeyResponse{}, err
	}
	headers := map[string]string{
		auth.HeaderPolyAddress:   headersRaw.Get(auth.HeaderPolyAddress),
		auth.HeaderPolyTimestamp: headersRaw.Get(auth.HeaderPolyTimestamp),
		auth.HeaderPolyNonce:     headersRaw.Get(auth.HeaderPolyNonce),
		auth.HeaderPolySignature: headersRaw.Get(auth.HeaderPolySignature),
	}
	err = c.httpClient.CallWithHeaders(ctx, "GET", "/auth/derive-api-key", nil, nil, &resp, headers)
	return resp, err
}
func (c *clientImpl) ClosedOnlyStatus(ctx context.Context) (clobtypes.ClosedOnlyResponse, error) {
	var resp clobtypes.ClosedOnlyResponse
	err := c.httpClient.Get(ctx, "/auth/ban-status/closed-only", nil, &resp)
	return resp, err
}
func (c *clientImpl) CreateReadonlyAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	err := c.httpClient.Post(ctx, "/auth/readonly-api-key", nil, &resp)
	return resp, err
}
func (c *clientImpl) ListReadonlyAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error) {
	var resp clobtypes.APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/readonly-api-keys", nil, &resp)
	return resp, err
}
func (c *clientImpl) DeleteReadonlyAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	body := map[string]string{"key": id}
	err := c.httpClient.Delete(ctx, "/auth/readonly-api-key", body, &resp)
	return resp, err
}
func (c *clientImpl) ValidateReadonlyAPIKey(ctx context.Context, req *clobtypes.ValidateReadonlyAPIKeyRequest) (clobtypes.ValidateReadonlyAPIKeyResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Address != "" {
			q.Set("address", req.Address)
		}
		if req.APIKey != "" {
			q.Set("key", req.APIKey)
		}
	}
	var resp clobtypes.ValidateReadonlyAPIKeyResponse
	err := c.httpClient.Get(ctx, "/auth/validate-readonly-api-key", q, &resp)
	return resp, err
}
func (c *clientImpl) CreateBuilderAPIKey(ctx context.Context) (clobtypes.APIKeyResponse, error) {
	var resp clobtypes.APIKeyResponse
	err := c.httpClient.Post(ctx, "/auth/builder-api-key", nil, &resp)
	return resp, err
}
func (c *clientImpl) ListBuilderAPIKeys(ctx context.Context) (clobtypes.APIKeyListResponse, error) {
	var resp clobtypes.APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/builder-api-key", nil, &resp)
	return resp, err
}
func (c *clientImpl) RevokeBuilderAPIKey(ctx context.Context, id string) (clobtypes.APIKeyResponse, error) {
	// Endpoint returns empty body; ignore response.
	err := c.httpClient.Call(ctx, "DELETE", "/auth/builder-api-key", nil, nil, nil, nil)
	return clobtypes.APIKeyResponse{}, err
}
func (c *clientImpl) BuilderTrades(ctx context.Context, req *clobtypes.BuilderTradesRequest) (clobtypes.BuilderTradesResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.ID != "" {
			q.Set("id", req.ID)
		}
		if req.Maker != "" {
			q.Set("maker", req.Maker)
		}
		if req.Market != "" {
			q.Set("market", req.Market)
		}
		if req.AssetID != "" {
			q.Set("asset_id", req.AssetID)
		}
		if req.Before > 0 {
			q.Set("before", strconv.FormatInt(req.Before, 10))
		}
		if req.After > 0 {
			q.Set("after", strconv.FormatInt(req.After, 10))
		}
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		nextCursor := req.NextCursor
		if nextCursor == "" {
			nextCursor = req.Cursor
		}
		if nextCursor != "" {
			q.Set("next_cursor", nextCursor)
		}
	}
	var resp clobtypes.BuilderTradesResponse
	err := c.httpClient.Get(ctx, "/builder/trades", q, &resp)
	return resp, err
}
