package clob

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
)

func mustSigner(t *testing.T) auth.Signer {
	t.Helper()
	signer, err := auth.NewPrivateKeySigner("0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318", 137)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}
	return signer
}

func TestBuildMarketPriceValidation(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = "0.01"
	stub.feeRate = 0

	_, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(10).
		OrderType(clobtypes.OrderTypeFAK).
		Price(0.123).
		BuildMarket()
	if err == nil || !strings.Contains(err.Error(), "decimal places") {
		t.Fatalf("expected decimal place validation error, got %v", err)
	}
}

func TestBuildMarketAmountSharesValidation(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = "0.01"
	stub.feeRate = 0

	_, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("SELL").
		AmountShares(1.234).
		OrderType(clobtypes.OrderTypeFAK).
		BuildMarket()
	if err == nil || !strings.Contains(err.Error(), "amount has too many decimal places") {
		t.Fatalf("expected amount decimal validation error, got %v", err)
	}
}

func TestBuildMarketAmountUSDCValidation(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = "0.01"
	stub.feeRate = 0

	_, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(0.0000001).
		OrderType(clobtypes.OrderTypeFAK).
		BuildMarket()
	if err == nil || !strings.Contains(err.Error(), "amount has too many decimal places") {
		t.Fatalf("expected amount decimal validation error, got %v", err)
	}
}

func TestBuildMarketUsesOrderBookDepth(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = "0.01"
	stub.feeRate = 0
	stub.book = clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.6", Size: "100"},
			{Price: "0.55", Size: "100"},
			{Price: "0.5", Size: "100"},
		},
	}

	signable, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(50).
		OrderType(clobtypes.OrderTypeFAK).
		BuildMarket()
	if err != nil {
		t.Fatalf("BuildMarket failed: %v", err)
	}

	expectedMaker := decimal.NewFromInt(50_000_000)
	expectedTaker := decimal.NewFromInt(100_000_000)

	if !signable.Order.MakerAmount.Equal(expectedMaker) {
		t.Fatalf("maker amount mismatch: got %s want %s", signable.Order.MakerAmount.String(), expectedMaker.String())
	}
	if !signable.Order.TakerAmount.Equal(expectedTaker) {
		t.Fatalf("taker amount mismatch: got %s want %s", signable.Order.TakerAmount.String(), expectedTaker.String())
	}
}

func TestBuildMarketFOKInsufficientLiquidity(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = "0.01"
	stub.feeRate = 0
	stub.book = clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.6", Size: "1"},
		},
	}

	_, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(100).
		OrderType(clobtypes.OrderTypeFOK).
		BuildMarket()
	if err == nil || !strings.Contains(err.Error(), "insufficient liquidity") {
		t.Fatalf("expected insufficient liquidity error, got %v", err)
	}
}

func TestBuildMarketFAKUsesTopPriceWhenInsufficient(t *testing.T) {
	stub := newStubClient()
	stub.tickSize = "0.01"
	stub.feeRate = 0
	stub.book = clobtypes.OrderBookResponse{
		Asks: []clobtypes.PriceLevel{
			{Price: "0.6", Size: "1"},
			{Price: "0.55", Size: "1"},
		},
	}

	signable, err := NewOrderBuilder(stub, mustSigner(t)).
		TokenID("123").
		Side("BUY").
		AmountUSDC(100).
		OrderType(clobtypes.OrderTypeFAK).
		BuildMarket()
	if err != nil {
		t.Fatalf("BuildMarket failed: %v", err)
	}

	price := decimal.RequireFromString("0.6")
	tickScale := decimalPlaces(decimal.RequireFromString("0.01"))
	rawAmount := decimal.NewFromInt(100)
	takerAmount := rawAmount.Div(price).Truncate(tickScale + lotSizeScale)
	expectedTaker := toFixedDecimal(takerAmount)

	if !signable.Order.MakerAmount.Equal(decimal.NewFromInt(100_000_000)) {
		t.Fatalf("maker amount mismatch: got %s", signable.Order.MakerAmount.String())
	}
	if !signable.Order.TakerAmount.Equal(expectedTaker) {
		t.Fatalf("taker amount mismatch: got %s want %s", signable.Order.TakerAmount.String(), expectedTaker.String())
	}
}
