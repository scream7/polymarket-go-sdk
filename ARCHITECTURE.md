# Architecture

## Goals

- One SDK, multiple services: CLOB REST, CLOB WS, RTDS WS, Gamma, Data, Bridge, CTF.
- Direct alignment with official endpoints and payloads.
- Predictable error handling, retries, and observability hooks.

## Layers

1) Client Layer
- Top-level Client aggregates service clients.
- Provides consistent configuration and lifecycle.

2) Transport Layer
- HTTP: retry, timeout, rate limiting, middleware.
- WebSocket: reconnect, backoff, subscription management.

3) Auth Layer
- EIP-712 signing and API key creation/derivation.
- Signature types for EOA/Magic/Proxy/Safe.
- Builder authentication support.

4) Service Modules
- clob: trading REST + RFQ + heartbeats
- clobws: CLOB WebSocket market/user streams
- rtds: RTDS WebSocket streams
- gamma: market discovery
- data: read-only analytics
- bridge: deposits and supported assets
- ctf: on-chain split/merge/redeem

5) Shared Types
- Address/U256/Decimal primitives
- Pagination, filters, sorting
- Error model and response envelopes

## Configuration Model

- BaseURLs per module with sane defaults.
- Custom HTTP client and WebSocket dialer.
- Per-request timeout override.
- Optional logging and metrics hooks.

## Error Handling

- Unified APIError with HTTP method, URL, status, code, and message.
- Transport errors preserve raw body for debugging.
- WS errors include subscription context.

## Security

- Secrets handled via interface, never logged.
- Optional builder signing uses remote signer endpoint.
- Signature type must be explicit to avoid mis-signed orders.
