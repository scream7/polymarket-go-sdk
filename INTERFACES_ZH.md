# 接口草图

这是命名与结构的高层草图，不是生产代码。

## 顶层 Client

```go
package polymarket

type Client struct {
  CLOB   *clob.Client
  CLOBWS *clobws.Client
  Gamma  *gamma.Client
  Data   *data.Client
  Bridge *bridge.Client
  RTDS   *rtds.Client
  CTF    *ctf.Client
}
```

## CLOB REST

```go
package clob

type Client interface {
  WithAuth(signer auth.Signer, apiKey *auth.APIKey) Client
  WithBuilderConfig(config *auth.BuilderConfig) Client
  WithUseServerTime(use bool) Client

  Health(ctx context.Context) (string, error)
  Time(ctx context.Context) (TimeResponse, error)

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
  Spread(ctx context.Context, req *SpreadRequest) (SpreadResponse, error)
  Spreads(ctx context.Context, req *SpreadsRequest) (SpreadsResponse, error)
  LastTradePrice(ctx context.Context, req *LastTradePriceRequest) (LastTradePriceResponse, error)
  LastTradesPrices(ctx context.Context, req *LastTradesPricesRequest) (LastTradesPricesResponse, error)
  TickSize(ctx context.Context, req *TickSizeRequest) (TickSizeResponse, error)
  NegRisk(ctx context.Context, req *NegRiskRequest) (NegRiskResponse, error)
  FeeRate(ctx context.Context, req *FeeRateRequest) (FeeRateResponse, error)
  PricesHistory(ctx context.Context, req *PricesHistoryRequest) (PricesHistoryResponse, error)

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
```

## CLOB WS

```go
package clobws

type Client interface {
  Subscribe(ctx context.Context, req *SubscriptionRequest) error
  Unsubscribe(ctx context.Context, req *SubscriptionRequest) error
  SubscribeOrderbook(ctx context.Context, assetIDs []string) (<-chan OrderbookEvent, error)
  SubscribePrices(ctx context.Context, assetIDs []string) (<-chan PriceEvent, error)
  SubscribeMidpoints(ctx context.Context, assetIDs []string) (<-chan MidpointEvent, error)
  SubscribeLastTradePrices(ctx context.Context, assetIDs []string) (<-chan LastTradePriceEvent, error)
  SubscribeTickSizeChanges(ctx context.Context, assetIDs []string) (<-chan TickSizeChangeEvent, error)
  SubscribeBestBidAsk(ctx context.Context, assetIDs []string) (<-chan BestBidAskEvent, error)
  SubscribeNewMarkets(ctx context.Context, assetIDs []string) (<-chan NewMarketEvent, error)
  SubscribeMarketResolutions(ctx context.Context, assetIDs []string) (<-chan MarketResolvedEvent, error)
  SubscribeOrders(ctx context.Context) (<-chan OrderEvent, error)
  SubscribeTrades(ctx context.Context) (<-chan TradeEvent, error)
  SubscribeUserOrders(ctx context.Context, markets []string) (<-chan OrderEvent, error)
  SubscribeUserTrades(ctx context.Context, markets []string) (<-chan TradeEvent, error)
  UnsubscribeMarketAssets(ctx context.Context, assetIDs []string) error
  UnsubscribeUserMarkets(ctx context.Context, markets []string) error
  Close() error
}
```

## Gamma API

```go
package gamma

type Client interface {
  Status(ctx context.Context) (StatusResponse, error)
  Teams(ctx context.Context, req *TeamsRequest) ([]Team, error)
  Sports(ctx context.Context) ([]SportsMetadata, error)
  SportsMarketTypes(ctx context.Context) (SportsMarketTypesResponse, error)
  Tags(ctx context.Context, req *TagsRequest) ([]Tag, error)
  TagByID(ctx context.Context, req *TagByIDRequest) (*Tag, error)
  TagBySlug(ctx context.Context, req *TagBySlugRequest) (*Tag, error)
  RelatedTagsByID(ctx context.Context, req *RelatedTagsByIDRequest) ([]RelatedTag, error)
  RelatedTagsBySlug(ctx context.Context, req *RelatedTagsBySlugRequest) ([]RelatedTag, error)
  TagsRelatedToTagByID(ctx context.Context, req *RelatedTagsByIDRequest) ([]Tag, error)
  TagsRelatedToTagBySlug(ctx context.Context, req *RelatedTagsBySlugRequest) ([]Tag, error)
  Events(ctx context.Context, req *EventsRequest) ([]Event, error)
  EventByID(ctx context.Context, req *EventByIDRequest) (*Event, error)
  EventBySlug(ctx context.Context, req *EventBySlugRequest) (*Event, error)
  EventTags(ctx context.Context, req *EventTagsRequest) ([]Tag, error)
  Markets(ctx context.Context, req *MarketsRequest) ([]Market, error)
  MarketByID(ctx context.Context, req *MarketByIDRequest) (*Market, error)
  MarketBySlug(ctx context.Context, req *MarketBySlugRequest) (*Market, error)
  MarketTags(ctx context.Context, req *MarketTagsRequest) ([]Tag, error)
  Series(ctx context.Context, req *SeriesRequest) ([]Series, error)
  SeriesByID(ctx context.Context, req *SeriesByIDRequest) (*Series, error)
  Comments(ctx context.Context, req *CommentsRequest) ([]Comment, error)
  CommentByID(ctx context.Context, req *CommentByIDRequest) ([]Comment, error)
  CommentsByUserAddress(ctx context.Context, req *CommentsByUserAddressRequest) ([]Comment, error)
  PublicProfile(ctx context.Context, req *PublicProfileRequest) (*PublicProfile, error)
  PublicSearch(ctx context.Context, req *PublicSearchRequest) (SearchResults, error)

  GetMarkets(ctx context.Context, req *MarketsRequest) ([]Market, error)
  GetMarket(ctx context.Context, id string) (*Market, error)
  GetEvents(ctx context.Context, req *MarketsRequest) ([]Event, error)
  GetEvent(ctx context.Context, id string) (*Event, error)
}
```

