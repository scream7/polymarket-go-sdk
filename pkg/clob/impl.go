package clob

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"go-polymarket-sdk/pkg/auth"
	"go-polymarket-sdk/pkg/transport"
	"go-polymarket-sdk/pkg/types"

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

// NewClientWithGeoblock creates a new CLOB client with an explicit geoblock host.
func NewClientWithGeoblock(httpClient *transport.Client, geoblockHost string) Client {
	if geoblockHost == "" {
		geoblockHost = DefaultGeoblockHost
	}
	c := &clientImpl{
		httpClient:     httpClient,
		cache:          newClientCache(),
		geoblockHost:   geoblockHost,
		geoblockClient: nil,
	}
	if httpClient != nil {
		c.geoblockClient = httpClient.CloneWithBaseURL(geoblockHost)
	}
	return c
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

func (c *clientImpl) Time(ctx context.Context) (TimeResponse, error) {
	var ts int64
	err := c.httpClient.Get(ctx, "/time", nil, &ts)
	if err != nil {
		return TimeResponse{}, err
	}
	return TimeResponse{Timestamp: ts}, nil
}

func (c *clientImpl) Geoblock(ctx context.Context) (GeoblockResponse, error) {
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
	var resp GeoblockResponse
	err := geo.Get(ctx, "/api/geoblock", nil, &resp)
	return resp, err
}

func (c *clientImpl) Markets(ctx context.Context, req *MarketsRequest) (MarketsResponse, error) {
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

	var resp MarketsResponse
	err := c.httpClient.Get(ctx, "/markets", q, &resp)
	return resp, err
}

func (c *clientImpl) Market(ctx context.Context, id string) (MarketResponse, error) {
	var resp MarketResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/markets/%s", id), nil, &resp)
	return resp, err
}

func (c *clientImpl) SimplifiedMarkets(ctx context.Context, req *MarketsRequest) (MarketsResponse, error) {
	q := url.Values{}
	// similar param handling...
	var resp MarketsResponse
	err := c.httpClient.Get(ctx, "/simplified-markets", q, &resp)
	return resp, err
}

func (c *clientImpl) SamplingMarkets(ctx context.Context, req *MarketsRequest) (MarketsResponse, error) {
	var resp MarketsResponse
	err := c.httpClient.Get(ctx, "/sampling-markets", nil, &resp)
	return resp, err
}

func (c *clientImpl) SamplingSimplifiedMarkets(ctx context.Context, req *MarketsRequest) (MarketsResponse, error) {
	var resp MarketsResponse
	err := c.httpClient.Get(ctx, "/sampling-simplified-markets", nil, &resp)
	return resp, err
}

func (c *clientImpl) OrderBook(ctx context.Context, req *BookRequest) (OrderBookResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
		if req.Side != "" {
			q.Set("side", req.Side)
		}
	}
	var resp OrderBookResponse
	err := c.httpClient.Get(ctx, "/book", q, &resp)
	return resp, err
}

