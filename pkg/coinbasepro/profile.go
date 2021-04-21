package coinbasepro

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Profile struct {
	Active    bool   `json:"active"`
	CreatedAt Time   `json:"created_at"`
	ID        string `json:"id"`
	IsDefault bool   `json:"is_default"`
	Name      string `json:"name"`
	UserID    string `json:"user_id"`
}

type ProfileFilter struct {
	Active bool `json:"active"`
}

func (p ProfileFilter) Params() []string {
	var params []string
	if p.Active {
		params = append(params, "active")
	}
	return params
}

type ProfileTransferSpec struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency CurrencyName    `json:"currency"`
	From     string          `json:"from"`
	To       string          `json:"to"`
}

func (p *ProfileTransferSpec) UnmarshalJSON(b []byte) error {
	type Alias ProfileTransferSpec
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	return json.Unmarshal(b, &aux)
}

type ProfileTransfer struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency CurrencyName    `json:"currency"`
	From     string          `json:"from"`
	To       string          `json:"to"`
}
