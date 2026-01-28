package clobws

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-polymarket-sdk/pkg/auth"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

const (
	ProdBaseURL = "wss://ws-subscriptions-clob.polymarket.com"
)

type clientImpl struct {
	baseURL      string
	marketURL    string
	userURL      string
	conn         *websocket.Conn
	userConn     *websocket.Conn
	signer       auth.Signer
	apiKey       *auth.APIKey
	mu           sync.Mutex
	userMu       sync.Mutex
	marketInitMu sync.Mutex
	userInitMu   sync.Mutex
	done         chan struct{}
	closeOnce    sync.Once
	closing      atomic.Bool
	// Subscription state
	debug          bool
	disablePing    bool
	reconnect      bool
	reconnectMax   int
	reconnectDelay time.Duration

	subMu          sync.Mutex
	marketAssets   map[string]struct{}
	userMarkets    map[string]struct{}
	customFeatures bool

	// Channels
	orderbookCh      chan OrderbookEvent
	priceCh          chan PriceEvent
	midpointCh       chan MidpointEvent
	lastTradeCh      chan LastTradePriceEvent
	tickSizeCh       chan TickSizeChangeEvent
	bestBidAskCh     chan BestBidAskEvent
	newMarketCh      chan NewMarketEvent
	marketResolvedCh chan MarketResolvedEvent
	tradeCh          chan TradeEvent
	orderCh          chan OrderEvent

	// Callbacks or listeners could be added here
}

func NewClient(url string, signer auth.Signer, apiKey *auth.APIKey) (Client, error) {
	marketURL, userURL, baseURL := normalizeWSURLs(url)

	reconnect := true
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT")); raw != "" {
		reconnect = raw != "0" && strings.ToLower(raw) != "false"
	}
	reconnectDelay := 2 * time.Second
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_DELAY_MS")); raw != "" {
		if ms, err := strconv.Atoi(raw); err == nil && ms > 0 {
			reconnectDelay = time.Duration(ms) * time.Millisecond
		}
	}
	reconnectMax := 5
	if raw := strings.TrimSpace(os.Getenv("CLOB_WS_RECONNECT_MAX")); raw != "" {
		if max, err := strconv.Atoi(raw); err == nil {
			reconnectMax = max
		}
	}

	c := &clientImpl{
		baseURL:          baseURL,
		marketURL:        marketURL,
		userURL:          userURL,
		signer:           signer,
		apiKey:           apiKey,
		debug:            os.Getenv("CLOB_WS_DEBUG") != "",
		disablePing:      os.Getenv("CLOB_WS_DISABLE_PING") != "",
		reconnect:        reconnect,
		reconnectDelay:   reconnectDelay,
		reconnectMax:     reconnectMax,
		done:             make(chan struct{}),
		marketAssets:     make(map[string]struct{}),
		userMarkets:      make(map[string]struct{}),
		orderbookCh:      make(chan OrderbookEvent, 100),
		priceCh:          make(chan PriceEvent, 100),
		midpointCh:       make(chan MidpointEvent, 100),
		lastTradeCh:      make(chan LastTradePriceEvent, 100),
		tickSizeCh:       make(chan TickSizeChangeEvent, 100),
		bestBidAskCh:     make(chan BestBidAskEvent, 100),
		newMarketCh:      make(chan NewMarketEvent, 100),
		marketResolvedCh: make(chan MarketResolvedEvent, 100),
		tradeCh:          make(chan TradeEvent, 100),
		orderCh:          make(chan OrderEvent, 100),
	}

	if err := c.ensureMarketConn(); err != nil {
		return nil, err
	}
	return c, nil
}

func normalizeWSURLs(raw string) (string, string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = ProdBaseURL
	}
	trimmed := strings.TrimRight(raw, "/")
	switch {
	case strings.HasSuffix(trimmed, "/ws/market"):
		base := strings.TrimSuffix(trimmed, "/ws/market")
		return trimmed, base + "/ws/user", base
	case strings.HasSuffix(trimmed, "/ws/user"):
		base := strings.TrimSuffix(trimmed, "/ws/user")
		return base + "/ws/market", trimmed, base
	default:
		return trimmed + "/ws/market", trimmed + "/ws/user", trimmed
	}
}

