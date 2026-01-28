package clobws

import "context"

// Client defines the CLOB WebSocket interface.
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
