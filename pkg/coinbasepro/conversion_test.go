package coinbasepro

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestStablecoinConversionSpec_UnmarshalJSON(t *testing.T) {
	raw := `{"from":"BTC","to":"USD","amount":"1.01"}`
	var s StablecoinConversionSpec
	err := s.UnmarshalJSON([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, StablecoinConversionSpec{
		From:   "BTC",
		To:     "USD",
		Amount: decimal.NewFromFloat(1.01),
	}, s)
}
