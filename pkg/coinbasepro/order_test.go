package coinbasepro

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderFilter(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("ValidEmpty", func(t *testing.T) {
			var filter OrderFilter
			err := filter.Validate()
			assert.NoError(t, err)
		})
		t.Run("ValidStatus", func(t *testing.T) {
			filter := OrderFilter{
				Status: []OrderStatusParam{OrderStatusParamActive},
			}
			err := filter.Validate()
			assert.NoError(t, err)
		})
		t.Run("Invalid", func(t *testing.T) {
			filter := OrderFilter{
				Status: []OrderStatusParam{"blah"},
			}
			err := filter.Validate()
			require.Error(t, err)
			assert.Regexp(t, "status.*blah.* not valid", err.Error())
		})
	})
	t.Run("Params", func(t *testing.T) {
		var f OrderFilter
		assert.Empty(t, f.Params())
		filter := OrderFilter{
			ProductID: "BTC-USD",
			Status: []OrderStatusParam{
				OrderStatusParamActive, OrderStatusParamDone, OrderStatusParamOpen,
				OrderStatusParamPending, OrderStatusParamReceived, OrderStatusParamSettled,
			},
		}
		assert.ElementsMatch(t, []string{
			"product_id=BTC-USD",
			"status=active", "status=done", "status=open", "status=pending", "status=received", "status=settled",
		}, filter.Params())
		filter = OrderFilter{
			Status: []OrderStatusParam{OrderStatusParamAll},
		}
		assert.ElementsMatch(t, []string{
			"status=all",
		}, filter.Params())
	})
}

func TestLimitOrder(t *testing.T) {
	stopPrice := decimal.NewFromFloat(1.01)
	validLimitOrder := func() LimitOrder {
		return LimitOrder{
			ClientOrderID:       "client_oid",
			ProductID:           "BTC-USD",
			SelfTradePrevention: SelfTradeDecrementAndCancel,
			Side:                SideBuy,
			Stop:                StopLoss,
			StopPrice:           &stopPrice,
			Type:                OrderTypeLimit,
			CancelAfter:         "1,1,1",
			PostOnly:            true,
			Price:               decimal.NewFromFloat(2.12),
			Size:                decimal.NewFromFloat(3.23),
			TimeInForce:         TimeInForceGoodTillTime,
		}
	}
	limitOrder := validLimitOrder()
	err := limitOrder.Validate()
	require.NoError(t, err)

	t.Run("Unmarshal", func(t *testing.T) {
		raw := `
{
  "client_oid": "client_oid",
	"product_id": "BTC-USD",
  "stp": "",
  "side": "buy",
  "stop": "loss",
  "stop_price": "1.01",
  "cancel_after": "1,1,1",
  "post_only": true,
  "price": "2.12",
  "size": "3.23",
  "time_in_force": "GTT"
}
`
		var l LimitOrder
		err := json.Unmarshal([]byte(raw), &l)
		require.NoError(t, err)
		assert.Equal(t, limitOrder, l)
	})

	t.Run("Validation", func(t *testing.T) {
		t.Run("OrderTypeMustBeLimit", func(t *testing.T) {
			l := validLimitOrder()
			l.Type = OrderTypeMarket
			err = l.Validate()
			require.Error(t, err)
			assert.Regexp(t, "type", err.Error())
		})
		t.Run("ProductIDRequired", func(t *testing.T) {
			l := validLimitOrder()
			l.ProductID = ""
			err = l.Validate()
			require.Error(t, err)
			assert.Regexp(t, "product_id", err.Error())
		})
		t.Run("SideMustBeValid", func(t *testing.T) {
			l := validLimitOrder()
			l.Side = "blah"
			err = l.Validate()
			require.Error(t, err)
			assert.Regexp(t, "side", err.Error())
		})
		t.Run("StopMustBeValid", func(t *testing.T) {
			l := validLimitOrder()
			l.Stop = "blah"
			err = l.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*blah.* not valid", err.Error())
		})
		t.Run("StopPriceMustBeValid", func(t *testing.T) {
			l := validLimitOrder()
			l.Stop = StopLoss
			l.StopPrice = nil
			err = l.Validate()
			require.Error(t, err)
			assert.Regexp(t, "requires .*stop_price.*", err.Error())
		})
	})
}

