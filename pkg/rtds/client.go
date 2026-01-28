package rtds

import "context"

// Client defines the RTDS WebSocket interface.
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
	ConnectionStateStream(ctx context.Context) (*Stream[ConnectionStateEvent], error)
	SubscriptionCount() int
	Close() error
}
