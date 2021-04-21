package coinbasepro

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDepositType_Valid(t *testing.T) {
	assert.NoError(t, DepositTypeDeposit.Valid())
	assert.NoError(t, DepositTypeInternal.Valid())
	assert.Error(t, DepositType("invalid").Valid())
}

func TestDepositFilter_Params(t *testing.T) {
	filter := DepositFilter{
		ProfileID: "profile_id",
		Type:      DepositTypeDeposit,
	}
	assert.ElementsMatch(t, []string{"type=deposit", "profile_id=profile_id"}, filter.Params())
}

func TestDeposits_UnmarshalJSON(t *testing.T) {
	var timestamp Time
	require.NoError(t, timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z")))
	raw := `[
  {
    "account_id": "account_id",
		"amount": "1.0",
		"canceled_at": "2021-04-09T19:04:58.964459Z",
		"completed_at": "2021-04-09T19:04:58.964459Z",
		"created_at": "2021-04-09T19:04:58.964459Z",
    "currency": "USD",
    "details": {
      "coinbase_account_id": "coinbase_account_id",
      "coinbase_deposit_id": "coinbase_deposit_id",
      "coinbase_payment_method_id": "coinbase_payment_method_id",
      "coinbase_payment_method_type": "coinbase_payment_method_type",
      "coinbase_payout_at": "2021-04-09T11:04:58-08:00"
    },
    "processed_at": "2021-04-09T19:04:58.964459Z",
    "id": "id",
    "type": "internal_deposit",
    "user_id": "user_id",
    "user_nonce": "user_nonce"
  }
]`
	var deposits Deposits
	require.NoError(t, json.Unmarshal([]byte(raw), &deposits))
	amt, err := decimal.NewFromString("1.0")
	require.NoError(t, err)
	assert.Equal(t, []*Deposit{{
		AccountID:   "account_id",
		Amount:      amt,
		CanceledAt:  &timestamp,
		CompletedAt: &timestamp,
		CreatedAt:   timestamp,
		Currency:    "USD",
		Details: DepositDetails{
			"coinbase_account_id":          "coinbase_account_id",
			"coinbase_deposit_id":          "coinbase_deposit_id",
			"coinbase_payment_method_id":   "coinbase_payment_method_id",
			"coinbase_payment_method_type": "coinbase_payment_method_type",
			"coinbase_payout_at":           "2021-04-09T11:04:58-08:00",
		},
		ProcessedAt: &timestamp,
		ID:          "id",
		Type:        DepositTypeInternal,
		UserID:      "user_id",
		UserNonce:   "user_nonce",
	}}, deposits.Deposits)
}

func TestPaymentMethodDeposit_UnmarshalJSON(t *testing.T) {
	expected := PaymentMethodDepositSpec{
		Amount:          decimal.NewFromFloat(1.01),
		Currency:        "USD",
		PaymentMethodID: "payment_method_id",
	}
	var paymentMethodDeposit PaymentMethodDepositSpec
	err := json.Unmarshal([]byte(`{"amount":"1.01","currency":"USD","payment_method_id":"payment_method_id"}`),
		&paymentMethodDeposit)
	require.NoError(t, err)
	assert.Equal(t, expected, paymentMethodDeposit)
}

func TestCoinbaseAccountDeposit_UnmarshalJSON(t *testing.T) {
	expected := CoinbaseAccountDeposit{
		Amount:            decimal.NewFromFloat(1.01),
		Currency:          "USD",
		CoinbaseAccountID: "coinbase_account_id",
	}
	var coinbaseAccountDeposit CoinbaseAccountDeposit
	err := json.Unmarshal([]byte(`{"amount":"1.01","currency":"USD","coinbase_account_id":"coinbase_account_id"}`),
		&coinbaseAccountDeposit)
	require.NoError(t, err)
	assert.Equal(t, expected, coinbaseAccountDeposit)
}
