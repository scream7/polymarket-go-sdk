# 模块说明

## clob
- CLOB REST 交易客户端。
- API key 管理、下单/撤单、余额、通知。
- RFQ 端点与心跳。

## clobws
- CLOB WebSocket 客户端。
- 市场与用户通道（订单簿、价格、订单、成交）。
- 重连与订阅管理。

## gamma
- 市场发现与元数据 API。
- 事件、市场、标签、系列、评论、公开资料、搜索。

## data
- 只读数据分析 API。
- 持仓、成交、活动、持有人、OI、交易量、排行榜。

## bridge
- 跨链入金地址与支持资产。

## rtds
- 实时数据 WebSocket（加密价格、评论）。

## ctf
- 链上条件代币框架操作。
- ID 计算、拆分/合并/赎回。

## auth（共享）
- EIP-712 签名、API key 创建/派生。
- 签名类型选择（EOA/Magic/Proxy/Safe）。
- Builder 认证支持。

## transport（共享）
- HTTP 客户端：重试、限流、指标、日志钩子。
- WS 客户端：重连、退避、订阅管理。