func (c *clientImpl) pingLoop(channel Channel) {
	ticker := time.NewTicker(10 * time.Second) // WSS quickstart uses 10s PING interval
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			// CLOB WS uses "PING" string for Keep-Alive
			err := c.writeMessage(channel, []byte("PING"))
			if err != nil {
				return
			}
		}
	}
}

func (c *clientImpl) ensureMarketConn() error {
	c.marketInitMu.Lock()
	defer c.marketInitMu.Unlock()
	if c.getConn(ChannelMarket) != nil {
		return nil
	}
	if err := c.connectMarket(); err != nil {
		return err
	}
	go c.readLoop(ChannelMarket)
	if !c.disablePing {
		go c.pingLoop(ChannelMarket)
	}
	return nil
}

func (c *clientImpl) ensureUserConn() error {
	c.userInitMu.Lock()
	defer c.userInitMu.Unlock()
	if c.getConn(ChannelUser) != nil {
		return nil
	}
	if err := c.connectUser(); err != nil {
		return err
	}
	go c.readLoop(ChannelUser)
	if !c.disablePing {
		go c.pingLoop(ChannelUser)
	}
	return nil
}

func (c *clientImpl) ensureConn(channel Channel) error {
	switch channel {
	case ChannelMarket:
		return c.ensureMarketConn()
	case ChannelUser:
		return c.ensureUserConn()
	default:
		return errors.New("unknown subscription channel")
	}
}

func (c *clientImpl) connect(url string, setConn func(*websocket.Conn)) error {
	headers := http.Header{}
	headers.Set("User-Agent", "Go-Polymarket-SDK/1.0")
	headers.Set("Origin", "https://polymarket.com") // Set Origin to bypass potential WAF/CORS checks

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		return err
	}
	setConn(conn)

	// If authenticated, send auth message or headers?
	// Polymarket WS usually requires auth for private channels (orders, trades).
	// Public channels don't need auth.
	// For now, we assume public channels work without auth.
	// If we need auth, we might need to send a specialized message after connect.

	return nil
}

func (c *clientImpl) connectMarket() error {
	return c.connect(c.marketURL, c.setMarketConn)
}

func (c *clientImpl) connectUser() error {
	return c.connect(c.userURL, c.setUserConn)
}

