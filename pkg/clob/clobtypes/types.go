package clobtypes

import "github.com/GoPolymarket/polymarket-go-sdk/pkg/types"

// OrderType represents time-in-force / order type values.
type OrderType string

const (
	OrderTypeGTC OrderType = "GTC"
	OrderTypeGTD OrderType = "GTD"
	OrderTypeFAK OrderType = "FAK"
	OrderTypeFOK OrderType = "FOK"
)

const (
	InitialCursor = "MA=="
	EndCursor     = "LTE="
)

// Request types.
type (
	MarketsRequest struct {
		Limit   int    `json:"limit,omitempty"`
		Cursor  string `json:"cursor,omitempty"`
		Active  *bool  `json:"active,omitempty"`
		AssetID string `json:"asset_id,omitempty"`
	}
	BookRequest struct {
		TokenID string `json:"token_id"`
		Side    string `json:"side,omitempty"`
	}
	BooksRequest struct {
		TokenIDs []string `json:"token_ids"`
	}
	MidpointRequest struct {
		TokenID string `json:"token_id"`
	}
	MidpointsRequest struct {
		TokenIDs []string `json:"token_ids"`
	}
	PriceRequest struct {
		TokenID string `json:"token_id"`
		Side    string `json:"side,omitempty"`
	}
	PricesRequest struct {
		TokenIDs []string `json:"token_ids"`
		Side     string   `json:"side,omitempty"`
	}
	SpreadRequest struct {
		TokenID string `json:"token_id"`
	}
	SpreadsRequest struct {
		TokenIDs []string `json:"token_ids"`
	}
	LastTradePriceRequest struct {
		TokenID string `json:"token_id"`
	}
	LastTradesPricesRequest struct {
		TokenIDs []string `json:"token_ids"`
	}
	TickSizeRequest struct {
		TokenID string `json:"token_id"`
	}
	NegRiskRequest struct {
		TokenID string `json:"token_id"`
	}
	FeeRateRequest struct {
		TokenID string `json:"token_id"`
	}
	PricesHistoryRequest struct {
		TokenID    string `json:"token_id"`
		StartTs    int64  `json:"start_ts,omitempty"`
		EndTs      int64  `json:"end_ts,omitempty"`
		Resolution string `json:"resolution,omitempty"` // "1m", "1h", "1d"
	}
	SignableOrder struct {
		Order     *Order    `json:"order"`
		OrderType OrderType `json:"order_type"`
		PostOnly  *bool     `json:"post_only,omitempty"`
	}
	OrderOptions struct {
		OrderType OrderType
		PostOnly  *bool
		DeferExec *bool
	}
	SignedOrder struct {
		Order     Order  `json:"order"`
		Signature string `json:"signature"`
		Owner     string `json:"owner"`

		// Options used when submitting the order (not serialized directly).
		OrderType OrderType `json:"-"`
		PostOnly  *bool     `json:"-"`
		DeferExec *bool     `json:"-"`
	}
	SignedOrders struct {
		Orders []SignedOrder `json:"orders"`
	}
	CancelOrderRequest struct {
		ID string `json:"id"`
	}
	CancelOrdersRequest struct {
		IDs []string `json:"ids"`
	}
	CancelMarketOrdersRequest struct {
		MarketID string `json:"market_id"`
	}
	OrdersRequest struct {
		ID         string `json:"id,omitempty"`
		Market     string `json:"market,omitempty"`
		AssetID    string `json:"asset_id,omitempty"`
		Limit      int    `json:"limit,omitempty"`
		Cursor     string `json:"cursor,omitempty"`
		NextCursor string `json:"next_cursor,omitempty"`
	}
	TradesRequest struct {
		ID         string `json:"id,omitempty"`
		Taker      string `json:"taker,omitempty"`
		Maker      string `json:"maker,omitempty"`
		Market     string `json:"market,omitempty"`
		AssetID    string `json:"asset_id,omitempty"`
		Before     int64  `json:"before,omitempty"`
		After      int64  `json:"after,omitempty"`
		Limit      int    `json:"limit,omitempty"`
		Cursor     string `json:"cursor,omitempty"`
		NextCursor string `json:"next_cursor,omitempty"`
	}
	OrderScoringRequest struct {
		ID string `json:"id"`
	}
	OrdersScoringRequest struct {
		IDs []string `json:"ids"`
	}
	BalanceAllowanceRequest struct {
		Asset string `json:"asset,omitempty"`
	}
	BalanceAllowanceUpdateRequest struct {
		Asset  string `json:"asset"`
		Amount string `json:"amount"`
	}
	NotificationsRequest struct {
		Limit int `json:"limit,omitempty"`
	}
	DropNotificationsRequest struct {
		ID string `json:"id"`
	}
	UserEarningsRequest struct {
		Asset string `json:"asset,omitempty"`
	}
	UserTotalEarningsRequest struct {
		Asset string `json:"asset,omitempty"`
	}
	UserRewardPercentagesRequest struct{}
	UserRewardsByMarketRequest   struct {
		MarketID string `json:"market_id"`
	}
	ValidateReadonlyAPIKeyRequest struct {
		Address string `json:"address"`
		APIKey  string `json:"key"`
	}
	BuilderTradesRequest struct {
		ID         string `json:"id,omitempty"`
		Maker      string `json:"maker,omitempty"`
		Market     string `json:"market,omitempty"`
		AssetID    string `json:"asset_id,omitempty"`
		Before     int64  `json:"before,omitempty"`
		After      int64  `json:"after,omitempty"`
		Limit      int    `json:"limit,omitempty"`
		Cursor     string `json:"cursor,omitempty"`
		NextCursor string `json:"next_cursor,omitempty"`
	}
)

