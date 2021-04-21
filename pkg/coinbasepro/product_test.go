package coinbasepro

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookLevel(t *testing.T) {
	t.Run("Params", func(t *testing.T) {
		assert.Equal(t, []string{"level=1"}, BookLevelUndefined.Params())
		assert.Equal(t, []string{"level=1"}, BookLevelBest.Params())
		assert.Equal(t, []string{"level=2"}, BookLevelTop50.Params())
		assert.Equal(t, []string{"level=3"}, BookLevelFull.Params())
	})
}

func TestAggregatedBookEntry(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			raw := `[1.01,2.12,111]`
			var a AggregatedBookEntry
			err := json.Unmarshal([]byte(raw), &a)
			require.NoError(t, err)
			assert.Equal(t, AggregatedBookEntry{
				Price:     decimal.NewFromFloat(1.01),
				Size:      decimal.NewFromFloat(2.12),
				NumOrders: 111,
			}, a)
		})
		t.Run("Error", func(t *testing.T) {
			require.Error(t, json.Unmarshal([]byte(`[1.01,2.12]`), &AggregatedBookEntry{}))
			require.Error(t, json.Unmarshal([]byte(`["X",2.12,111]`), &AggregatedBookEntry{}))
			require.Error(t, json.Unmarshal([]byte(`[1.01,"X",111]`), &AggregatedBookEntry{}))
			require.Error(t, json.Unmarshal([]byte(`[1.01,2.12,"X"]`), &AggregatedBookEntry{}))
		})
	})
}

func TestBookEntry(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			raw := `[1.01,2.12,"order_id"]`
			var b BookEntry
			err := json.Unmarshal([]byte(raw), &b)
			require.NoError(t, err)
			assert.Equal(t, BookEntry{
				Price:   decimal.NewFromFloat(1.01),
				Size:    decimal.NewFromFloat(2.12),
				OrderID: "order_id",
			}, b)
		})
		t.Run("Error", func(t *testing.T) {
			require.Error(t, json.Unmarshal([]byte(`[1.01,2.12]`), &BookEntry{}))
			require.Error(t, json.Unmarshal([]byte(`["X",2.12,"order_id"]`), &BookEntry{}))
			require.Error(t, json.Unmarshal([]byte(`[1.01,"X","order_id"]`), &BookEntry{}))
			require.Error(t, json.Unmarshal([]byte(`[1.01,2.12,[]]`), &BookEntry{}))
		})
	})
}

func TestTimesliceParam(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "60s")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "1m")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "300s")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "5m")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "900s")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "15m")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "3600s")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "60m")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "1h")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "21600s")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "6h")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "86400s")).Validate())
		assert.NoError(t, TimesliceParam(mustParseDuration(t, "24h")).Validate())
		assert.Error(t, TimesliceParam(mustParseDuration(t, "1s")).Validate())
	})
	t.Run("Unmarshal", func(t *testing.T) {
		var p TimesliceParam
		require.NoError(t, json.Unmarshal([]byte(`"300s"`), &p))
	})
}

func TestHistoricRateFilter(t *testing.T) {
	var timestamp Time
	err := timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z"))
	require.NoError(t, err)
	h := HistoricRateFilter{
		Granularity: Timeslice1Day,
		End:         timestamp,
		Start:       timestamp,
	}
	assert.ElementsMatch(t, []string{
		"granularity=86400",
		"end=2021-04-09T19:04:58.964459Z",
		"start=2021-04-09T19:04:58.964459Z",
	}, h.Params())
}

func TestHistoricRates(t *testing.T) {
	var timestamp Time
	err := timestamp.UnmarshalJSON([]byte("2021-04-13T21:35:00Z"))
	require.NoError(t, err)
	raw := `
[
  [ 1618349700, 1.01, 2.12, 3.23, 4.34, 5.45 ]
]
`
	var h HistoricRates
	err = json.Unmarshal([]byte(raw), &h)
	require.NoError(t, err)
	assert.Equal(t, []*Candle{{
		Time:   timestamp,
		Low:    decimal.NewFromFloat(1.01),
		High:   decimal.NewFromFloat(2.12),
		Open:   decimal.NewFromFloat(3.23),
		Close:  decimal.NewFromFloat(4.34),
		Volume: decimal.NewFromFloat(5.45),
	}}, h.Candles)
}

func TestCandle(t *testing.T) {
	assert.NoError(t, json.Unmarshal([]byte(`[ 1618349700, 1.01, 2.12, 3.23, 4.34, 5.45 ]`), &Candle{}))
	assert.Error(t, json.Unmarshal([]byte(`[ "X", 1.01, 2.12, 3.23, 4.34, 5.45 ]`), &Candle{}))
	assert.Error(t, json.Unmarshal([]byte(`[ 1618349700, "X", 2.12, 3.23, 4.34, 5.45 ]`), &Candle{}))
	assert.Error(t, json.Unmarshal([]byte(`[ 1618349700, 1.01, "X", 3.23, 4.34, 5.45 ]`), &Candle{}))
	assert.Error(t, json.Unmarshal([]byte(`[ 1618349700, 1.01, 2.12, "X", 4.34, 5.45 ]`), &Candle{}))
	assert.Error(t, json.Unmarshal([]byte(`[ 1618349700, 1.01, 2.12, 3.23, "X", 5.45 ]`), &Candle{}))
	assert.Error(t, json.Unmarshal([]byte(`[ 1618349700, 1.01, 2.12, 3.23, 4.34, "X" ]`), &Candle{}))
	assert.Error(t, json.Unmarshal([]byte(`[]`), &Candle{}))
}

func TestProductTrades(t *testing.T) {
	var timestamp Time
	err := timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z"))
	require.NoError(t, err)
	raw := `
[
  {
    "price": "1.01",
    "side": "buy",
    "size": "2.12",
    "time": "2021-04-09T19:04:58.964459Z",
    "trade_id": 1
  }
]
`
	var p ProductTrades
	err = json.Unmarshal([]byte(raw), &p)
	require.NoError(t, err)
}

func mustParseDuration(t *testing.T, s string) time.Duration {
	t.Helper()
	d, err := time.ParseDuration(s)
	require.NoError(t, err)
	return d
}
