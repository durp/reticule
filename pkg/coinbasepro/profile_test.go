package coinbasepro

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileFilter(t *testing.T) {
	f := ProfileFilter{
		Active: true,
	}
	assert.ElementsMatch(t, []string{"active"}, f.Params())
}

func TestProfileTransferSpec(t *testing.T) {
	raw := `
  {
    "amount": "1.01",
    "currency": "USD",
    "from": "from",
    "to": "to"
  }
`
	var p ProfileTransferSpec
	err := json.Unmarshal([]byte(raw), &p)
	require.NoError(t, err)
	assert.Equal(t, ProfileTransferSpec{
		Amount:   decimal.NewFromFloat(1.01),
		Currency: "USD",
		From:     "from",
		To:       "to",
	}, p)
}
