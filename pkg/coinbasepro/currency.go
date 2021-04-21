package coinbasepro

import "github.com/shopspring/decimal"

type CurrencyName string

type Currency struct {
	ConvertibleTo []CurrencyName  `json:"convertible_to"`
	Details       CurrencyDetails `json:"details"`
	ID            string          `json:"id"`
	MaxPrecision  decimal.Decimal `json:"max_precision"`
	Message       string          `json:"message"`
	MinSize       decimal.Decimal `json:"min_size"`
	Name          string          `json:"name"`
	Status        string          `json:"status"`
}

type CurrencyDetails map[string]interface{}
