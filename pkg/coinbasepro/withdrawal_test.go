package coinbasepro

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalFilter(t *testing.T) {
	f := WithdrawalFilter{
		ProfileID: "profile_id",
		Type:      WithdrawalTypeWithdraw,
	}
	assert.ElementsMatch(t,
		[]string{"type=withdraw", "profile_id=profile_id"},
		f.Params())
}

func TestWithdrawals(t *testing.T) {
	raw := `
[
  {
    "details": {
      "something": "else"
    },
    "canceled_at": "2021-04-09T19:04:58.964459Z",
    "completed_at": "2021-04-09T19:04:58.964459Z",
    "created_at": "2021-04-09T19:04:58.964459Z",
    "processed_at": "2021-04-09T19:04:58.964459Z",
    "account_id": "account_id",
    "amount": "1.01",
    "currency": "USD",
    "id": "id",
    "type": "internal_withdraw",
    "user_id": "user_id",
    "user_nonce": "user_nonce"
  }
]
`
	var timestamp Time
	require.NoError(t, timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z")))
	var w Withdrawals
	require.NoError(t, json.Unmarshal([]byte(raw), &w))
	assert.Equal(t, []*Withdrawal{
		{
			Details:     map[string]interface{}{"something": "else"},
			CanceledAt:  &timestamp,
			CompletedAt: &timestamp,
			CreatedAt:   timestamp,
			ProcessedAt: &timestamp,
			AccountID:   "account_id",
			Amount:      decimal.NewFromFloat(1.01),
			Currency:    "USD",
			ID:          "id",
			Type:        WithdrawalTypeInternal,
			UserID:      "user_id",
			UserNonce:   "user_nonce",
		},
	}, w.Withdrawals)
}

func TestWithdrawalCancelCode(t *testing.T) {
	assert.Equal(t, "default error", WithdrawalCancelCode(0).String())
	assert.Equal(t, "transaction exceeds transaction limit", WithdrawalCancelCode(1000).String())
	assert.Equal(t, "amount must be greater than 0", WithdrawalCancelCode(1001).String())
	assert.Equal(t, "account does not support withdrawal", WithdrawalCancelCode(1002).String())
	assert.Equal(t, "payment method does not support withdrawal", WithdrawalCancelCode(1003).String())
	assert.Equal(t, "cannot withdraw this currency with this payment method", WithdrawalCancelCode(1004).String())
	assert.Equal(t, "withdrawal amount exceeds funds in account", WithdrawalCancelCode(1005).String())
	assert.Equal(t, "withdrawal amount must be at least 1.00", WithdrawalCancelCode(1006).String())
	assert.Equal(t, "withdrawal amount too small", WithdrawalCancelCode(1007).String())
	assert.Equal(t, "payment method cannot be used with recurring transfers", WithdrawalCancelCode(1008).String())
	assert.Equal(t, "withdrawal currency does not match account currency", WithdrawalCancelCode(1009).String())
	assert.Equal(t, "payment method unsupported", WithdrawalCancelCode(1010).String())
	assert.Equal(t, "withdrawal rate limit exceeded: try again in a few hours", WithdrawalCancelCode(1011).String())
	assert.Equal(t, "amount would exceed maximum account balance", WithdrawalCancelCode(1012).String())
	assert.Equal(t, "amount must be positive", WithdrawalCancelCode(1013).String())
	assert.Equal(t, "warning: with no tag name, recipient may lose funds: confirm that recipient does not require tag name", WithdrawalCancelCode(2000).String())
	assert.Equal(t, "amount would exceed send limits: try a smaller amount or try again later", WithdrawalCancelCode(2004).String())
	assert.Equal(t, "maximum number of sends per hour exceeded: contact support if you require a higher limit or try again later", WithdrawalCancelCode(2005).String())
	assert.Equal(t, "send amount is below the minimum amount required to send on-blockchain", WithdrawalCancelCode(2007).String())
	assert.Equal(t, "two-step verification code required to complete this request: resend request with CB-2FA-Token header", WithdrawalCancelCode(2008).String())
	assert.Equal(t, "withdrawal currency requires tag name", WithdrawalCancelCode(2009).String())
	assert.Equal(t, "amount is invalid", WithdrawalCancelCode(2010).String())
	assert.Equal(t, "withdrawal of this currency is temporarily disabled", WithdrawalCancelCode(2011).String())
	assert.Equal(t, "withdrawal amount exceeds maximum withdrawal amount for currency", WithdrawalCancelCode(2012).String())
	assert.Equal(t, "withdrawal amount exceeds maximum send amount for currency", WithdrawalCancelCode(2013).String())
	assert.Equal(t, "send from fiat accounts is temporarily disabled: try again later", WithdrawalCancelCode(2014).String())
	assert.Equal(t, "payment request has expired", WithdrawalCancelCode(2015).String())
	assert.Equal(t, "send from this account not allowed", WithdrawalCancelCode(2016).String())
	assert.Equal(t, "unable to send to this address", WithdrawalCancelCode(2017).String())
	assert.Equal(t, "recipient address is not whitelisted", WithdrawalCancelCode(2018).String())
	assert.Equal(t, "recipient address whitelist pending: 48 hour hold: try again later", WithdrawalCancelCode(2020).String())
	assert.Equal(t, "unable to send to this user", WithdrawalCancelCode(2021).String())
	assert.Equal(t, "cannot send from an account to itself", WithdrawalCancelCode(2022).String())
	assert.Equal(t, "too many sends, too quickly: wait for some transactions to confirm before sending more", WithdrawalCancelCode(2023).String())
	assert.Equal(t, "invalid email or network address", WithdrawalCancelCode(2024).String())
	assert.Equal(t, "account does not support this currency", WithdrawalCancelCode(2025).String())
	assert.Equal(t, "unknown withdrawal cancel code(1)", WithdrawalCancelCode(1).String())
}

func TestPaymentMethodWithdrawal(t *testing.T) {
	raw := `
{
  "amount": "1.01",
  "currency": "USD",
  "payment_method_id": "payment_method_id"
}
`
	var p PaymentMethodWithdrawalSpec
	require.NoError(t, json.Unmarshal([]byte(raw), &p))
	assert.Equal(t, PaymentMethodWithdrawalSpec{
		Amount:          decimal.NewFromFloat(1.01),
		Currency:        "USD",
		PaymentMethodID: "payment_method_id",
	}, p)
}

func TestCoinbaseAccountWithdrawal(t *testing.T) {
	raw := `
{
  "amount": "1.01",
  "currency": "USD",
  "coinbase_account_id": "coinbase_account_id"
}
`
	var c CoinbaseAccountWithdrawalSpec
	require.NoError(t, json.Unmarshal([]byte(raw), &c))
	require.Equal(t, CoinbaseAccountWithdrawalSpec{
		Amount:            decimal.NewFromFloat(1.01),
		Currency:          "USD",
		CoinbaseAccountID: "coinbase_account_id",
	}, c)
}

func TestCryptoAddressWithdrawal(t *testing.T) {
	raw := `
{
  "add_network_fee_to_total": true,
  "amount": "1.01",
  "crypto_address": "crypto_address",
  "currency": "BTC",
  "destination_tag": "destination_tag",
  "no_destination_tag": false
}
`
	var c CryptoAddressWithdrawalSpec
	require.NoError(t, json.Unmarshal([]byte(raw), &c))
	require.Equal(t, CryptoAddressWithdrawalSpec{
		AddNetworkFeeToTotal: true,
		Amount:               decimal.NewFromFloat(1.01),
		CryptoAddress:        "crypto_address",
		Currency:             "BTC",
		DestinationTag:       "destination_tag",
		NoDestinationTag:     false,
	}, c)
}

func TestCryptoAddress(t *testing.T) {
	t.Run("Params", func(t *testing.T) {
		c := CryptoAddress{
			Currency:      "BTC",
			CryptoAddress: "crypto_address",
		}
		assert.ElementsMatch(t,
			[]string{"currency=BTC", "crypto_address=crypto_address"},
			c.Params())
	})
	t.Run("Unmarshal", func(t *testing.T) {
		raw := `
    {
      "currency": "BTC",
      "crypto_address": "crypto_address"
    }
`
		var c CryptoAddress
		require.NoError(t, json.Unmarshal([]byte(raw), &c))
		require.Equal(t, CryptoAddress{
			Currency:      "BTC",
			CryptoAddress: "crypto_address",
		}, c)
	})
}
