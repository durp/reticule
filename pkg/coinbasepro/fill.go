package coinbasepro

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// Fill
// Settlement and Fees
// Fees are recorded in two stages. Immediately after the matching engine completes a match, the Fill is inserted into
// our datastore. Once the Fill is recorded, a settlement process will settle the Fill and credit both trading counterparties.
type Fill struct {
	// CreatedAt is the fill creation time
	CreatedAt Time `json:"created_at"`
	// Fee indicates the fees charged for this individual Fill.
	Fee decimal.Decimal `json:"fee"`
	// Liquidity indicates if the fill was the result of a liquidity maker or liquidity taker
	Liquidity LiquidityType `json:"liquidity"`
	// OrderID identifies the order associated with the Fill
	OrderID string `json:"order_id"`
	// Price per unit of Product
	Price decimal.Decimal `json:"price"`
	// ProductID identifies the Product associated with the Order
	ProductID ProductID `json:"product_id"`
	// Settled indicates if the Fill has been settled and the counterparties credited
	Settled bool `json:"settled"`
	// Side of Order, `buy` or `sell`
	Side Side `json:"side"`
	// Size indicates the amount of Product filled
	Size decimal.Decimal `json:"size"`
	// TradeID TODO: ??
	TradeID int64 `json:"trade_id"`
}

type LiquidityType string

const (
	// LiquidityTypeMaker indicates the Fill was the result of a liquidity provider
	LiquidityTypeMaker LiquidityType = "M"
	// LiquidityTypeTaker indicates the Fill was the result of a liquidity taker
	LiquidityTypeTaker LiquidityType = "T"
)

type FillFilter struct {
	// OrderID limits the list of Fills to those with the specified OrderID
	OrderID string `json:"order-id"`
	// ProductID limits the list of Fills to those with the specified ProductID
	ProductID ProductID `json:"product-id"`
}

func (f FillFilter) Params() []string {
	var params []string
	if f.OrderID != "" {
		params = append(params, fmt.Sprintf("order_id=%s", f.OrderID))
	}
	if f.ProductID != "" {
		params = append(params, fmt.Sprintf("product_id=%s", f.ProductID))
	}
	return params
}

// Fills is a paged collection of Fills
type Fills struct {
	Fills []*Fill     `json:"fills"`
	Page  *Pagination `json:"page,omitempty"`
}

func (f *Fills) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &f.Fills)
}
