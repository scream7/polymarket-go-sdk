package rfq

import (
	"math/big"
	"testing"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/clob/clobtypes"
	"github.com/GoPolymarket/polymarket-go-sdk/pkg/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func TestRFQRequestItemToDetail(t *testing.T) {
	item := RFQRequestItem{
		RequestID:    "req-1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		Side:         "BUY",
		SizeIn:       "10",
		SizeOut:      "5",
		Price:        "0.5",
		Expiry:       123456,
	}

	detail, err := item.ToDetail()
	if err != nil {
		t.Fatalf("ToDetail failed: %v", err)
	}
	if detail.RequestID != "req-1" {
		t.Fatalf("requestID mismatch: %s", detail.RequestID)
	}
	if detail.TokenID == nil || detail.TokenID.String() != "123" {
		t.Fatalf("tokenID mismatch: %v", detail.TokenID)
	}
	if detail.Price.String() != "0.5" {
		t.Fatalf("price mismatch: %s", detail.Price.String())
	}
}

func TestRFQQuoteItemToDetail(t *testing.T) {
	item := RFQQuoteItem{
		QuoteID:      "quote-1",
		RequestID:    "req-1",
		UserAddress:  "0x0000000000000000000000000000000000000001",
		ProxyAddress: "0x0000000000000000000000000000000000000002",
		Token:        "123",
		Complement:   "456",
		Side:         "SELL",
		SizeIn:       "10",
		SizeOut:      "5",
		Price:        "0.5",
	}

	detail, err := item.ToDetail()
	if err != nil {
		t.Fatalf("ToDetail failed: %v", err)
	}
	if detail.QuoteID != "quote-1" {
		t.Fatalf("quoteID mismatch: %s", detail.QuoteID)
	}
	if detail.TokenID == nil || detail.TokenID.String() != "123" {
		t.Fatalf("tokenID mismatch: %v", detail.TokenID)
	}
}

func TestBuildRFQAcceptRequestFromSignedOrder(t *testing.T) {
	signed := clobtypes.SignedOrder{
		Order: clobtypes.Order{
			Salt:        types.U256{Int: big.NewInt(1)},
			Maker:       common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Signer:      common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Taker:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
			TokenID:     types.U256{Int: big.NewInt(123)},
			MakerAmount: decimal.NewFromInt(100),
			TakerAmount: decimal.NewFromInt(50),
			Side:        "BUY",
			Expiration:  types.U256{Int: big.NewInt(0)},
			FeeRateBps:  decimal.NewFromInt(0),
			Nonce:       types.U256{Int: big.NewInt(10)},
		},
		Signature: "0xsig",
		Owner:     "owner",
	}

	req, err := BuildRFQAcceptRequestFromSignedOrder("req-1", "quote-1", &signed)
	if err != nil {
		t.Fatalf("BuildRFQAcceptRequestFromSignedOrder failed: %v", err)
	}
	if req.RequestID != "req-1" || req.QuoteIDV2 != "quote-1" {
		t.Fatalf("request/quote IDs mismatch")
	}
	if req.TokenID != "123" || req.Nonce != "10" {
		t.Fatalf("order fields mismatch: token=%s nonce=%s", req.TokenID, req.Nonce)
	}
}
