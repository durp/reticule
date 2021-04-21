package coinbasepro

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type StablecoinConversionSpec struct {
	From   CurrencyName    `json:"from"`
	To     CurrencyName    `json:"to"`
	Amount decimal.Decimal `json:"amount"`
}

func (s *StablecoinConversionSpec) UnmarshalJSON(b []byte) error {
	type Alias StablecoinConversionSpec
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal(b, &aux)
}

type StablecoinConversion struct {
	Amount        decimal.Decimal `json:"amount"`
	From          CurrencyName    `json:"from"`
	FromAccountID string          `json:"from_account_id"`
	ID            string          `json:"id"`
	To            CurrencyName    `json:"to"`
	ToAccountID   string          `json:"to_account_id"`
}
