package coinbasepro

import "github.com/shopspring/decimal"

// Fees describes the current maker & taker fee rates, as well as the 30-day trailing volume.
// Quoted rates are subject to change.
// Note: the docs (https://docs.pro.coinbase.com/#fees) are wrong; the response is an object, not an array
type Fees struct {
	MakerFeeRate decimal.Decimal `json:"maker_fee_rate"`
	TakerFeeRate decimal.Decimal `json:"taker_fee_rate"`
	USDVolume    decimal.Decimal `json:"usd_volume"`
}