func (c *clientImpl) readLoop(channel Channel) {
	for {
		conn := c.getConn(channel)
		if conn == nil {
			if c.closing.Load() {
				break
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			if c.closing.Load() {
				break
			}
			if c.reconnect {
				if c.debug {
					log.Printf("read error: %v (reconnecting)", err)
				}
				if err := c.reconnectLoop(channel); err == nil {
					continue
				}
			}
			log.Printf("read error: %v", err)
			break
		}

		// Debug: Print raw message to troubleshoot "no events"
		if c.debug {
			log.Printf("Raw WS Message: %s", string(message))
		}

		// Parse generic message to determine type
		var rawObj map[string]interface{}
		var rawArr []map[string]interface{}

		// Try unmarshal as array first
		if err := json.Unmarshal(message, &rawArr); err == nil {
			for _, item := range rawArr {
				c.processEvent(item)
			}
			continue
		}

		// Try unmarshal as single object
		if err := json.Unmarshal(message, &rawObj); err == nil {
			c.processEvent(rawObj)
			continue
		}
	}
	if c.closing.Load() {
		c.shutdown()
	}
}

func (c *clientImpl) processEvent(raw map[string]interface{}) {
	eventType, _ := raw["event_type"].(string)
	if eventType == "" {
		eventType, _ = raw["type"].(string)
	}

	// Re-marshal to bytes to use existing logic or decode from map directly
	// For simplicity, let's just use the map or re-marshal for struct decoding
	// Re-marshalling is inefficient but safe for now to reuse struct definitions
	msgBytes, _ := json.Marshal(raw)

	switch eventType {
	case "book", "orderbook": // Orderbook snapshot/update
		var wire struct {
			AssetID   string           `json:"asset_id"`
			Market    string           `json:"market"`
			Bids      []OrderbookLevel `json:"bids"`
			Asks      []OrderbookLevel `json:"asks"`
			Buys      []OrderbookLevel `json:"buys"`
			Sells     []OrderbookLevel `json:"sells"`
			Hash      string           `json:"hash"`
			Timestamp string           `json:"timestamp"`
		}
		if err := json.Unmarshal(msgBytes, &wire); err == nil {
			event := OrderbookEvent{
				AssetID:   wire.AssetID,
				Market:    wire.Market,
				Bids:      wire.Bids,
				Asks:      wire.Asks,
				Hash:      wire.Hash,
				Timestamp: wire.Timestamp,
			}
			if len(event.Bids) == 0 && len(wire.Buys) > 0 {
				event.Bids = wire.Buys
			}
			if len(event.Asks) == 0 && len(wire.Sells) > 0 {
				event.Asks = wire.Sells
			}
			select {
			case c.orderbookCh <- event:
			default:
			}

			if len(event.Bids) > 0 && len(event.Asks) > 0 {
				bid, bidErr := decimal.NewFromString(event.Bids[0].Price)
				ask, askErr := decimal.NewFromString(event.Asks[0].Price)
				if bidErr == nil && askErr == nil {
					mid := bid.Add(ask).Div(decimal.NewFromInt(2))
					select {
					case c.midpointCh <- MidpointEvent{AssetID: event.AssetID, Midpoint: mid.String()}:
					default:
					}
				}
			}
		}
	case "price", "price_change":
		var event PriceEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			select {
			case c.priceCh <- event:
			default:
			}
		}
	case "midpoint":
		var event MidpointEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			select {
			case c.midpointCh <- event:
			default:
			}
		}
	case "last_trade_price":
		var event LastTradePriceEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			select {
			case c.lastTradeCh <- event:
			default:
			}
		}
	case "tick_size_change":
		var event TickSizeChangeEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			select {
			case c.tickSizeCh <- event:
			default:
			}
		}
	case "best_bid_ask":
		var event BestBidAskEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			select {
			case c.bestBidAskCh <- event:
			default:
			}
		}
	case "new_market":
		var wire struct {
			ID           string        `json:"id"`
			Question     string        `json:"question"`
			Market       string        `json:"market"`
			Slug         string        `json:"slug"`
			Description  string        `json:"description"`
			AssetIDs     []string      `json:"assets_ids"`
			AssetIDsAlt  []string      `json:"asset_ids"`
			Outcomes     []string      `json:"outcomes"`
			EventMessage *EventMessage `json:"event_message"`
			Timestamp    string        `json:"timestamp"`
		}
		if err := json.Unmarshal(msgBytes, &wire); err == nil {
			assets := wire.AssetIDs
			if len(assets) == 0 {
				assets = wire.AssetIDsAlt
			}
			event := NewMarketEvent{
				ID:           wire.ID,
				Question:     wire.Question,
				Market:       wire.Market,
				Slug:         wire.Slug,
				Description:  wire.Description,
				AssetIDs:     assets,
				Outcomes:     wire.Outcomes,
				EventMessage: wire.EventMessage,
				Timestamp:    wire.Timestamp,
			}
			select {
			case c.newMarketCh <- event:
			default:
			}
		}
	case "market_resolved":
		var wire struct {
			ID             string        `json:"id"`
			Question       string        `json:"question"`
			Market         string        `json:"market"`
			Slug           string        `json:"slug"`
			Description    string        `json:"description"`
			AssetIDs       []string      `json:"assets_ids"`
			AssetIDsAlt    []string      `json:"asset_ids"`
			Outcomes       []string      `json:"outcomes"`
			WinningAssetID string        `json:"winning_asset_id"`
			WinningOutcome string        `json:"winning_outcome"`
			EventMessage   *EventMessage `json:"event_message"`
			Timestamp      string        `json:"timestamp"`
		}
		if err := json.Unmarshal(msgBytes, &wire); err == nil {
			assets := wire.AssetIDs
			if len(assets) == 0 {
				assets = wire.AssetIDsAlt
			}
			event := MarketResolvedEvent{
				ID:             wire.ID,
				Question:       wire.Question,
				Market:         wire.Market,
				Slug:           wire.Slug,
				Description:    wire.Description,
				AssetIDs:       assets,
				Outcomes:       wire.Outcomes,
				WinningAssetID: wire.WinningAssetID,
				WinningOutcome: wire.WinningOutcome,
				EventMessage:   wire.EventMessage,
				Timestamp:      wire.Timestamp,
			}
			select {
			case c.marketResolvedCh <- event:
			default:
			}
		}
	case "trade":
		var event TradeEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			select {
			case c.tradeCh <- event:
			default:
			}
		}
	case "order":
		var event OrderEvent
		if err := json.Unmarshal(msgBytes, &event); err == nil {
			select {
			case c.orderCh <- event:
			default:
			}
		}
	}
}

