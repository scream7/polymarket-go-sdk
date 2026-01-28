# 架构 (Go 实现版)

## 目标

- 一个 SDK 覆盖多服务：CLOB REST、CLOB WS、RTDS WS、Gamma、Data、Bridge、CTF。
- **Go 语言原生适配**：采用 `context` 控制并发，`decimal` 处理高精度数值，`go-ethereum` 处理签名。
- 直接对齐官方端点与数据结构。
- 可预期的错误处理、重试与可观测性钩子。

## 分层

1）客户端层
- 顶层 Client 聚合各服务客户端。
- 统一配置与生命周期管理。

2）传输层
- HTTP：封装 `http.Client`，统一处理 Request 构建、JSON 编解码、错误解析 (`APIError`)。
- WebSocket：重连、退避、订阅恢复与心跳（RTDS 提供 per‑sub stream，CLOB WS 覆盖市场/用户双通道）。

3）认证层
- **Signer 接口**：抽象 EIP-712 签名能力。
- **PrivateKeySigner**：基于 `go-ethereum` 实现的 EOA 签名器。
- API Key 管理：通过 `APIKey` 结构体传递 L2 凭证。

4）服务模块
- **clob**：核心交易模块。
  - **WithAuth**：链式调用注入认证信息，区分只读/交易上下文。
  - **CreateOrder**：自动构建 EIP-712 Typed Data -> 签名 -> 发送交易。
- clobws：CLOB WebSocket 市场/用户流，支持重连与订阅恢复。
- rtds：RTDS WebSocket 实时流，支持 per‑sub stream 与 lag 通知。
- gamma：市场发现 API 客户端。
- data：只读数据分析 API 客户端。
- bridge：Bridge API（入金地址/支持资产/状态）与 EVM 转账式入金/提现（跨链提现流程需额外配置）。
- ctf：链上拆分/合并/赎回/PrepareCondition 与 ID 计算。

5）共享类型 (`types` 包)
- **Decimal**：使用 `github.com/shopspring/decimal` 替代 `float64`，确保金额/价格精度。
- **U256**：封装 `math/big.Int`，处理 TokenID/Nonce，提供自定义 JSON 编解码。
- **Address**：复用 `go-ethereum/common.Address`。

## 配置模型

- 各模块可配置 Base URL。
- 支持注入自定义 `transport.Client`。
- `context.Context` 贯穿所有 IO 操作，控制超时与取消。

## 错误处理

- 统一 `types.Error`：包含 HTTP 状态码、业务错误码 (`code`)、错误消息。
- 签名错误 (`auth.ErrMissingSigner`) 与网络错误明确区分。

## 安全

- 私钥仅通过 `Signer` 接口操作，不暴露给业务层。
- 交易构建时强制校验 `Signer` 和 `APIKey` 的存在性。
- EIP-712 域分隔符严格匹配主网合约地址。