func (c *clientImpl) OrderBooks(ctx context.Context, req *BooksRequest) (OrderBooksResponse, error) {
	var resp OrderBooksResponse
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

func (c *clientImpl) Midpoint(ctx context.Context, req *MidpointRequest) (MidpointResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	var resp MidpointResponse
	err := c.httpClient.Get(ctx, "/midpoint", q, &resp)
	return resp, err
}

func (c *clientImpl) Midpoints(ctx context.Context, req *MidpointsRequest) (MidpointsResponse, error) {
	var resp MidpointsResponse
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

func (c *clientImpl) Price(ctx context.Context, req *PriceRequest) (PriceResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
		if req.Side != "" {
			q.Set("side", req.Side)
		}
	}
	var resp PriceResponse
	err := c.httpClient.Get(ctx, "/price", q, &resp)
	return resp, err
}

func (c *clientImpl) Prices(ctx context.Context, req *PricesRequest) (PricesResponse, error) {
	var resp PricesResponse
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

func (c *clientImpl) AllPrices(ctx context.Context) (PricesResponse, error) {
	var resp PricesResponse
	err := c.httpClient.Get(ctx, "/prices", nil, &resp)
	return resp, err
}

func (c *clientImpl) Spread(ctx context.Context, req *SpreadRequest) (SpreadResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	var resp SpreadResponse
	err := c.httpClient.Get(ctx, "/spread", q, &resp)
	return resp, err
}

func (c *clientImpl) Spreads(ctx context.Context, req *SpreadsRequest) (SpreadsResponse, error) {
	var resp SpreadsResponse
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

func (c *clientImpl) LastTradePrice(ctx context.Context, req *LastTradePriceRequest) (LastTradePriceResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	var resp LastTradePriceResponse
	err := c.httpClient.Get(ctx, "/last-trade-price", q, &resp)
	return resp, err
}

func (c *clientImpl) LastTradesPrices(ctx context.Context, req *LastTradesPricesRequest) (LastTradesPricesResponse, error) {
	var resp LastTradesPricesResponse
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

func (c *clientImpl) TickSize(ctx context.Context, req *TickSizeRequest) (TickSizeResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.tickSizes[req.TokenID]; ok && cached != "" {
			c.cache.mu.RUnlock()
			return TickSizeResponse{MinimumTickSize: cached}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp TickSizeResponse
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

func (c *clientImpl) NegRisk(ctx context.Context, req *NegRiskRequest) (NegRiskResponse, error) {
	q := url.Values{}
	if req != nil {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.negRisk[req.TokenID]; ok {
			c.cache.mu.RUnlock()
			return NegRiskResponse{NegRisk: cached}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp NegRiskResponse
	err := c.httpClient.Get(ctx, "/neg-risk", q, &resp)
	if err == nil && req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.Lock()
		c.cache.negRisk[req.TokenID] = resp.NegRisk
		c.cache.mu.Unlock()
	}
	return resp, err
}

func (c *clientImpl) FeeRate(ctx context.Context, req *FeeRateRequest) (FeeRateResponse, error) {
	q := url.Values{}
	if req != nil && req.TokenID != "" {
		q.Set("token_id", req.TokenID)
	}
	if req != nil && req.TokenID != "" && c.cache != nil {
		c.cache.mu.RLock()
		if cached, ok := c.cache.feeRates[req.TokenID]; ok {
			c.cache.mu.RUnlock()
			return FeeRateResponse{BaseFee: int(cached)}, nil
		}
		c.cache.mu.RUnlock()
	}
	var resp FeeRateResponse
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

func (c *clientImpl) PricesHistory(ctx context.Context, req *PricesHistoryRequest) (PricesHistoryResponse, error) {
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
	var resp PricesHistoryResponse
	err := c.httpClient.Get(ctx, "/prices-history", q, &resp)
	return resp, err
}

// CreateOrder builds and signs an order, then posts it to the CLOB.
// This is a higher-level helper that combines signing and posting.
func (c *clientImpl) CreateOrder(ctx context.Context, order *Order) (OrderResponse, error) {
	return c.CreateOrderWithOptions(ctx, order, nil)
}

func (c *clientImpl) CreateOrderWithOptions(ctx context.Context, order *Order, opts *OrderOptions) (OrderResponse, error) {
	signed, err := c.signOrder(order)
	if err != nil {
		return OrderResponse{}, err
	}
	if opts != nil {
		signed.OrderType = opts.OrderType
		signed.PostOnly = opts.PostOnly
		signed.DeferExec = opts.DeferExec
	}
	return c.PostOrder(ctx, signed)
}

func (c *clientImpl) CreateOrderFromSignable(ctx context.Context, order *SignableOrder) (OrderResponse, error) {
	if order == nil || order.Order == nil {
		return OrderResponse{}, fmt.Errorf("order is required")
	}
	opts := &OrderOptions{
		OrderType: order.OrderType,
		PostOnly:  order.PostOnly,
	}
	return c.CreateOrderWithOptions(ctx, order.Order, opts)
}

func (c *clientImpl) signOrder(order *Order) (*SignedOrder, error) {
	return signOrderWithCreds(c.signer, c.apiKey, order)
}

// SignOrder builds an EIP-712 signature for the given order without posting it.
func SignOrder(signer auth.Signer, apiKey *auth.APIKey, order *Order) (*SignedOrder, error) {
	return signOrderWithCreds(signer, apiKey, order)
}

func signOrderWithCreds(signer auth.Signer, apiKey *auth.APIKey, order *Order) (*SignedOrder, error) {
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
		"Order": {
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

	sig, err := signer.SignTypedData(domain, typesDef, message, "Order")
	if err != nil {
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	owner := apiKey.Key
	if owner == "" {
		owner = signer.Address().String()
	}

	return &SignedOrder{
		Order:     *order,
		Signature: hexutil.Encode(sig),
		Owner:     owner,
	}, nil
}

func (c *clientImpl) PostOrder(ctx context.Context, req *SignedOrder) (OrderResponse, error) {
	var resp OrderResponse
	payload, err := buildOrderPayload(req)
	if err != nil {
		return resp, err
	}
	err = c.httpClient.Post(ctx, "/order", payload, &resp)
	return resp, err
}

func (c *clientImpl) PostOrders(ctx context.Context, req *SignedOrders) (PostOrdersResponse, error) {
	var resp PostOrdersResponse
	payload, err := buildOrdersPayload(req)
	if err != nil {
		return resp, err
	}
	err = c.httpClient.Post(ctx, "/orders", payload, &resp)
	return resp, err
}

func (c *clientImpl) CancelOrder(ctx context.Context, req *CancelOrderRequest) (CancelResponse, error) {
	var resp CancelResponse
	err := c.httpClient.Delete(ctx, "/order", req, &resp)
	return resp, err
}

func (c *clientImpl) CancelOrders(ctx context.Context, req *CancelOrdersRequest) (CancelResponse, error) {
	var resp CancelResponse
	err := c.httpClient.Delete(ctx, "/orders", req, &resp)
	return resp, err
}

func (c *clientImpl) CancelAll(ctx context.Context) (CancelAllResponse, error) {
	var resp CancelAllResponse
	err := c.httpClient.Delete(ctx, "/cancel-all", nil, &resp)
	return resp, err
}

func (c *clientImpl) CancelMarketOrders(ctx context.Context, req *CancelMarketOrdersRequest) (CancelMarketOrdersResponse, error) {
	var resp CancelMarketOrdersResponse
	err := c.httpClient.Delete(ctx, "/cancel-market-orders", req, &resp)
	return resp, err
}

func (c *clientImpl) Order(ctx context.Context, id string) (OrderResponse, error) {
	var resp OrderResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/data/order/%s", id), nil, &resp)
	return resp, err
}

func (c *clientImpl) Orders(ctx context.Context, req *OrdersRequest) (OrdersResponse, error) {
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
	var resp OrdersResponse
	err := c.httpClient.Get(ctx, "/data/orders", q, &resp)
	return resp, err
}

func (c *clientImpl) Trades(ctx context.Context, req *TradesRequest) (TradesResponse, error) {
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
	var resp TradesResponse
	err := c.httpClient.Get(ctx, "/data/trades", q, &resp)
	return resp, err
}

func (c *clientImpl) OrdersAll(ctx context.Context, req *OrdersRequest) ([]OrderResponse, error) {
	var results []OrderResponse
	cursor := InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = InitialCursor
	}

	for cursor != EndCursor {
		nextReq := OrdersRequest{}
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

func (c *clientImpl) TradesAll(ctx context.Context, req *TradesRequest) ([]Trade, error) {
	var results []Trade
	cursor := InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = InitialCursor
	}

	for cursor != EndCursor {
		nextReq := TradesRequest{}
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

func (c *clientImpl) BuilderTradesAll(ctx context.Context, req *BuilderTradesRequest) ([]Trade, error) {
	var results []Trade
	cursor := InitialCursor
	if req != nil {
		if req.NextCursor != "" {
			cursor = req.NextCursor
		} else if req.Cursor != "" {
			cursor = req.Cursor
		}
	}
	if cursor == "" {
		cursor = InitialCursor
	}

	for cursor != EndCursor {
		nextReq := BuilderTradesRequest{}
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
func (c *clientImpl) OrderScoring(ctx context.Context, req *OrderScoringRequest) (OrderScoringResponse, error) {
	q := url.Values{}
	if req != nil && req.ID != "" {
		q.Set("order_id", req.ID)
	}
	var resp OrderScoringResponse
	err := c.httpClient.Get(ctx, "/order-scoring", q, &resp)
	return resp, err
}
func (c *clientImpl) OrdersScoring(ctx context.Context, req *OrdersScoringRequest) (OrdersScoringResponse, error) {
	var resp OrdersScoringResponse
	var body []string
	if req != nil {
		body = req.IDs
	}
	err := c.httpClient.Post(ctx, "/orders-scoring", body, &resp)
	return resp, err
}
func (c *clientImpl) BalanceAllowance(ctx context.Context, req *BalanceAllowanceRequest) (BalanceAllowanceResponse, error) {
	q := url.Values{}
	if req != nil && req.Asset != "" {
		q.Set("asset", req.Asset)
	}
	var resp BalanceAllowanceResponse
	err := c.httpClient.Get(ctx, "/balance-allowance", q, &resp)
	return resp, err
}

func (c *clientImpl) UpdateBalanceAllowance(ctx context.Context, req *BalanceAllowanceUpdateRequest) (BalanceAllowanceResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Asset != "" {
			q.Set("asset", req.Asset)
		}
		if req.Amount != "" {
			q.Set("amount", req.Amount)
		}
	}
	var resp BalanceAllowanceResponse
	err := c.httpClient.Get(ctx, "/balance-allowance/update", q, &resp)
	return resp, err
}

func (c *clientImpl) Notifications(ctx context.Context, req *NotificationsRequest) (NotificationsResponse, error) {
	q := url.Values{}
	if req != nil && req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	var resp NotificationsResponse
	err := c.httpClient.Get(ctx, "/notifications", q, &resp)
	return resp, err
}

func (c *clientImpl) DropNotifications(ctx context.Context, req *DropNotificationsRequest) (DropNotificationsResponse, error) {
	q := url.Values{}
	if req != nil && req.ID != "" {
		q.Set("id", req.ID)
	}
	var resp DropNotificationsResponse
	var err error
	if len(q) > 0 {
		err = c.httpClient.Call(ctx, "DELETE", "/notifications", q, nil, &resp, nil)
	} else {
		err = c.httpClient.Delete(ctx, "/notifications", nil, &resp)
	}
	return resp, err
}

func (c *clientImpl) UserEarnings(ctx context.Context, req *UserEarningsRequest) (UserEarningsResponse, error) {
	q := url.Values{}
	if req != nil && req.Asset != "" {
		q.Set("asset", req.Asset)
	}
	var resp UserEarningsResponse
	err := c.httpClient.Get(ctx, "/rewards/user", q, &resp)
	return resp, err
}

func (c *clientImpl) UserTotalEarnings(ctx context.Context, req *UserTotalEarningsRequest) (UserTotalEarningsResponse, error) {
	q := url.Values{}
	if req != nil && req.Asset != "" {
		q.Set("asset", req.Asset)
	}
	var resp UserTotalEarningsResponse
	err := c.httpClient.Get(ctx, "/rewards/user/total", q, &resp)
	return resp, err
}

func (c *clientImpl) UserRewardPercentages(ctx context.Context, req *UserRewardPercentagesRequest) (UserRewardPercentagesResponse, error) {
	var resp UserRewardPercentagesResponse
	err := c.httpClient.Get(ctx, "/rewards/user/percentages", nil, &resp)
	return resp, err
}

func (c *clientImpl) RewardsMarketsCurrent(ctx context.Context) (RewardsMarketsResponse, error) {
	var resp RewardsMarketsResponse
	err := c.httpClient.Get(ctx, "/rewards/markets/current", nil, &resp)
	return resp, err
}

func (c *clientImpl) RewardsMarkets(ctx context.Context, id string) (RewardsMarketResponse, error) {
	var resp RewardsMarketResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/rewards/markets/%s", id), nil, &resp)
	return resp, err
}

func (c *clientImpl) UserRewardsByMarket(ctx context.Context, req *UserRewardsByMarketRequest) (UserRewardsByMarketResponse, error) {
	q := url.Values{}
	if req != nil && req.MarketID != "" {
		q.Set("market_id", req.MarketID)
	}
	var resp UserRewardsByMarketResponse
	err := c.httpClient.Get(ctx, "/rewards/user/markets", q, &resp)
	return resp, err
}

func (c *clientImpl) MarketTradesEvents(ctx context.Context, id string) (MarketTradesEventsResponse, error) {
	var resp MarketTradesEventsResponse
	err := c.httpClient.Get(ctx, fmt.Sprintf("/live-activity/events/%s", id), nil, &resp)
	return resp, err
}

func (c *clientImpl) Heartbeat(ctx context.Context, req *HeartbeatRequest) (HeartbeatResponse, error) {
	var resp HeartbeatResponse
	var body interface{}
	if req != nil && req.HeartbeatID != "" {
		body = map[string]string{"heartbeat_id": req.HeartbeatID}
	} else {
		body = map[string]interface{}{"heartbeat_id": nil}
	}
	err := c.httpClient.Post(ctx, "/v1/heartbeats", body, &resp)
	return resp, err
}

func (c *clientImpl) CreateAPIKey(ctx context.Context) (APIKeyResponse, error) {
	if c.signer == nil {
		return APIKeyResponse{}, auth.ErrMissingSigner
	}

	headersRaw, err := auth.BuildL1Headers(c.signer, 0, 0)
	if err != nil {
		return APIKeyResponse{}, err
	}

	headers := map[string]string{
		auth.HeaderPolyAddress:   headersRaw.Get(auth.HeaderPolyAddress),
		auth.HeaderPolyTimestamp: headersRaw.Get(auth.HeaderPolyTimestamp),
		auth.HeaderPolyNonce:     headersRaw.Get(auth.HeaderPolyNonce),
		auth.HeaderPolySignature: headersRaw.Get(auth.HeaderPolySignature),
	}

	var resp APIKeyResponse
	// Note: We use CallWithHeaders to inject L1 headers.
	// CreateAPIKey uses POST /auth/api-key
	err = c.httpClient.CallWithHeaders(ctx, "POST", "/auth/api-key", nil, nil, &resp, headers)
	return resp, err
}

func (c *clientImpl) ListAPIKeys(ctx context.Context) (APIKeyListResponse, error) {
	var resp APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/api-keys", nil, &resp)
	return resp, err
}

func (c *clientImpl) DeleteAPIKey(ctx context.Context, id string) (APIKeyResponse, error) {
	var resp APIKeyResponse
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

func (c *clientImpl) DeriveAPIKey(ctx context.Context) (APIKeyResponse, error) {
	var resp APIKeyResponse
	headersRaw, err := auth.BuildL1Headers(c.signer, 0, 0)
	if err != nil {
		return APIKeyResponse{}, err
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
func (c *clientImpl) ClosedOnlyStatus(ctx context.Context) (ClosedOnlyResponse, error) {
	var resp ClosedOnlyResponse
	err := c.httpClient.Get(ctx, "/auth/ban-status/closed-only", nil, &resp)
	return resp, err
}
func (c *clientImpl) CreateReadonlyAPIKey(ctx context.Context) (APIKeyResponse, error) {
	var resp APIKeyResponse
	err := c.httpClient.Post(ctx, "/auth/readonly-api-key", nil, &resp)
	return resp, err
}
func (c *clientImpl) ListReadonlyAPIKeys(ctx context.Context) (APIKeyListResponse, error) {
	var resp APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/readonly-api-keys", nil, &resp)
	return resp, err
}
func (c *clientImpl) DeleteReadonlyAPIKey(ctx context.Context, id string) (APIKeyResponse, error) {
	var resp APIKeyResponse
	body := map[string]string{"key": id}
	err := c.httpClient.Delete(ctx, "/auth/readonly-api-key", body, &resp)
	return resp, err
}
func (c *clientImpl) ValidateReadonlyAPIKey(ctx context.Context, req *ValidateReadonlyAPIKeyRequest) (ValidateReadonlyAPIKeyResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Address != "" {
			q.Set("address", req.Address)
		}
		if req.APIKey != "" {
			q.Set("key", req.APIKey)
		}
	}
	var resp ValidateReadonlyAPIKeyResponse
	err := c.httpClient.Get(ctx, "/auth/validate-readonly-api-key", q, &resp)
	return resp, err
}
func (c *clientImpl) CreateBuilderAPIKey(ctx context.Context) (APIKeyResponse, error) {
	var resp APIKeyResponse
	err := c.httpClient.Post(ctx, "/auth/builder-api-key", nil, &resp)
	return resp, err
}
func (c *clientImpl) ListBuilderAPIKeys(ctx context.Context) (APIKeyListResponse, error) {
	var resp APIKeyListResponse
	err := c.httpClient.Get(ctx, "/auth/builder-api-key", nil, &resp)
	return resp, err
}
func (c *clientImpl) RevokeBuilderAPIKey(ctx context.Context, id string) (APIKeyResponse, error) {
	// Endpoint returns empty body; ignore response.
	err := c.httpClient.Call(ctx, "DELETE", "/auth/builder-api-key", nil, nil, nil, nil)
	return APIKeyResponse{}, err
}
func (c *clientImpl) BuilderTrades(ctx context.Context, req *BuilderTradesRequest) (BuilderTradesResponse, error) {
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
	var resp BuilderTradesResponse
	err := c.httpClient.Get(ctx, "/builder/trades", q, &resp)
	return resp, err
}
func (c *clientImpl) CreateRFQRequest(ctx context.Context, req *RFQRequest) (RFQRequestResponse, error) {
	var resp RFQRequestResponse
	err := c.httpClient.Post(ctx, "/rfq/request", req, &resp)
	return resp, err
}

func (c *clientImpl) CancelRFQRequest(ctx context.Context, req *RFQCancelRequest) (RFQCancelResponse, error) {
	var resp RFQCancelResponse
	err := c.httpClient.Delete(ctx, "/rfq/request", req, &resp)
	return resp, err
}

func (c *clientImpl) RFQRequests(ctx context.Context, req *RFQRequestsQuery) (RFQRequestsResponse, error) {
	var resp RFQRequestsResponse
	q := url.Values{}
	if req != nil {
		applyRFQPagination(&q, req.Limit, req.Offset, req.Cursor)
		applyRFQFilters(&q, req.State, req.RequestIDs, nil, req.Markets, req.SizeMin, req.SizeMax, req.SizeUsdcMin, req.SizeUsdcMax, req.PriceMin, req.PriceMax, req.SortBy, req.SortDir)
	}
	err := c.httpClient.Get(ctx, "/rfq/data/requests", q, &resp)
	return resp, err
}

func (c *clientImpl) CreateRFQQuote(ctx context.Context, req *RFQQuote) (RFQQuoteResponse, error) {
	var resp RFQQuoteResponse
	err := c.httpClient.Post(ctx, "/rfq/quote", req, &resp)
	return resp, err
}

func (c *clientImpl) CancelRFQQuote(ctx context.Context, req *RFQCancelQuote) (RFQCancelResponse, error) {
	var resp RFQCancelResponse
	err := c.httpClient.Delete(ctx, "/rfq/quote", req, &resp)
	return resp, err
}

func (c *clientImpl) RFQRequesterQuotes(ctx context.Context, req *RFQRequesterQuotesQuery) (RFQQuotesResponse, error) {
	var resp RFQQuotesResponse
	q := url.Values{}
	if req != nil {
		requestIDs := req.RequestIDs
		if req.RequestID != "" {
			q.Set("request_id", req.RequestID)
			if len(requestIDs) == 0 {
				requestIDs = []string{req.RequestID}
			}
		}
		applyRFQPagination(&q, req.Limit, req.Offset, req.Cursor)
		applyRFQFilters(&q, req.State, requestIDs, req.QuoteIDs, req.Markets, req.SizeMin, req.SizeMax, req.SizeUsdcMin, req.SizeUsdcMax, req.PriceMin, req.PriceMax, req.SortBy, req.SortDir)
	}
	err := c.httpClient.Get(ctx, "/rfq/data/requester/quotes", q, &resp)
	return resp, err
}

func (c *clientImpl) RFQQuoterQuotes(ctx context.Context, req *RFQQuoterQuotesQuery) (RFQQuotesResponse, error) {
	var resp RFQQuotesResponse
	q := url.Values{}
	if req != nil {
		requestIDs := req.RequestIDs
		if req.RequestID != "" {
			q.Set("request_id", req.RequestID)
			if len(requestIDs) == 0 {
				requestIDs = []string{req.RequestID}
			}
		}
		applyRFQPagination(&q, req.Limit, req.Offset, req.Cursor)
		applyRFQFilters(&q, req.State, requestIDs, req.QuoteIDs, req.Markets, req.SizeMin, req.SizeMax, req.SizeUsdcMin, req.SizeUsdcMax, req.PriceMin, req.PriceMax, req.SortBy, req.SortDir)
	}
	err := c.httpClient.Get(ctx, "/rfq/data/quoter/quotes", q, &resp)
	return resp, err
}

func (c *clientImpl) RFQBestQuote(ctx context.Context, req *RFQBestQuoteQuery) (RFQBestQuoteResponse, error) {
	var resp RFQBestQuoteResponse
	q := url.Values{}
	if req != nil {
		requestIDs := req.RequestIDs
		if req.RequestID != "" {
			q.Set("request_id", req.RequestID)
			if len(requestIDs) == 0 {
				requestIDs = []string{req.RequestID}
			}
		}
		if len(requestIDs) > 0 {
			q.Set("requestIds", strings.Join(requestIDs, ","))
		}
	}
	err := c.httpClient.Get(ctx, "/rfq/data/best-quote", q, &resp)
	return resp, err
}

func (c *clientImpl) RFQRequestAccept(ctx context.Context, req *RFQAcceptRequest) (RFQAcceptResponse, error) {
	var resp RFQAcceptResponse
	err := c.httpClient.Post(ctx, "/rfq/request/accept", req, &resp)
	return resp, err
}

func (c *clientImpl) RFQQuoteApprove(ctx context.Context, req *RFQApproveQuote) (RFQApproveResponse, error) {
	var resp RFQApproveResponse
	err := c.httpClient.Post(ctx, "/rfq/quote/approve", req, &resp)
	return resp, err
}

func (c *clientImpl) RFQConfig(ctx context.Context) (RFQConfigResponse, error) {
	var resp RFQConfigResponse
	err := c.httpClient.Get(ctx, "/rfq/config", nil, &resp)
	return resp, err
}

func applyRFQPagination(q *url.Values, limit int, offset, cursor string) {
	if q == nil {
		return
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset == "" && cursor != "" {
		offset = cursor
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if offset != "" {
		q.Set("offset", offset)
	}
}

func applyRFQFilters(q *url.Values, state RFQState, requestIDs, quoteIDs, markets []string, sizeMin, sizeMax, sizeUsdcMin, sizeUsdcMax, priceMin, priceMax string, sortBy RFQSortBy, sortDir RFQSortDir) {
	if q == nil {
		return
	}
	if state != "" {
		q.Set("state", string(state))
	}
	if len(requestIDs) > 0 {
		joined := strings.Join(requestIDs, ",")
		q.Set("requestIds", joined)
		q.Set("request_ids", joined)
	}
	if len(quoteIDs) > 0 {
		joined := strings.Join(quoteIDs, ",")
		q.Set("quoteIds", joined)
		q.Set("quote_ids", joined)
	}
	if len(markets) > 0 {
		q.Set("markets", strings.Join(markets, ","))
	}
	if sizeMin != "" {
		q.Set("sizeMin", sizeMin)
		q.Set("size_min", sizeMin)
	}
	if sizeMax != "" {
		q.Set("sizeMax", sizeMax)
		q.Set("size_max", sizeMax)
	}
	if sizeUsdcMin != "" {
		q.Set("sizeUsdcMin", sizeUsdcMin)
		q.Set("size_usdc_min", sizeUsdcMin)
	}
	if sizeUsdcMax != "" {
		q.Set("sizeUsdcMax", sizeUsdcMax)
		q.Set("size_usdc_max", sizeUsdcMax)
	}
	if priceMin != "" {
		q.Set("priceMin", priceMin)
		q.Set("price_min", priceMin)
	}
	if priceMax != "" {
		q.Set("priceMax", priceMax)
		q.Set("price_max", priceMax)
	}
	if sortBy != "" {
		q.Set("sortBy", string(sortBy))
		q.Set("sort_by", string(sortBy))
	}
	if sortDir != "" {
		q.Set("sortDir", string(sortDir))
		q.Set("sort_dir", string(sortDir))
	}
}