func (c *clientImpl) SubscribeOrderbook(ctx context.Context, assetIDs []string) (<-chan OrderbookEvent, error) {
	if err := c.subscribeMarketAssets(assetIDs); err != nil {
		return nil, err
	}
	return c.orderbookCh, nil
}

func (c *clientImpl) SubscribePrices(ctx context.Context, assetIDs []string) (<-chan PriceEvent, error) {
	if err := c.subscribeMarketAssets(assetIDs); err != nil {
		return nil, err
	}
	return c.priceCh, nil
}

func (c *clientImpl) SubscribeMidpoints(ctx context.Context, assetIDs []string) (<-chan MidpointEvent, error) {
	if err := c.subscribeMarketAssets(assetIDs); err != nil {
		return nil, err
	}
	return c.midpointCh, nil
}

func (c *clientImpl) SubscribeLastTradePrices(ctx context.Context, assetIDs []string) (<-chan LastTradePriceEvent, error) {
	if err := c.subscribeMarketAssets(assetIDs); err != nil {
		return nil, err
	}
	return c.lastTradeCh, nil
}

func (c *clientImpl) SubscribeTickSizeChanges(ctx context.Context, assetIDs []string) (<-chan TickSizeChangeEvent, error) {
	if err := c.subscribeMarketAssets(assetIDs); err != nil {
		return nil, err
	}
	return c.tickSizeCh, nil
}

func (c *clientImpl) SubscribeBestBidAsk(ctx context.Context, assetIDs []string) (<-chan BestBidAskEvent, error) {
	req := NewMarketSubscription(assetIDs)
	if req == nil {
		return nil, errors.New("assetIDs required")
	}
	req.WithCustomFeatures(true)
	if err := c.Subscribe(ctx, req); err != nil {
		return nil, err
	}
	return c.bestBidAskCh, nil
}

func (c *clientImpl) SubscribeNewMarkets(ctx context.Context, assetIDs []string) (<-chan NewMarketEvent, error) {
	req := NewMarketSubscription(assetIDs)
	if req == nil {
		return nil, errors.New("assetIDs required")
	}
	req.WithCustomFeatures(true)
	if err := c.Subscribe(ctx, req); err != nil {
		return nil, err
	}
	return c.newMarketCh, nil
}

func (c *clientImpl) SubscribeMarketResolutions(ctx context.Context, assetIDs []string) (<-chan MarketResolvedEvent, error) {
	req := NewMarketSubscription(assetIDs)
	if req == nil {
		return nil, errors.New("assetIDs required")
	}
	req.WithCustomFeatures(true)
	if err := c.Subscribe(ctx, req); err != nil {
		return nil, err
	}
	return c.marketResolvedCh, nil
}

func (c *clientImpl) SubscribeOrders(ctx context.Context) (<-chan OrderEvent, error) {
	return nil, errors.New("markets required: use SubscribeUserOrders")
}

func (c *clientImpl) SubscribeTrades(ctx context.Context) (<-chan TradeEvent, error) {
	return nil, errors.New("markets required: use SubscribeUserTrades")
}

func (c *clientImpl) SubscribeUserOrders(ctx context.Context, markets []string) (<-chan OrderEvent, error) {
	if err := c.Subscribe(ctx, NewUserSubscription(markets)); err != nil {
		return nil, err
	}
	return c.orderCh, nil
}

func (c *clientImpl) SubscribeUserTrades(ctx context.Context, markets []string) (<-chan TradeEvent, error) {
	if err := c.Subscribe(ctx, NewUserSubscription(markets)); err != nil {
		return nil, err
	}
	return c.tradeCh, nil
}