// Response types.
type (
	TimeResponse struct {
		ServerTime string `json:"server_time,omitempty"`
		Timestamp  int64  `json:"timestamp"`
	}
	MarketsResponse struct {
		Data       []Market `json:"data"`
		NextCursor string   `json:"next_cursor"`
		Limit      int      `json:"limit"`
		Count      int      `json:"count"`
	}
	MarketResponse     Market
	OrderBookResponse  OrderBook
	OrderBooksResponse []OrderBook
	MidpointResponse   struct {
		Midpoint string `json:"midpoint"`
	}
	MidpointsResponse []MidpointResponse
	PriceResponse     struct {
		Price string `json:"price"`
	}
	PricesResponse []PriceResponse
	SpreadResponse struct {
		Spread string `json:"spread"`
	}
	SpreadsResponse        []SpreadResponse
	LastTradePriceResponse struct {
		Price string `json:"price"`
	}
	LastTradesPricesResponse []LastTradePriceResponse
	TickSizeResponse         struct {
		MinimumTickSize string `json:"minimum_tick_size,omitempty"`
		TickSize        string `json:"tick_size,omitempty"`
	}
	NegRiskResponse struct {
		NegRisk bool `json:"neg_risk"`
	}
	FeeRateResponse struct {
		BaseFee int    `json:"base_fee,omitempty"`
		FeeRate string `json:"fee_rate,omitempty"`
	}
	GeoblockResponse struct {
		Blocked bool   `json:"blocked"`
		IP      string `json:"ip"`
		Country string `json:"country"`
		Region  string `json:"region"`
	}
	PricesHistoryResponse []PriceHistoryPoint
	OrderResponse         struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	PostOrdersResponse []OrderResponse
	OrdersResponse     struct {
		Data       []OrderResponse `json:"data"`
		NextCursor string          `json:"next_cursor"`
		Limit      int             `json:"limit"`
		Count      int             `json:"count"`
	}
	CancelResponse struct {
		Status string `json:"status"`
	}
	CancelAllResponse struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	CancelMarketOrdersResponse struct {
		Status string `json:"status"`
	}
	TradesResponse struct {
		Data       []Trade `json:"data"`
		NextCursor string  `json:"next_cursor"`
		Limit      int     `json:"limit"`
		Count      int     `json:"count"`
	}
	OrderScoringResponse struct {
		Score string `json:"score"`
	}
	OrdersScoringResponse    []OrderScoringResponse
	BalanceAllowanceResponse struct {
		Balance   string `json:"balance"`
		Allowance string `json:"allowance"`
	}
	NotificationsResponse     []Notification
	DropNotificationsResponse struct {
		Status string `json:"status"`
	}
	UserEarningsResponse struct {
		Earnings string `json:"earnings"`
	}
	UserTotalEarningsResponse struct {
		TotalEarnings string `json:"total_earnings"`
	}
	UserRewardPercentagesResponse struct {
		Percentages map[string]string `json:"percentages"`
	}
	RewardsMarketsResponse      []RewardsMarket
	RewardsMarketResponse       RewardsMarket
	UserRewardsByMarketResponse struct {
		Rewards string `json:"rewards"`
	}
	MarketTradesEventsResponse []TradeEvent
	APIKeyResponse             struct {
		APIKey     string `json:"apiKey"`
		Secret     string `json:"secret,omitempty"`
		Passphrase string `json:"passphrase,omitempty"`
	}
	APIKeyListResponse struct {
		APIKeys []APIKeyResponse `json:"apiKeys"`
	}
	ClosedOnlyResponse struct {
		ClosedOnly bool `json:"closed_only"`
	}
	ValidateReadonlyAPIKeyResponse struct {
		Valid bool `json:"valid"`
	}
	BuilderTradesResponse struct {
		Data       []Trade `json:"data"`
		NextCursor string  `json:"next_cursor"`
		Limit      int     `json:"limit"`
		Count      int     `json:"count"`
	}
)

