# Go Polymarket SDK（全栈）

统一的 Go SDK，覆盖 Polymarket CLOB REST、CLOB WebSocket、RTDS WebSocket、
Gamma API、Data API、Bridge API 以及 CTF 链上操作。

## 范围

- 统一客户端与共享配置、传输层、认证层。
- 直接对齐官方公开端点，不经过代理层。
- 跨模块一致的请求/响应类型与错误模型。

## 快速开始

```go
client := polymarket.NewClient(polymarket.WithUseServerTime(true))
authClient := client.CLOB.WithAuth(signer, apiKey)

// 限价单（GTC）
signable, _ := clob.NewOrderBuilder(authClient, signer).
  TokenID("1234567890").
  Side("BUY").
  Price(0.45).
  Size(10).
  BuildSignable()

resp, _ := authClient.CreateOrderFromSignable(ctx, signable)
```

## 市价单（FAK/FOK）

```go
signable, _ := clob.NewOrderBuilder(authClient, signer).
  TokenID("1234567890").
  Side("BUY").
  AmountUSDC(100).
  OrderType(clob.OrderTypeFAK).
  BuildMarket()
```

## 分页辅助

```go
trades, _ := authClient.TradesAll(ctx, &clob.TradesRequest{Limit: 50})
orders, _ := authClient.OrdersAll(ctx, &clob.OrdersRequest{Limit: 50})
```

## 内容

- ARCHITECTURE_ZH.md：分层架构与横切关注点。
- API_ALIGNMENT_ZH.md：各服务的端点对齐表。
- INTERFACES_ZH.md：Go 风格方法命名与接口草图。
- MODULES_ZH.md：模块职责与功能拆分。

## 示例

- `examples/order_builder`：限价单构建（离线 tick size 覆盖）
- `examples/market_order`：市价单构建（FAK/FOK）
- `examples/gtd_order`：GTD + 过期时间
- `examples/pagination`：next_cursor 分页示例
- `examples/builder_flow`：端到端 Token 获取 + Builder API Key 流程
- `examples/data_client`：Data API 查询
- `examples/ctf_client`：CTF 链上交易
- `examples/rtds_client`：RTDS 行情流
- `examples/ws_user_client`：CLOB WS 用户频道（订单/成交）
- `examples/ws_client`：CLOB WS 行情频道
- `examples/gamma_client`：Gamma API 查询
- `examples/rfq_flow`：RFQ 创建/报价/接受/批准流程

### builder_flow 环境变量

- `POLYMARKET_PK`
- `POLYMARKET_API_KEY` / `POLYMARKET_API_SECRET` / `POLYMARKET_API_PASSPHRASE`（可选，缺失时自动派生）
- `POLYMARKET_BUILDER_API_KEY` / `POLYMARKET_BUILDER_API_SECRET` / `POLYMARKET_BUILDER_API_PASSPHRASE`（可选）
- `POLYMARKET_BUILDER_REMOTE_HOST` / `POLYMARKET_BUILDER_REMOTE_TOKEN`（可选，优先于本地 Builder Key）

### ws_user_client 环境变量

- `POLYMARKET_PK`
- `POLYMARKET_API_KEY` / `POLYMARKET_API_SECRET` / `POLYMARKET_API_PASSPHRASE`（可选，缺失时自动派生）
- `CLOB_WS_DEBUG`（可选，打印 WS 原始消息）
- `CLOB_WS_DISABLE_PING`（可选，禁用 PING 保活）
- `CLOB_WS_RECONNECT`（可选，设置 0/false 关闭重连）
- `CLOB_WS_RECONNECT_DELAY_MS`（可选，默认 2000）
- `CLOB_WS_RECONNECT_MAX`（可选，默认 5，0 = 无限重试）

### rtds_client 环境变量

- `RTDS_WS_RECONNECT`（可选，设置 0/false 关闭重连）
- `RTDS_WS_RECONNECT_DELAY_MS`（可选，默认 2000）
- `RTDS_WS_RECONNECT_MAX`（可选，默认 5，0 = 无限重试）

### rfq_flow 环境变量

- `POLYMARKET_PK`
- `POLYMARKET_API_KEY` / `POLYMARKET_API_SECRET` / `POLYMARKET_API_PASSPHRASE`（可选，缺失时自动派生）
- `RFQ_ASSET_IN` / `RFQ_ASSET_OUT` / `RFQ_AMOUNT_IN` / `RFQ_AMOUNT_OUT` / `RFQ_USER_TYPE`
- `RFQ_SIGN_TOKEN_ID` / `RFQ_SIGN_SIDE` / `RFQ_SIGN_PRICE` / `RFQ_SIGN_SIZE`（可选，生成签名订单）
- `RFQ_SIGN_ORDER_TYPE` / `RFQ_SIGN_POST_ONLY`（可选，签名订单参数）
- `RFQ_ACCEPT_*` / `RFQ_APPROVE_*`（可选，用签名订单发送 accept/approve）

### data_client 环境变量

- `DATA_USER_ADDRESS`（用于用户查询，必填）

### ctf_client 环境变量

- `CTF_RPC_URL`
- `CTF_PRIVATE_KEY`
- `CTF_CHAIN_ID`（可选，默认 137）
- `CTF_NEG_RISK`（可选，设置 1 使用 NegRisk 适配器）
- `CTF_DO_TX`（可选，设置 1 才会发送交易）
- `CTF_ACTION`（当 CTF_DO_TX=1 时必填，split|merge|redeem|redeem_neg_risk）
- `CTF_COLLATERAL` / `CTF_CONDITION_ID`
- `CTF_PARENT_COLLECTION_ID`（可选，默认 0x0）
- `CTF_PARTITION`（split/merge，用逗号分隔整数）
- `CTF_INDEX_SETS`（redeem，用逗号分隔整数）
- `CTF_AMOUNTS`（redeem_neg_risk，用逗号分隔整数）
- `CTF_AMOUNT`（split/merge，整数，最小单位）
