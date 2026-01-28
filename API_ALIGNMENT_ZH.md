# API 对齐表

本表根据当前仓库内可见端点集合进行对齐。
主要来源：
- CLOB REST：clob-client/src/endpoints.ts 与 py-clob-client/py_clob_client/endpoints.py
- Gamma/Data/Bridge/RTDS/CTF：rs-clob-client/src/*/mod.rs
- CLOB WS 频道规则：rs-clob-client/src/clob/ws/client.rs

## CLOB REST（https://clob.polymarket.com）

认证 / API Keys
- /auth/api-key
- /auth/api-keys
- /auth/derive-api-key
- /auth/ban-status/closed-only
- /auth/readonly-api-key
- /auth/readonly-api-keys
- /auth/validate-readonly-api-key
- /auth/builder-api-key

市场 / 订单簿
- /markets
- /markets/{id}
- /simplified-markets
- /sampling-markets
- /sampling-simplified-markets
- /book
- /books

价格 / 价差
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

订单 / 成交
- /order
- /orders
- /data/order/{id}
- /data/orders
- /data/trades
- /cancel-all
- /cancel-market-orders
- /order-scoring
- /orders-scoring

余额 / 通知
- /balance-allowance
- /balance-allowance/update
- /notifications

奖励
- /rewards/user
- /rewards/user/total
- /rewards/user/percentages
- /rewards/markets/current
- /rewards/markets/{id}
- /rewards/user/markets

活动 / 心跳 / Builder
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

Base endpoint + 频道后缀：
- /ws/market
- /ws/user

## Gamma API（https://gamma-api.polymarket.com）

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

## Data API（https://data-api.polymarket.com）

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

## Bridge API（https://bridge.polymarket.com）

- POST /deposit
- GET /supported-assets
- GET /status/{address}

## RTDS WebSocket

流类型：
- Crypto prices（Binance）
- Crypto prices（Chainlink）
- Comments

## CTF（链上）

操作：
- Condition/Collection/Position ID 计算
- PrepareCondition（初始化条件）
- Split、Merge、Redeem、NegRisk Redeem
