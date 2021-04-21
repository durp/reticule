package coinbasepro

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFills_UnmarshalJSON(t *testing.T) {
	var timestamp Time
	err := timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z"))
	require.NoError(t, err)
	raw := `
[
  {
    "created_at": "2021-04-09T19:04:58.964459Z",
    "fee": "1.05",
    "liquidity": "M",
    "order_id": "order_id",
    "price": "1.01",
    "product_id": "BTC-ETH",
    "settled": true,
    "side": "sell",
    "size": "2.1",
    "trade_id": 12345
  }
]
`
	var fills Fills
	err = json.Unmarshal([]byte(raw), &fills)
	require.NoError(t, err)
	assert.Equal(t, []*Fill{{
		CreatedAt: timestamp,
		Fee:       decimal.NewFromFloat(1.05),
		Liquidity: LiquidityTypeMaker,
		OrderID:   "order_id",
		Price:     decimal.NewFromFloat(1.01),
		ProductID: "BTC-ETH",
		Settled:   true,
		Side:      "sell",
		Size:      decimal.NewFromFloat(2.1),
		TradeID:   12345,
	},
	}, fills.Fills)
}

func TestFillFilter_Params(t *testing.T) {
	filter := FillFilter{
		OrderID:   "order_id",
		ProductID: "BTC-USD",
	}
	assert.ElementsMatch(t, []string{"order_id=order_id", "product_id=BTC-USD"}, filter.Params())
}