// Auxiliary types.
type (
	Market struct {
		ID          string        `json:"id"`
		Question    string        `json:"question"`
		ConditionID string        `json:"condition_id"`
		Slug        string        `json:"slug"`
		Resolution  string        `json:"resolution"`
		EndDate     string        `json:"end_date"`
		Tokens      []MarketToken `json:"tokens"`
		// Add minimal fields to match "Simplified" or "Active"
		Active bool `json:"active"`
		Closed bool `json:"closed"`
	}

	MarketToken struct {
		TokenID string  `json:"token_id"`
		Outcome string  `json:"outcome"`
		Price   float64 `json:"price"`
	}

	OrderBook struct {
		MarketID string       `json:"market_id"`
		Bids     []PriceLevel `json:"bids"`
		Asks     []PriceLevel `json:"asks"`
		Hash     string       `json:"hash"`
	}

	PriceLevel struct {
		Price string `json:"price"`
		Size  string `json:"size"`
	}

	Order struct {
		// Define order fields
		Salt          types.U256    `json:"salt"`
		Signer        types.Address `json:"signer"`
		Maker         types.Address `json:"maker"`
		Taker         types.Address `json:"taker"`
		TokenID       types.U256    `json:"token_id"`
		MakerAmount   types.Decimal `json:"maker_amount"`
		TakerAmount   types.Decimal `json:"taker_amount"`
		Expiration    types.U256    `json:"expiration"`
		Side          string        `json:"side"` // BUY/SELL
		FeeRateBps    types.Decimal `json:"fee_rate_bps"`
		Nonce         types.U256    `json:"nonce"`
		SignatureType *int          `json:"signature_type,omitempty"` // 0=EOA, 1=Proxy, 2=Safe
	}

	PriceHistoryPoint struct {
		Timestamp int64  `json:"t"`
		Price     string `json:"p"`
	}

	Trade struct {
		ID        string `json:"id"`
		Price     string `json:"price"`
		Size      string `json:"size"`
		Side      string `json:"side"`
		Timestamp int64  `json:"timestamp"`
	}

	Notification struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	RewardsMarket struct {
		ID          string `json:"id"`
		ConditionID string `json:"condition_id"`
		// ...
	}

	TradeEvent struct {
		// ...
	}

	APIKeyInfo struct {
		APIKey string `json:"apiKey"`
		Type   string `json:"type"`
	}
)
