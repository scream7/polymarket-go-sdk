package clob

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"go-polymarket-sdk/pkg/transport"
)

type staticDoer struct {
	responses map[string]string
}

func (d *staticDoer) Do(req *http.Request) (*http.Response, error) {
	key := req.URL.Path
	if req.URL.RawQuery != "" {
		key += "?" + req.URL.RawQuery
	}
	payload, ok := d.responses[key]
	if !ok {
		return nil, fmt.Errorf("unexpected request %q", key)
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(payload)),
		Header:     make(http.Header),
	}
	return resp, nil
}

func buildKey(path string, q url.Values) string {
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}

func TestOrdersAllPagination(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			buildKey("/data/orders", url.Values{"limit": {"1"}, "next_cursor": {InitialCursor}}): `{"data":[{"id":"1"}],"next_cursor":"NEXT"}`,
			buildKey("/data/orders", url.Values{"limit": {"1"}, "next_cursor": {"NEXT"}}):        `{"data":[{"id":"2"}],"next_cursor":"LTE="}`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
		cache:      newClientCache(),
	}

	results, err := client.OrdersAll(context.Background(), &OrdersRequest{Limit: 1})
	if err != nil {
		t.Fatalf("OrdersAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(results))
	}
}

func TestTradesAllPagination(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			buildKey("/data/trades", url.Values{"limit": {"1"}, "next_cursor": {InitialCursor}}): `{"data":[{"id":"1"}],"next_cursor":"NEXT"}`,
			buildKey("/data/trades", url.Values{"limit": {"1"}, "next_cursor": {"NEXT"}}):        `{"data":[{"id":"2"}],"next_cursor":"LTE="}`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
		cache:      newClientCache(),
	}

	results, err := client.TradesAll(context.Background(), &TradesRequest{Limit: 1})
	if err != nil {
		t.Fatalf("TradesAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(results))
	}
}

func TestBuilderTradesAllPagination(t *testing.T) {
	doer := &staticDoer{
		responses: map[string]string{
			buildKey("/builder/trades", url.Values{"limit": {"1"}, "next_cursor": {InitialCursor}}): `{"data":[{"id":"1"}],"next_cursor":"NEXT"}`,
			buildKey("/builder/trades", url.Values{"limit": {"1"}, "next_cursor": {"NEXT"}}):        `{"data":[{"id":"2"}],"next_cursor":"LTE="}`,
		},
	}
	client := &clientImpl{
		httpClient: transport.NewClient(doer, "http://example"),
		cache:      newClientCache(),
	}

	results, err := client.BuilderTradesAll(context.Background(), &BuilderTradesRequest{Limit: 1})
	if err != nil {
		t.Fatalf("BuilderTradesAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 builder trades, got %d", len(results))
	}
}