func (c *clientImpl) Subscribe(ctx context.Context, req *SubscriptionRequest) error {
	return c.applySubscription(req, OperationSubscribe)
}

func (c *clientImpl) Unsubscribe(ctx context.Context, req *SubscriptionRequest) error {
	if req == nil {
		return errors.New("subscription request is required")
	}
	req.Operation = OperationUnsubscribe
	return c.applySubscription(req, OperationUnsubscribe)
}

func (c *clientImpl) UnsubscribeMarketAssets(ctx context.Context, assetIDs []string) error {
	if len(assetIDs) == 0 {
		return errors.New("assetIDs required")
	}
	return c.Unsubscribe(ctx, NewMarketUnsubscribe(assetIDs))
}

func (c *clientImpl) UnsubscribeUserMarkets(ctx context.Context, markets []string) error {
	if len(markets) == 0 {
		return errors.New("markets required")
	}
	return c.Unsubscribe(ctx, NewUserUnsubscribe(markets))
}

func (c *clientImpl) applySubscription(req *SubscriptionRequest, defaultOp Operation) error {
	if req == nil {
		return errors.New("subscription request is required")
	}
	if req.Type == "" {
		if len(req.AssetIDs) > 0 {
			req.Type = ChannelMarket
		} else if len(req.Markets) > 0 {
			req.Type = ChannelUser
		} else {
			return errors.New("subscription type is required")
		}
	}

	switch req.Type {
	case ChannelMarket:
		if len(req.AssetIDs) == 0 {
			return errors.New("assetIDs required")
		}
	case ChannelUser:
		if len(req.Markets) == 0 {
			return errors.New("markets required")
		}
		if req.Auth == nil {
			req.Auth = c.authPayload()
		}
		if req.Auth == nil {
			return errors.New("user subscription requires API key credentials")
		}
	default:
		return errors.New("unknown subscription channel")
	}

	if req.Operation == "" {
		req.Operation = defaultOp
	}
	c.trackSubscription(req)
	if err := c.ensureConn(req.Type); err != nil {
		return err
	}
	return c.writeJSON(req.Type, req)
}

func (c *clientImpl) Close() error {
	c.closing.Store(true)
	c.cleanupSubscriptions()
	c.closeConn(ChannelMarket)
	c.closeConn(ChannelUser)
	c.shutdown()
	return nil
}

func (c *clientImpl) writeJSON(channel Channel, v interface{}) error {
	switch channel {
	case ChannelUser:
		c.userMu.Lock()
		defer c.userMu.Unlock()
		if c.userConn == nil {
			return errors.New("connection is not established")
		}
		return c.userConn.WriteJSON(v)
	default:
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.conn == nil {
			return errors.New("connection is not established")
		}
		return c.conn.WriteJSON(v)
	}
}

func (c *clientImpl) writeMessage(channel Channel, payload []byte) error {
	switch channel {
	case ChannelUser:
		c.userMu.Lock()
		defer c.userMu.Unlock()
		if c.userConn == nil {
			return errors.New("connection is not established")
		}
		return c.userConn.WriteMessage(websocket.TextMessage, payload)
	default:
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.conn == nil {
			return errors.New("connection is not established")
		}
		return c.conn.WriteMessage(websocket.TextMessage, payload)
	}
}

func (c *clientImpl) subscribeMarketAssets(assetIDs []string) error {
	req := NewMarketSubscription(assetIDs)
	if req == nil {
		return errors.New("assetIDs required")
	}
	return c.Subscribe(context.Background(), req)
}

func (c *clientImpl) authPayload() *AuthPayload {
	if c.apiKey == nil {
		return nil
	}
	if c.apiKey.Key == "" || c.apiKey.Secret == "" || c.apiKey.Passphrase == "" {
		return nil
	}
	return &AuthPayload{
		APIKey:     c.apiKey.Key,
		Secret:     c.apiKey.Secret,
		Passphrase: c.apiKey.Passphrase,
	}
}

