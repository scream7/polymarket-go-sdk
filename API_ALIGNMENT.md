# API Alignment

This table aligns each module to the current known endpoint sets in the repo.
Primary sources:
- CLOB REST: clob-client/src/endpoints.ts and py-clob-client/py_clob_client/endpoints.py
- Gamma/Data/Bridge/RTDS/CTF: rs-clob-client/src/*/mod.rs
- CLOB WS channel rules: rs-clob-client/src/clob/ws/client.rs

## CLOB REST (https://clob.polymarket.com)

Auth / API Keys
- /auth/api-key
- /auth/api-keys
- /auth/derive-api-key
- /auth/ban-status/closed-only
- /auth/readonly-api-key
- /auth/readonly-api-keys
- /auth/validate-readonly-api-key
- /auth/builder-api-key

Markets / Orderbooks
- /markets
- /markets/{id}
- /simplified-markets
- /sampling-markets
- /sampling-simplified-markets
- /book
- /books

Pricing / Spread
- /midpoint
- /midpoints
- /price
- /prices
- /spread
- /spreads
- /last-trade-price
- /last-trades-prices
- /tick-size
- /neg-risk
- /fee-rate
- /prices-history

Orders / Trades
- /order
- /orders
- /data/order/{id}
- /data/orders
- /data/trades
- /cancel-all
- /cancel-market-orders
- /order-scoring
- /orders-scoring

Balances / Notifications
- /balance-allowance
- /balance-allowance/update
- /notifications

Rewards
- /rewards/user
- /rewards/user/total
- /rewards/user/percentages
- /rewards/markets/current
- /rewards/markets/{id}
- /rewards/user/markets

Activity / Heartbeats / Builder
- /live-activity/events/{id}
- /v1/heartbeats
- /builder/trades

RFQ
- /rfq/request
- /rfq/data/requests
- /rfq/quote
- /rfq/data/requester/quotes
- /rfq/data/quoter/quotes
- /rfq/data/best-quote
- /rfq/request/accept
- /rfq/quote/approve
- /rfq/config

## CLOB WebSocket

Base endpoint plus channel suffix:
- /ws/market
- /ws/user

## Gamma API (https://gamma-api.polymarket.com)

- /status
- /teams
- /sports
- /sports/market-types
- /tags
- /tags/{id}
- /tags/slug/{slug}
- /tags/{id}/related-tags
- /events
- /events/{id}
- /events/slug/{slug}
- /events/{id}/tags
- /markets
- /markets/{id}
- /markets/slug/{slug}
- /markets/{id}/tags
- /series
- /series/{id}
- /comments
- /comments/{id}
- /public-profile
- /public-search

## Data API (https://data-api.polymarket.com)

- /
- /positions
- /trades
- /activity
- /holders
- /value
- /closed-positions
- /traded
- /oi
- /live-volume
- /v1/leaderboard
- /v1/builders/leaderboard
- /v1/builders/volume

## Bridge API (https://bridge.polymarket.com)

- POST /deposit
- GET /supported-assets
- GET /status/{address}

## RTDS WebSocket

Streams:
- Crypto prices (Binance)
- Crypto prices (Chainlink)
- Comments

## CTF (On-chain)

Operations:
- Condition/Collection/Position ID calculation
- Prepare condition
- Split, merge, redeem, redeem (neg risk)
