package bridge

// Request types.
type (
	DepositRequest struct {
		Address string `json:"address"`
	}
	StatusRequest struct {
		Address string `json:"address"`
	}
)

// Response types.
type (
	DepositResponse struct {
		Address DepositAddresses `json:"address"`
		Note    string           `json:"note,omitempty"`
	}
	DepositAddresses struct {
		EVM string `json:"evm"`
		SVM string `json:"svm"`
		BTC string `json:"btc"`
	}
	SupportedAssetsResponse struct {
		SupportedAssets []SupportedAsset `json:"supported_assets"`
		Note            string           `json:"note,omitempty"`
	}
	SupportedAsset struct {
		ChainID        string `json:"chain_id"`
		ChainName      string `json:"chain_name"`
		Token          Token  `json:"token"`
		MinCheckoutUSD string `json:"min_checkout_usd"`
	}
	Token struct {
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
	}
	StatusResponse struct {
		Transactions []DepositTransaction `json:"transactions"`
	}
	DepositTransaction struct {
		FromChainID        string `json:"from_chain_id"`
		FromTokenAddress   string `json:"from_token_address"`
		FromAmountBaseUnit string `json:"from_amount_base_unit"`
		ToChainID          string `json:"to_chain_id"`
		ToTokenAddress     string `json:"to_token_address"`
		Status             string `json:"status"`
		TxHash             string `json:"tx_hash,omitempty"`
		CreatedTimeMS      *int64 `json:"created_time_ms,omitempty"`
	}
)
