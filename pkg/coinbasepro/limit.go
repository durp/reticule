package coinbasepro

import "github.com/shopspring/decimal"

// Limits provide payment method transfer limits, as well as buy/sell limits per currency.
type Limits struct {
	LimitCurrency  CurrencyName             `json:"limit_currency"`
	TransferLimits map[string]CurrencyLimit `json:"transfer_limits"`
}

type CurrencyLimit map[CurrencyName]Limit

// TODO: haven't ever seen PeriodInDays
type Limit struct {
	Max          decimal.Decimal `json:"max"`
	Remaining    decimal.Decimal `json:"remaining"`
	PeriodInDays int             `json:"period_in_days"`
}
