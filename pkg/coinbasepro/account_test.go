package coinbasepro

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccount(t *testing.T) {
	account := Account{
		Available:      decimal.NewFromFloat(1.0), // Note: trailing zeroes are lost
		Balance:        decimal.NewFromFloat(2.2),
		Currency:       "USD",
		Hold:           decimal.NewFromFloat(3.03),
		ID:             "id",
		ProfileID:      "profile_id",
		TradingEnabled: true,
	}
	b, err := json.Marshal(account)
	require.NoError(t, err)
	assert.JSONEq(t, `
{
	"available": "1",
	"balance": "2.2",
	"currency": "USD",
	"hold": "3.03",
	"id": "id",
	"profile_id": "profile_id",
	"trading_enabled": true
}`,
		string(b))
}

func TestLedger(t *testing.T) {
	var timestamp Time
	err := timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z"))
	require.NoError(t, err)
	ledgerEntry := LedgerEntry{
		Amount:    decimal.NewFromFloat(1.0),
		Balance:   decimal.NewFromFloat(2.2),
		CreatedAt: timestamp,
		Details: LedgerDetails{
			OrderID:   "order_id",
			ProductID: "product_id",
			TradeID:   "trade_id",
		},
		ID:   "id",
		Type: LedgerEntryTypeConversion,
	}
	t.Run("LedgerEntryMarshal", func(t *testing.T) {
		b, err := json.Marshal(ledgerEntry)
		require.NoError(t, err)
		assert.JSONEq(t, `
{
  "amount": "1",
  "balance": "2.2",
  "created_at": "2021-04-09T19:04:58.964459Z",
  "details": {
    "order_id": "order_id",
    "trade_id": "trade_id",
    "product_id": "product_id"
  },
  "id": "id",
  "type": "conversion"
}`, string(b))
	})
	t.Run("LedgerMarshal", func(t *testing.T) {
		ledger := Ledger{
			Entries: []*LedgerEntry{&ledgerEntry},
			Page: &Pagination{
				After:  "alpha",
				Before: "omega",
			},
		}
		b, err := json.Marshal(ledger)
		require.NoError(t, err)
		assert.JSONEq(t, `
{
  "entries": [
    {
      "amount": "1",
      "balance": "2.2",
      "created_at": "2021-04-09T19:04:58.964459Z",
      "details": {
        "order_id": "order_id",
        "trade_id": "trade_id",
        "product_id": "product_id"
      },
      "id": "id",
      "type": "conversion"
    }
  ],
  "page": {
    "after": "alpha",
    "before": "omega"
  }
}`,
			string(b))
	})
	t.Run("LedgerUnmarshal", func(t *testing.T) {
		var ledger Ledger
		err := json.Unmarshal([]byte(`
[
	{
		"amount": "1",
		"balance": "2.2",
		"created_at": "2021-04-09T19:04:58.964459Z",
		"details": {
			"order_id": "order_id",
			"trade_id": "trade_id",
			"product_id": "product_id"
		},
		"id": "id",
		"type": "conversion"
	}
]`), &ledger)
		require.NoError(t, err)
		// Note: Marshal does not handle Page population
		assert.Equal(t, Ledger{
			Entries: []*LedgerEntry{&ledgerEntry},
		}, ledger)
	})
}

func TestHolds(t *testing.T) {
	var timestamp Time
	err := timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z"))
	require.NoError(t, err)
	hold := Hold{
		AccountID: "account_id",
		Amount:    decimal.NewFromFloat(1.0),
		CreatedAt: timestamp,
		Ref:       "ref",
		Type:      HoldTypeOpenOrders,
		UpdatedAt: timestamp,
	}
	t.Run("HoldMarshal", func(t *testing.T) {
		b, err := json.Marshal(hold)
		require.NoError(t, err)
		assert.JSONEq(t, `
{
	"account_id": "account_id",
	"amount": "1",
	"created_at": "2021-04-09T19:04:58.964459Z",
	"ref": "ref",
	"type": "order",
	"updated_at": "2021-04-09T19:04:58.964459Z"
}`, string(b))
	})
	t.Run("HoldsMarshal", func(t *testing.T) {
		holds := Holds{
			Holds: []*Hold{&hold},
			Page: &Pagination{
				After:  "alpha",
				Before: "omega",
			},
		}
		b, err := json.Marshal(holds)
		require.NoError(t, err)
		assert.JSONEq(t,
			`{
  "holds": [
    {
      "account_id": "account_id",
			"amount": "1",
			"created_at": "2021-04-09T19:04:58.964459Z",
			"ref": "ref",
			"type": "order",
			"updated_at": "2021-04-09T19:04:58.964459Z"
    }
  ],
  "page": {
    "after": "alpha",
    "before": "omega"
  }
}`, string(b))
	})
	t.Run("HoldsUnmarshal", func(t *testing.T) {
		var holds Holds
		err := json.Unmarshal([]byte(`
[
	{
		"account_id": "account_id",
		"amount": "1",
		"created_at": "2021-04-09T19:04:58.964459Z",
		"ref": "ref",
		"type": "order",
		"updated_at": "2021-04-09T19:04:58.964459Z"
	}
]`), &holds)
		require.NoError(t, err)
		assert.Equal(t, Holds{
			Holds: []*Hold{&hold},
		}, holds)
	})
}
