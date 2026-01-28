package clob

import (
	"context"

	"go-polymarket-sdk/pkg/auth"
)

// Client defines the CLOB REST interface.
type Client interface {
	// Auth
	WithAuth(signer auth.Signer, apiKey *auth.APIKey) Client
	WithBuilderConfig(config *auth.BuilderConfig) Client
	WithUseServerTime(use bool) Client
	WithGeoblockHost(host string) Client

	// High-level Helpers
	CreateOrder(ctx context.Context, order *Order) (OrderResponse, error)

	Health(ctx context.Context) (string, error)
	Time(ctx context.Context) (TimeResponse, error)
	Geoblock(ctx context.Context) (GeoblockResponse, error)

	Markets(ctx context.Context, req *MarketsRequest) (MarketsResponse, error)
	Market(ctx context.Context, id string) (MarketResponse, error)
	SimplifiedMarkets(ctx context.Context, req *MarketsRequest) (MarketsResponse, error)
	SamplingMarkets(ctx context.Context, req *MarketsRequest) (MarketsResponse, error)
	SamplingSimplifiedMarkets(ctx context.Context, req *MarketsRequest) (MarketsResponse, error)

	OrderBook(ctx context.Context, req *BookRequest) (OrderBookResponse, error)
	OrderBooks(ctx context.Context, req *BooksRequest) (OrderBooksResponse, error)

	Midpoint(ctx context.Context, req *MidpointRequest) (MidpointResponse, error)
	Midpoints(ctx context.Context, req *MidpointsRequest) (MidpointsResponse, error)
	Price(ctx context.Context, req *PriceRequest) (PriceResponse, error)
	Prices(ctx context.Context, req *PricesRequest) (PricesResponse, error)
	AllPrices(ctx context.Context) (PricesResponse, error)
	Spread(ctx context.Context, req *SpreadRequest) (SpreadResponse, error)
	Spreads(ctx context.Context, req *SpreadsRequest) (SpreadsResponse, error)
	LastTradePrice(ctx context.Context, req *LastTradePriceRequest) (LastTradePriceResponse, error)
	LastTradesPrices(ctx context.Context, req *LastTradesPricesRequest) (LastTradesPricesResponse, error)
	TickSize(ctx context.Context, req *TickSizeRequest) (TickSizeResponse, error)
	NegRisk(ctx context.Context, req *NegRiskRequest) (NegRiskResponse, error)
	FeeRate(ctx context.Context, req *FeeRateRequest) (FeeRateResponse, error)
	PricesHistory(ctx context.Context, req *PricesHistoryRequest) (PricesHistoryResponse, error)

	// Cache helpers
	InvalidateCaches()
	SetTickSize(tokenID, tickSize string)
	SetNegRisk(tokenID string, negRisk bool)
	SetFeeRateBps(tokenID string, feeRateBps int64)

	PostOrder(ctx context.Context, req *SignedOrder) (OrderResponse, error)
	PostOrders(ctx context.Context, req *SignedOrders) (PostOrdersResponse, error)
	CancelOrder(ctx context.Context, req *CancelOrderRequest) (CancelResponse, error)
	CancelOrders(ctx context.Context, req *CancelOrdersRequest) (CancelResponse, error)
	CancelAll(ctx context.Context) (CancelAllResponse, error)
	CancelMarketOrders(ctx context.Context, req *CancelMarketOrdersRequest) (CancelMarketOrdersResponse, error)
	Order(ctx context.Context, id string) (OrderResponse, error)
	Orders(ctx context.Context, req *OrdersRequest) (OrdersResponse, error)
	Trades(ctx context.Context, req *TradesRequest) (TradesResponse, error)

	CreateOrderWithOptions(ctx context.Context, order *Order, opts *OrderOptions) (OrderResponse, error)
	CreateOrderFromSignable(ctx context.Context, order *SignableOrder) (OrderResponse, error)

	OrdersAll(ctx context.Context, req *OrdersRequest) ([]OrderResponse, error)
	TradesAll(ctx context.Context, req *TradesRequest) ([]Trade, error)
	BuilderTradesAll(ctx context.Context, req *BuilderTradesRequest) ([]Trade, error)

	OrderScoring(ctx context.Context, req *OrderScoringRequest) (OrderScoringResponse, error)
	OrdersScoring(ctx context.Context, req *OrdersScoringRequest) (OrdersScoringResponse, error)

	BalanceAllowance(ctx context.Context, req *BalanceAllowanceRequest) (BalanceAllowanceResponse, error)
	UpdateBalanceAllowance(ctx context.Context, req *BalanceAllowanceUpdateRequest) (BalanceAllowanceResponse, error)

	Notifications(ctx context.Context, req *NotificationsRequest) (NotificationsResponse, error)
	DropNotifications(ctx context.Context, req *DropNotificationsRequest) (DropNotificationsResponse, error)

	UserEarnings(ctx context.Context, req *UserEarningsRequest) (UserEarningsResponse, error)
	UserTotalEarnings(ctx context.Context, req *UserTotalEarningsRequest) (UserTotalEarningsResponse, error)
	UserRewardPercentages(ctx context.Context, req *UserRewardPercentagesRequest) (UserRewardPercentagesResponse, error)
	RewardsMarketsCurrent(ctx context.Context) (RewardsMarketsResponse, error)
	RewardsMarkets(ctx context.Context, id string) (RewardsMarketResponse, error)
	UserRewardsByMarket(ctx context.Context, req *UserRewardsByMarketRequest) (UserRewardsByMarketResponse, error)

	MarketTradesEvents(ctx context.Context, id string) (MarketTradesEventsResponse, error)
	Heartbeat(ctx context.Context, req *HeartbeatRequest) (HeartbeatResponse, error)

	CreateAPIKey(ctx context.Context) (APIKeyResponse, error)
	ListAPIKeys(ctx context.Context) (APIKeyListResponse, error)
	DeleteAPIKey(ctx context.Context, id string) (APIKeyResponse, error)
	DeriveAPIKey(ctx context.Context) (APIKeyResponse, error)
	ClosedOnlyStatus(ctx context.Context) (ClosedOnlyResponse, error)

	CreateReadonlyAPIKey(ctx context.Context) (APIKeyResponse, error)
	ListReadonlyAPIKeys(ctx context.Context) (APIKeyListResponse, error)
	DeleteReadonlyAPIKey(ctx context.Context, id string) (APIKeyResponse, error)
	ValidateReadonlyAPIKey(ctx context.Context, req *ValidateReadonlyAPIKeyRequest) (ValidateReadonlyAPIKeyResponse, error)

	CreateBuilderAPIKey(ctx context.Context) (APIKeyResponse, error)
	ListBuilderAPIKeys(ctx context.Context) (APIKeyListResponse, error)
	RevokeBuilderAPIKey(ctx context.Context, id string) (APIKeyResponse, error)
	BuilderTrades(ctx context.Context, req *BuilderTradesRequest) (BuilderTradesResponse, error)

	CreateRFQRequest(ctx context.Context, req *RFQRequest) (RFQRequestResponse, error)
	CancelRFQRequest(ctx context.Context, req *RFQCancelRequest) (RFQCancelResponse, error)
	RFQRequests(ctx context.Context, req *RFQRequestsQuery) (RFQRequestsResponse, error)
	CreateRFQQuote(ctx context.Context, req *RFQQuote) (RFQQuoteResponse, error)
	CancelRFQQuote(ctx context.Context, req *RFQCancelQuote) (RFQCancelResponse, error)
	RFQRequesterQuotes(ctx context.Context, req *RFQRequesterQuotesQuery) (RFQQuotesResponse, error)
	RFQQuoterQuotes(ctx context.Context, req *RFQQuoterQuotesQuery) (RFQQuotesResponse, error)
	RFQBestQuote(ctx context.Context, req *RFQBestQuoteQuery) (RFQBestQuoteResponse, error)
	RFQRequestAccept(ctx context.Context, req *RFQAcceptRequest) (RFQAcceptResponse, error)
	RFQQuoteApprove(ctx context.Context, req *RFQApproveQuote) (RFQApproveResponse, error)
	RFQConfig(ctx context.Context) (RFQConfigResponse, error)
}
