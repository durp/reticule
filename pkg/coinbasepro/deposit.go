package coinbasepro

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// Deposit represents the movement of Currency into accounts from both external and internal sources.
// Deposits are implemented as Transfers, but I assume this was confusing/naive, as the documentation
// rebrands Transfers as Deposits. I have followed the hint and done the same.
type Deposit struct {
	// AccountID identifies the Account to which the Deposit applies
	AccountID string `json:"account_id"`
	// Amount is the amount of the Deposit
	Amount decimal.Decimal `json:"amount"`
	// CanceledAt is the time of cancellation, if the Deposit was canceled
	CanceledAt *Time `json:"canceled_at"`
	// CompletedAt is the time of completion, if the Deposit was completed
	CompletedAt *Time `json:"completed_at"`
	// CreatedAt is the time of creation
	CreatedAt Time `json:"created_at"`
	// CreatedAt is the time the Deposit was created
	Currency CurrencyName `json:"currency"`
	// Details provides more fine-grained information describing the Deposit
	Details DepositDetails `json:"details"`
	// ProcessedAt is the time the Deposit was processed
	ProcessedAt *Time `json:"processed_at"`
	// ID uniquely identifies the Deposit
	ID string `json:"id"`
	// Type identifies the type of the Deposit (`deposit` or `internal_deposit`)
	Type DepositType `json:"type"`
	// UserID that initiated the Deposit
	UserID    string `json:"user_id"`
	UserNonce string `json:"user_nonce"`
}

type DepositType string

const (
	// DepositTypeDeposit indicates a deposit to a portfolio from an external source
	DepositTypeDeposit DepositType = "deposit"
	// DepositTypeInternal indicates a transfer between portfolios
	DepositTypeInternal DepositType = "internal_deposit"
)

func (d DepositType) Valid() error {
	switch d {
	case DepositTypeDeposit, DepositTypeInternal:
		return nil
	default:
		return fmt.Errorf("'deposit_type' %q is invalid", d)
	}
}

// TODO: DepositDetails is a kitchen sink; hard to tell if it should be anything more than a set of annotations
// or labels. Below is an example of an abandoned attempt to impose structure. For any given DepositType, a subset
// of the information below might be provided.
/*
type DepositDetails struct {
	CoinbaseAccountID         string `json:"coinbase_account_id"`
	CoinbaseDepositID         string `json:"coinbase_deposit_id"`
	CoinbasePaymentMethodID   string `json:"coinbase_payment_method_id"`
	CoinbasePaymentMethodType string `json:"coinbase_payment_method_type"`
  \\ The value I found here were incompatible with coinbasepro.Time,
	\\ `error: parsing time "2015-02-18T16:54:00-08:00" as "2006-01-02 15:04:05.999999+00": cannot parse "T16:54:00-08:00" as " "`
  \\ The values are valid time.RFC3999; added to Time
  CoinbasePayoutAt      *Time  `json:"coinbase_payout_at"`
  CoinbaseTransactionID string `json:"coinbase_transaction_id"`
  CryptoAddress         string `json:"crypto_address"`
  CryptoTransactionHash string `json:"crypto_transaction_hash"`
  CryptoTransactionID   string `json:"crypto_transaction_id"`
  DestinationTag        int64  `json:"destination_tag"`
  DestinationTagName    string `json:"destination_tag_name"`
}
*/

// DepositDetails is not well documented; until proven or requested otherwise, I will simply treat the details
// as free form annotations or labels.
type DepositDetails map[string]interface{}

// DepositFilter filters the list of deposits to be retrieved.
type DepositFilter struct {
	// ProfileID limits the list of Deposits to the ProfileID. By default, Deposits retrieves Deposits for the default profile.
	ProfileID string `json:"profile_id"`
	// Type identifies the type of the Deposit (`deposit` or `internal_deposit`)
	Type DepositType `json:"type"`
}

// Params transforms the filter into query params.
func (d *DepositFilter) Params() []string {
	var params []string
	if d.ProfileID != "" {
		params = append(params, fmt.Sprintf("profile_id=%s", d.ProfileID))
	}
	if d.Type != "" {
		params = append(params, fmt.Sprintf("type=%s", d.Type))
	}
	return params
}

// Deposits is a paged collection of Deposits
type Deposits struct {
	Deposits []*Deposit  `json:"deposits"`
	Page     *Pagination `json:"page,omitempty"`
}

func (d *Deposits) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &d.Deposits)
}

// PaymentMethodDepositSpec deposits funds from a PaymentMethod
type PaymentMethodDepositSpec struct {
	Amount          decimal.Decimal `json:"amount"`
	Currency        CurrencyName    `json:"currency"`
	PaymentMethodID string          `json:"payment_method_id"`
}

func (p *PaymentMethodDepositSpec) UnmarshalJSON(b []byte) error {
	type Alias PaymentMethodDepositSpec
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	return json.Unmarshal(b, &aux)
}

// PaymentMethodDeposit describes the payout from a PaymentMethodDepositSpec
type PaymentMethodDeposit struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency CurrencyName    `json:"currency"`
	ID       string          `json:"id"`
}

// CoinbaseAccountDeposit describes the payout from a CoinbaseAccount
type CoinbaseAccountDeposit struct {
	Amount            decimal.Decimal `json:"amount"`
	Currency          CurrencyName    `json:"currency"`
	CoinbaseAccountID string          `json:"coinbase_account_id"`
}

func (c *CoinbaseAccountDeposit) UnmarshalJSON(b []byte) error {
	type Alias CoinbaseAccountDeposit
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	return json.Unmarshal(b, &aux)
}

// PaymentMethod describes a source of currency
type PaymentMethod struct {
	AllowBuy     bool                `json:"allow_buy"`
	AllowDeposit bool                `json:"allow_deposit"`
	AllowSell    bool                `json:"allow_sell"`
	Currency     CurrencyName        `json:"currency"`
	ID           string              `json:"id"`
	Limits       PaymentMethodLimits `json:"limits"`
	Name         string              `json:"name"`
	PrimaryBuy   bool                `json:"primary_buy"`
	PrimarySell  bool                `json:"primary_sell"`
	Type         string              `json:"type"`
}

type PaymentMethodLimits struct {
	Buy        []PaymentMethodLimit `json:"buy"`
	Deposit    []PaymentMethodLimit `json:"deposit"`
	InstantBuy []PaymentMethodLimit `json:"instant_buy"`
	Sell       []PaymentMethodLimit `json:"sell"`
}

type PaymentMethodLimit struct {
	PeriodInDays int              `json:"period_in_days"`
	Total        AmountOfCurrency `json:"total"`
	Remaining    AmountOfCurrency `json:"remaining"`
}

type AmountOfCurrency struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency CurrencyName    `json:"currency"`
}

type CryptoDepositAddress struct {
	Address                string      `json:"address"`
	AddressInfo            AddressInfo `json:"address_info"`
	CreatedAt              Time        `json:"created_at"`
	DepositURI             string      `json:"deposit_uri"` // ?url.URL
	DestinationTag         string      `json:"destination_tag"`
	ExchangeDepositAddress bool        `json:"exchange_deposit_address"`
	ID                     string      `json:"id"`
	Network                string      `json:"network"`
	Resource               string      `json:"resource"`
	UpdatedAt              Time        `json:"updated_at"`
}

type AddressInfo struct {
	Address        string `json:"address"`
	DestinationTag string `json:"destination_tag"`
}
