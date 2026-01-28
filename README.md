# Go Polymarket SDK (Full-Stack)

Unified Go SDK covering Polymarket CLOB REST, CLOB WebSocket, RTDS WebSocket,
Gamma API, Data API, Bridge API, and CTF on-chain operations.

## Scope

- Unified client with shared config, transport, and auth layers.
- Direct, zero-proxy alignment to Polymarket public endpoints.
- Consistent request/response types across modules.

## Quickstart

```go
client := polymarket.NewClient(polymarket.WithUseServerTime(true))
authClient := client.CLOB.WithAuth(signer, apiKey)

// Limit order using OrderBuilder (GTC)
signable, _ := clob.NewOrderBuilder(authClient, signer).
  TokenID("1234567890").
  Side("BUY").
  Price(0.45).
  Size(10).
  BuildSignable()

resp, _ := authClient.CreateOrderFromSignable(ctx, signable)
```

## Market Orders (FAK/FOK)

```go
signable, _ := clob.NewOrderBuilder(authClient, signer).
  TokenID("1234567890").
  Side("BUY").
  AmountUSDC(100).
  OrderType(clob.OrderTypeFAK).
  BuildMarket()
```

## Pagination Helpers

```go
trades, _ := authClient.TradesAll(ctx, &clob.TradesRequest{Limit: 50})
orders, _ := authClient.OrdersAll(ctx, &clob.OrdersRequest{Limit: 50})
```

## Contents

- ARCHITECTURE.md: Layered architecture and cross-cutting concerns.
- API_ALIGNMENT.md: Endpoint alignment table across services.
- INTERFACES.md: Go-style method naming and interface sketch.
- MODULES.md: Module responsibilities and feature breakdown.

## Examples

- `examples/order_builder`: limit order builder (offline tick size override)
- `examples/market_order`: market order builder (FAK/FOK)
- `examples/gtd_order`: GTD + expiration
- `examples/pagination`: next_cursor pagination helper
- `examples/builder_flow`: end-to-end token lookup + builder API key flow
- `examples/data_client`: Data API queries
- `examples/ctf_client`: CTF on-chain transactions
- `examples/rtds_client`: RTDS price stream
- `examples/ws_user_client`: CLOB WS user channels (orders/trades)
- `examples/ws_client`: CLOB WS market channels
- `examples/gamma_client`: Gamma API queries
- `examples/rfq_flow`: RFQ create/quote/accept/approve flow

### builder_flow env

- `POLYMARKET_PK`
- `POLYMARKET_API_KEY` / `POLYMARKET_API_SECRET` / `POLYMARKET_API_PASSPHRASE` (optional, will derive if missing)
- `POLYMARKET_BUILDER_API_KEY` / `POLYMARKET_BUILDER_API_SECRET` / `POLYMARKET_BUILDER_API_PASSPHRASE` (optional)
- `POLYMARKET_BUILDER_REMOTE_HOST` / `POLYMARKET_BUILDER_REMOTE_TOKEN` (optional, overrides local builder keys)

### ws_user_client env

- `POLYMARKET_PK`
- `POLYMARKET_API_KEY` / `POLYMARKET_API_SECRET` / `POLYMARKET_API_PASSPHRASE` (optional, will derive if missing)
- `CLOB_WS_DEBUG` (optional, log raw WS messages)
- `CLOB_WS_DISABLE_PING` (optional, disable PING keepalive)
- `CLOB_WS_RECONNECT` (optional, set 0/false to disable reconnects)
- `CLOB_WS_RECONNECT_DELAY_MS` (optional, default 2000)
- `CLOB_WS_RECONNECT_MAX` (optional, default 5, 0 = unlimited)

### rtds_client env

- `RTDS_WS_RECONNECT` (optional, set 0/false to disable reconnects)
- `RTDS_WS_RECONNECT_DELAY_MS` (optional, default 2000)
- `RTDS_WS_RECONNECT_MAX` (optional, default 5, 0 = unlimited)

### rfq_flow env

- `POLYMARKET_PK`
- `POLYMARKET_API_KEY` / `POLYMARKET_API_SECRET` / `POLYMARKET_API_PASSPHRASE` (optional, will derive if missing)
- `RFQ_ASSET_IN` / `RFQ_ASSET_OUT` / `RFQ_AMOUNT_IN` / `RFQ_AMOUNT_OUT` / `RFQ_USER_TYPE`
- `RFQ_SIGN_TOKEN_ID` / `RFQ_SIGN_SIDE` / `RFQ_SIGN_PRICE` / `RFQ_SIGN_SIZE` (optional, build a signed order)
- `RFQ_SIGN_ORDER_TYPE` / `RFQ_SIGN_POST_ONLY` (optional, signed order extras)
- `RFQ_ACCEPT_*` / `RFQ_APPROVE_*` (optional, send accept/approve using a signed order)

### data_client env

- `DATA_USER_ADDRESS` (required for user-specific queries)

### ctf_client env

- `CTF_RPC_URL`
- `CTF_PRIVATE_KEY`
- `CTF_CHAIN_ID` (optional, default 137)
- `CTF_NEG_RISK` (optional, set 1 to use NegRisk adapter)
- `CTF_DO_TX` (optional, set 1 to send a transaction)
- `CTF_ACTION` (required when CTF_DO_TX=1, split|merge|redeem|redeem_neg_risk)
- `CTF_COLLATERAL` / `CTF_CONDITION_ID`
- `CTF_PARENT_COLLECTION_ID` (optional, default 0x0)
- `CTF_PARTITION` (split/merge, comma-separated integers)
- `CTF_INDEX_SETS` (redeem, comma-separated integers)
- `CTF_AMOUNTS` (redeem_neg_risk, comma-separated integers)
- `CTF_AMOUNT` (split/merge, integer amount in base units)