func (c *clientImpl) reconnectLoop(channel Channel) error {
	var lastErr error
	delay := c.reconnectDelay

	for attempt := 0; c.reconnectMax <= 0 || attempt < c.reconnectMax; attempt++ {
		if c.closing.Load() {
			return lastErr
		}
		if c.debug {
			log.Printf("ws reconnect attempt %d in %s (%s)", attempt+1, delay, channel)
		}
		time.Sleep(delay)
		c.closeConn(channel)
		var err error
		switch channel {
		case ChannelMarket:
			err = c.connectMarket()
		case ChannelUser:
			err = c.connectUser()
		default:
			err = errors.New("unknown subscription channel")
		}
		if err == nil {
			if c.debug {
				log.Printf("ws reconnect success")
			}
			c.resubscribe(channel)
			return nil
		}
		lastErr = err
		if c.debug {
			log.Printf("ws reconnect failed: %v", err)
		}
		if delay < 30*time.Second {
			delay *= 2
		}
	}
	return lastErr
}

func (c *clientImpl) resubscribe(channel Channel) {
	assets, markets, custom := c.snapshotSubscriptions()
	switch channel {
	case ChannelMarket:
		if len(assets) == 0 {
			return
		}
		req := NewMarketSubscription(assets)
		if custom {
			req.WithCustomFeatures(true)
		}
		_ = c.Subscribe(context.Background(), req)
	case ChannelUser:
		if len(markets) == 0 {
			return
		}
		_ = c.Subscribe(context.Background(), NewUserSubscription(markets))
	}
}

func (c *clientImpl) snapshotSubscriptions() ([]string, []string, bool) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	assets := make([]string, 0, len(c.marketAssets))
	for id := range c.marketAssets {
		assets = append(assets, id)
	}
	markets := make([]string, 0, len(c.userMarkets))
	for id := range c.userMarkets {
		markets = append(markets, id)
	}
	return assets, markets, c.customFeatures
}

func (c *clientImpl) trackSubscription(req *SubscriptionRequest) {
	if req == nil {
		return
	}
	c.subMu.Lock()
	defer c.subMu.Unlock()

	switch req.Type {
	case ChannelMarket:
		for _, id := range req.AssetIDs {
			if req.Operation == OperationUnsubscribe {
				delete(c.marketAssets, id)
			} else {
				c.marketAssets[id] = struct{}{}
			}
		}
	case ChannelUser:
		for _, id := range req.Markets {
			if req.Operation == OperationUnsubscribe {
				delete(c.userMarkets, id)
			} else {
				c.userMarkets[id] = struct{}{}
			}
		}
	}
	if req.CustomFeatureEnabled != nil && *req.CustomFeatureEnabled {
		c.customFeatures = true
	}
}

func (c *clientImpl) shutdown() {
	c.closeOnce.Do(func() {
		close(c.done)
		close(c.orderbookCh)
		close(c.priceCh)
		close(c.midpointCh)
		close(c.lastTradeCh)
		close(c.tickSizeCh)
		close(c.bestBidAskCh)
		close(c.newMarketCh)
		close(c.marketResolvedCh)
		close(c.tradeCh)
		close(c.orderCh)
	})
}

func (c *clientImpl) cleanupSubscriptions() {
	assets, markets, _ := c.snapshotSubscriptions()
	if len(assets) > 0 && c.getConn(ChannelMarket) != nil {
		req := NewMarketUnsubscribe(assets)
		_ = c.writeJSON(ChannelMarket, req)
	}
	if len(markets) > 0 && c.getConn(ChannelUser) != nil {
		req := NewUserUnsubscribe(markets)
		if req.Auth == nil {
			req.Auth = c.authPayload()
		}
		_ = c.writeJSON(ChannelUser, req)
	}
}

func (c *clientImpl) getConn(channel Channel) *websocket.Conn {
	switch channel {
	case ChannelUser:
		c.userMu.Lock()
		conn := c.userConn
		c.userMu.Unlock()
		return conn
	default:
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()
		return conn
	}
}

func (c *clientImpl) setMarketConn(conn *websocket.Conn) {
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
}

func (c *clientImpl) setUserConn(conn *websocket.Conn) {
	c.userMu.Lock()
	c.userConn = conn
	c.userMu.Unlock()
}

func (c *clientImpl) closeConn(channel Channel) {
	conn := c.getConn(channel)
	if conn != nil {
		_ = conn.Close()
	}
}
