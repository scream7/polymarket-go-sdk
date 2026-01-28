package clob

import (
	"context"
	"fmt"
)

type stubClient struct {
	*clientImpl

	tickSize      string
	feeRate       int64
	book          OrderBookResponse
	orders        map[string]OrdersResponse
	trades        map[string]TradesResponse
	builderTrades map[string]BuilderTradesResponse
}

func newStubClient() *stubClient {
	return &stubClient{
		clientImpl:    &clientImpl{},
		orders:        make(map[string]OrdersResponse),
		trades:        make(map[string]TradesResponse),
		builderTrades: make(map[string]BuilderTradesResponse),
	}
}

func (s *stubClient) OrderBook(ctx context.Context, req *BookRequest) (OrderBookResponse, error) {
	return s.book, nil
}

func (s *stubClient) TickSize(ctx context.Context, req *TickSizeRequest) (TickSizeResponse, error) {
	return TickSizeResponse{MinimumTickSize: s.tickSize}, nil
}

func (s *stubClient) FeeRate(ctx context.Context, req *FeeRateRequest) (FeeRateResponse, error) {
	return FeeRateResponse{BaseFee: int(s.feeRate)}, nil
}

func (s *stubClient) Orders(ctx context.Context, req *OrdersRequest) (OrdersResponse, error) {
	cursor := cursorFromOrdersRequest(req)
	resp, ok := s.orders[cursor]
	if !ok {
		return OrdersResponse{}, fmt.Errorf("unexpected orders cursor %q", cursor)
	}
	return resp, nil
}

func (s *stubClient) Trades(ctx context.Context, req *TradesRequest) (TradesResponse, error) {
	cursor := cursorFromTradesRequest(req)
	resp, ok := s.trades[cursor]
	if !ok {
		return TradesResponse{}, fmt.Errorf("unexpected trades cursor %q", cursor)
	}
	return resp, nil
}

func (s *stubClient) BuilderTrades(ctx context.Context, req *BuilderTradesRequest) (BuilderTradesResponse, error) {
	cursor := cursorFromBuilderTradesRequest(req)
	resp, ok := s.builderTrades[cursor]
	if !ok {
		return BuilderTradesResponse{}, fmt.Errorf("unexpected builder trades cursor %q", cursor)
	}
	return resp, nil
}

func cursorFromOrdersRequest(req *OrdersRequest) string {
	if req == nil {
		return InitialCursor
	}
	if req.NextCursor != "" {
		return req.NextCursor
	}
	if req.Cursor != "" {
		return req.Cursor
	}
	return InitialCursor
}

func cursorFromTradesRequest(req *TradesRequest) string {
	if req == nil {
		return InitialCursor
	}
	if req.NextCursor != "" {
		return req.NextCursor
	}
	if req.Cursor != "" {
		return req.Cursor
	}
	return InitialCursor
}

func cursorFromBuilderTradesRequest(req *BuilderTradesRequest) string {
	if req == nil {
		return InitialCursor
	}
	if req.NextCursor != "" {
		return req.NextCursor
	}
	if req.Cursor != "" {
		return req.Cursor
	}
	return InitialCursor
}