## Data API

```go
package data

type Client interface {
  Health(ctx context.Context) (string, error)
  Positions(ctx context.Context, req *PositionsRequest) (PositionsResponse, error)
  Trades(ctx context.Context, req *TradesRequest) (TradesResponse, error)
  Activity(ctx context.Context, req *ActivityRequest) (ActivityResponse, error)
  Holders(ctx context.Context, req *HoldersRequest) (HoldersResponse, error)
  Value(ctx context.Context, req *ValueRequest) (ValueResponse, error)
  ClosedPositions(ctx context.Context, req *ClosedPositionsRequest) (ClosedPositionsResponse, error)
  Traded(ctx context.Context, req *TradedRequest) (TradedResponse, error)
  OpenInterest(ctx context.Context, req *OpenInterestRequest) (OpenInterestResponse, error)
  LiveVolume(ctx context.Context, req *LiveVolumeRequest) (LiveVolumeResponse, error)
  Leaderboard(ctx context.Context, req *LeaderboardRequest) (LeaderboardResponse, error)
  BuildersLeaderboard(ctx context.Context, req *BuildersLeaderboardRequest) (BuildersLeaderboardResponse, error)
  BuildersVolume(ctx context.Context, req *BuildersVolumeRequest) (BuildersVolumeResponse, error)
}
```

## Bridge API

```go
package bridge

type Client interface {
  Deposit(ctx context.Context, amount *big.Int, asset common.Address) (*types.Transaction, error)
  Withdraw(ctx context.Context, amount *big.Int, asset common.Address) (*types.Transaction, error)
  WithdrawTo(ctx context.Context, req *WithdrawRequest) (*types.Transaction, error)
  SupportedAssets(ctx context.Context) ([]common.Address, error)

  DepositAddress(ctx context.Context, req *DepositRequest) (DepositResponse, error)
  SupportedAssetsInfo(ctx context.Context) (SupportedAssetsResponse, error)
  Status(ctx context.Context, req *StatusRequest) (StatusResponse, error)
}
```

## RTDS WebSocket

```go
package rtds

type Stream[T any] struct {
  C <-chan T
  Err <-chan error
  Close() error
}

type Client interface {
  SubscribeCryptoPricesStream(ctx context.Context, symbols []string) (*Stream[CryptoPriceEvent], error)
  SubscribeChainlinkPricesStream(ctx context.Context, feeds []string) (*Stream[ChainlinkPriceEvent], error)
  SubscribeCommentsStream(ctx context.Context, req *CommentFilter) (*Stream[CommentEvent], error)
  SubscribeRawStream(ctx context.Context, sub *Subscription) (*Stream[RtdsMessage], error)
  SubscribeCryptoPrices(ctx context.Context, symbols []string) (<-chan CryptoPriceEvent, error)
  SubscribeChainlinkPrices(ctx context.Context, feeds []string) (<-chan ChainlinkPriceEvent, error)
  SubscribeComments(ctx context.Context, req *CommentFilter) (<-chan CommentEvent, error)
  SubscribeRaw(ctx context.Context, sub *Subscription) (<-chan RtdsMessage, error)
  UnsubscribeCryptoPrices(ctx context.Context) error
  UnsubscribeChainlinkPrices(ctx context.Context) error
  UnsubscribeComments(ctx context.Context, commentType *CommentType) error
  UnsubscribeRaw(ctx context.Context, sub *Subscription) error
  ConnectionState() ConnectionState
  SubscriptionCount() int
  Close() error
}
```

## CTF

```go
package ctf

type Client interface {
  PrepareCondition(ctx context.Context, req *PrepareConditionRequest) (PrepareConditionResponse, error)
  ConditionID(ctx context.Context, req *ConditionIDRequest) (ConditionIDResponse, error)
  CollectionID(ctx context.Context, req *CollectionIDRequest) (CollectionIDResponse, error)
  PositionID(ctx context.Context, req *PositionIDRequest) (PositionIDResponse, error)
  SplitPosition(ctx context.Context, req *SplitPositionRequest) (SplitPositionResponse, error)
  MergePositions(ctx context.Context, req *MergePositionsRequest) (MergePositionsResponse, error)
  RedeemPositions(ctx context.Context, req *RedeemPositionsRequest) (RedeemPositionsResponse, error)
  RedeemNegRisk(ctx context.Context, req *RedeemNegRiskRequest) (RedeemNegRiskResponse, error)
}
```