func TestMarkerOrder(t *testing.T) {

	validMarketOrder := func() MarketOrder {
		return MarketOrder{
			ClientOrderID:       "client_oid",
			ProductID:           "BTC-USD",
			SelfTradePrevention: SelfTradeDecrementAndCancel,
			Side:                "sell",
			Stop:                StopEntry,
			StopPrice:           nullable(t, "1.01"),
			Type:                OrderTypeMarket,
			Funds:               nullable(t, "2.12"),
			Size:                nullable(t, "3.23"),
		}
	}
	marketOrder := validMarketOrder()
	err := marketOrder.Validate()
	require.NoError(t, err)

	t.Run("Unmarshal", func(t *testing.T) {
		raw := `
{
  "client_oid": "client_oid",
	"product_id": "BTC-USD",
  "stp": "",
  "side": "sell",
  "stop": "entry",
  "stop_price": "1.01",
  "funds": "2.12",
  "size": "3.23"
}
`
		var m MarketOrder
		err := json.Unmarshal([]byte(raw), &m)
		require.NoError(t, err)
		assert.Equal(t, marketOrder, m)
	})

	t.Run("Validation", func(t *testing.T) {
		t.Run("OrderTypeMustBeMarket", func(t *testing.T) {
			m := validMarketOrder()
			m.Type = OrderTypeLimit
			err = m.Validate()
			require.Error(t, err)
			assert.Regexp(t, "type", err.Error())
		})
		t.Run("ProductIDRequired", func(t *testing.T) {
			m := validMarketOrder()
			m.ProductID = ""
			err = m.Validate()
			require.Error(t, err)
			assert.Regexp(t, "product_id", err.Error())
		})
		t.Run("SideMustBeValid", func(t *testing.T) {
			m := validMarketOrder()
			m.Side = "blah"
			err = m.Validate()
			require.Error(t, err)
			assert.Regexp(t, "side", err.Error())
		})
		t.Run("StopMustBeValid", func(t *testing.T) {
			m := validMarketOrder()
			m.Stop = "blah"
			err = m.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*blah.* not valid", err.Error())
		})
		t.Run("StopPriceMustBeValid", func(t *testing.T) {
			m := validMarketOrder()
			m.Stop = StopLoss
			m.StopPrice = nil
			err = m.Validate()
			require.Error(t, err)
			assert.Regexp(t, "requires .*stop_price.*", err.Error())
		})
		t.Run("WithoutFundsASizeIsRequired", func(t *testing.T) {
			m := validMarketOrder()
			m.Funds = nil
			m.Size = nil
			err = m.Validate()
			require.Error(t, err)
			assert.Regexp(t, "without .*funds.* .*size.* required", err.Error())
		})
	})
}

func TestStop(t *testing.T) {
	t.Run("Stop", func(t *testing.T) {
		positivePrice := decimal.NewFromFloat(1.01)
		t.Run("Invalid", func(t *testing.T) {
			stop := Stop("blah")
			err := stop.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*blah.* not valid", err.Error())
		})
		t.Run("StopPrice", func(t *testing.T) {
			t.Run("InvalidStop", func(t *testing.T) {
				stop := Stop("blah")
				err := stop.ValidatePrice(&positivePrice)
				assert.Error(t, err)
				assert.Regexp(t, ".*blah.* not valid", err.Error())
			})
			t.Run("StopNone", func(t *testing.T) {
				stop := StopNone
				err := stop.ValidatePrice(&positivePrice)
				require.Error(t, err)
				assert.Regexp(t, ".*stop_price.*", err.Error())
			})
			t.Run("StopLoss", func(t *testing.T) {
				stop := StopLoss
				t.Run("Valid", func(t *testing.T) {
					err := stop.ValidatePrice(&positivePrice)
					assert.NoError(t, err)
				})
				t.Run("Invalid", func(t *testing.T) {
					err := stop.ValidatePrice(nil)
					require.Error(t, err)
					assert.Regexp(t, ".*loss.* requires a .*stop_price", err.Error())
				})
			})
			t.Run("StopEntry", func(t *testing.T) {
				stop := StopEntry
				t.Run("Valid", func(t *testing.T) {
					err := stop.ValidatePrice(&positivePrice)
					assert.NoError(t, err)
				})
				t.Run("Invalid", func(t *testing.T) {
					err := stop.ValidatePrice(nil)
					require.Error(t, err)
					assert.Regexp(t, ".*entry.* requires a .*stop_price", err.Error())
				})
			})
		})
	})
}

