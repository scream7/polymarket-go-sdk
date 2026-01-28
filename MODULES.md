# Modules

## clob
- CLOB REST trading client.
- API key management, order placement, cancellations, balances, notifications.
- RFQ endpoints and heartbeats.

## clobws
- CLOB WebSocket client.
- Market and user channels (orderbook, prices, orders, trades).
- Reconnect and subscription management.

## gamma
- Market discovery and metadata API.
- Events, markets, tags, series, comments, profiles, search.

## data
- Read-only analytics API.
- Positions, trades, activity, holders, open interest, volume, leaderboards.

## bridge
- Cross-chain deposit addresses and supported assets.

## rtds
- Real-time data socket (WebSocket) for crypto prices and comments.

## ctf
- On-chain Conditional Token Framework operations.
- ID calculation, split/merge/redeem.

## auth (shared)
- EIP-712 signing, API key create/derive.
- Signature type selection (EOA/Magic/Proxy/Safe).
- Builder authentication support.

## transport (shared)
- HTTP client with retry, rate limit, metrics, logging hooks.
- WS client with reconnect, backoff, and subscriptions.