func TestSelfTrade(t *testing.T) {
	assert.NoError(t, SelfTradeDecrementAndCancel.Validate())
	assert.NoError(t, SelfTradeCancelNewest.Validate())
	assert.NoError(t, SelfTradeCancelOldest.Validate())
	assert.NoError(t, SelfTradeCancelBoth.Validate())
	assert.Error(t, SelfTrade("blah").Validate())

	var s SelfTrade
	err := json.Unmarshal([]byte(`""`), &s)
	require.NoError(t, err)
	assert.Equal(t, SelfTradeDecrementAndCancel, s)
}

func TestSide(t *testing.T) {
	assert.NoError(t, SideBuy.Validate())
	assert.NoError(t, SideSell.Validate())
	assert.Error(t, Side("blah").Validate())
}

func TestTimeInForce(t *testing.T) {
	t.Run("Validation", func(t *testing.T) {
		assert.NoError(t, TimeInForceGoodTillCanceled.Validate())
		assert.NoError(t, TimeInForceGoodTillTime.Validate())
		assert.NoError(t, TimeInForceImmediateOrCancel.Validate())
		assert.NoError(t, TimeInForceFillOrKill.Validate())
		assert.Error(t, TimeInForce("blah").Validate())
	})
	t.Run("CancelAfter", func(t *testing.T) {
		err := TimeInForceGoodTillTime.ValidateCancelAfter("")
		require.Error(t, err)
		assert.Regexp(t, ".*GTT.* requires .*cancel_after.*", err.Error())
		err = TimeInForceGoodTillTime.ValidateCancelAfter("1,1,1")
		require.NoError(t, err)
	})
}

func TestOrders(t *testing.T) {
	var timestamp Time
	err := timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z"))
	require.NoError(t, err)
	raw := `
[
  {
     "created_at": "2021-04-09T19:04:58.964459Z",
     "done_at": "2021-04-09T19:04:58.964459Z",
     "done_reason": "done_reason",
     "executed_value": "1.01",
     "fill_fees": "2.12",
     "filled_size": "3.23",
     "funds": "4.34",
     "id": "id",
     "post_only": true,
     "product_id": "BTC-USD",
     "settled": true,
     "side": "buy",
     "size": "5.45",
     "specified_funds": "6.56",
     "status": "settled",
     "stp": "dc",
     "type": "market"
  }
]
`
	var o Orders
	err = json.Unmarshal([]byte(raw), &o)
	require.NoError(t, err)
	assert.Equal(t, []*Order{
		{
			CreatedAt:           timestamp,
			DoneAt:              timestamp,
			DoneReason:          "done_reason",
			ExecutedValue:       nullable(t, "1.01"),
			FillFees:            nullable(t, "2.12"),
			FilledSize:          nullable(t, "3.23"),
			Funds:               nullable(t, "4.34"),
			ID:                  "id",
			PostOnly:            true,
			ProductID:           "BTC-USD",
			Settled:             true,
			Side:                SideBuy,
			Size:                decimal.NewFromFloat(5.45),
			SpecifiedFunds:      nullable(t, "6.56"),
			Status:              OrderStatusSettled,
			SelfTradePrevention: SelfTradeDecrementAndCancel,
			Type:                OrderTypeMarket,
		},
	}, o.Orders)
}

func nullable(t *testing.T, s string) *decimal.Decimal {
	t.Helper()
	if s == "" {
		return nil
	}
	v, err := decimal.NewFromString(s)
	require.NoError(t, err)
	return &v
}
