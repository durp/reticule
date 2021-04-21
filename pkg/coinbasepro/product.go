package coinbasepro

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Product
// Only a maximum of one of trading_disabled, cancel_only, post_only, limit_only can be true at once. If none are true,
// the product is trading normally.
// !! When limit_only is true, matching can occur if a limit order crosses the book.
// !! Product ID will not change once assigned to a Product but all other fields are subject to change.
type Product struct {
	// BaseCurrency is the base in the pair of currencies that comprise the Product
	BaseCurrency CurrencyName `json:"base_currency"`
	// BaseIncrement specifies the minimum increment for the BaseCurrency
	BaseIncrement decimal.Decimal `json:"base_increment"`
	// BaseMaxSize defines the maximum order size
	BaseMaxSize decimal.Decimal `json:"base_max_size"`
	// BaseMinSize defines the minimum order size
	BaseMinSize decimal.Decimal `json:"base_min_size"`
	DisplayName string          `json:"display_name"`
	ID          string          `json:"id"`
	// MaxMarketFunds defines the maximum funds allowed
	MaxMarketFunds decimal.Decimal `json:"max_market_funds"`
	// MinMarketFunds defines the minimum funds allowed
	MinMarketFunds decimal.Decimal `json:"min_market_funds"`
	QuoteCurrency  CurrencyName    `json:"quote_currency"`
	QuoteIncrement decimal.Decimal `json:"quote_increment"`
	Status         ProductStatus   `json:"status"`
	// StatusMessage provides any extra information regarding the status, if available
	StatusMessage string `json:"status_message"`
	// CancelOnly indicates whether this product only accepts cancel requests for orders.
	CancelOnly bool `json:"cancel_only"`
	// LimitOnly indicates whether this product only accepts limit orders.
	LimitOnly bool `json:"limit_only"`
	// PostOnly indicates whether only maker orders can be placed. No orders will be matched when post_only mode is active.
	PostOnly bool `json:"post_only"`
	// TradingDisabled indicates whether trading is currently restricted on this product, this includes whether both
	// new orders and order cancellations are restricted.
	TradingDisabled bool `json:"trading_disabled"`
}

// ProductID values could perhaps be dynamically validated from '/products' endpoint
type ProductID string

// ProductStatus has little documentation; all sandbox products have a status value of `online`
type ProductStatus string

// BookLevel represents the level of detail/aggregation in an OrderBook.
// BookLevelBest and BookLevelTop50 are aggregates.
// BookLevelFull requests the entire order book.
type BookLevel int

const (
	// BookLevelUndefined defaults to BookLevel_Best
	BookLevelUndefined BookLevel = 0
	// BookLevelBest requests only the best bid and ask and is aggregated.
	BookLevelBest BookLevel = 1
	// BookLevelTop50 requests the top 50 bids and asks and is aggregated.
	BookLevelTop50 BookLevel = 2
	// BookLevelFull is non-aggregated and returns the entire order book.
	BookLevelFull BookLevel = 3
)

func (p BookLevel) Params() []string {
	level := p
	if p == BookLevelUndefined {
		level = BookLevelBest
	}
	return []string{fmt.Sprintf("level=%d", level)}
}

type AggregatedOrderBook struct {
	Sequence int                   `json:"sequence"`
	Bids     []AggregatedBookEntry `json:"bids"`
	Asks     []AggregatedBookEntry `json:"asks"`
}

func (a *AggregatedBookEntry) UnmarshalJSON(b []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	if len(tmp) != 3 {
		return fmt.Errorf("AggregatedBookEntry must have 3 elements, only found %d", len(tmp))
	}
	if err := json.Unmarshal(tmp[0], &a.Price); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &a.Size); err != nil {
		return err
	}
	return json.Unmarshal(tmp[2], &a.NumOrders)
}

type AggregatedBookEntry struct {
	Price     decimal.Decimal `json:"price"`
	Size      decimal.Decimal `json:"size"`
	NumOrders int             `json:"num_orders"`
}

type OrderBook struct {
	Sequence int         `json:"sequence"`
	Bids     []BookEntry `json:"bids"`
	Asks     []BookEntry `json:"asks"`
}

func (b *BookEntry) UnmarshalJSON(raw []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(raw, &tmp); err != nil {
		return err
	}
	if len(tmp) != 3 {
		return fmt.Errorf("BookEntry must have 3 elements, only found %d", len(tmp))
	}
	if err := json.Unmarshal(tmp[0], &b.Price); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &b.Size); err != nil {
		return err
	}
	return json.Unmarshal(tmp[2], &b.OrderID)
}

type BookEntry struct {
	Price   decimal.Decimal `json:"price"`
	Size    decimal.Decimal `json:"size"`
	OrderID string          `json:"order_id"`
}

// HistoricRateFilter holds filters historic rates for a product by date and sets the granularity of the response.
// If either one of the start or end fields are not provided then both fields will be ignored.
// If a custom time range is not declared then one ending now is selected.
// The granularity field must be one of the following values:
//  {60, 300, 900, 3600, 21600, 86400}.
// Otherwise, the request will be rejected. These values correspond to time slices representing:
// one minute, five minutes, fifteen minutes, one hour, six hours, and one day, respectively.
// If data points are readily available, the response may contain as many as 300 candles and some of candles
// may precede the start value. The maximum number of data points for a single request is 300 candles.
// If the start/end time and granularity results in more than 300 data points, the request will be rejected.
// To retrieve finer granularity data over a larger time range, multiple requests with new start/end ranges must be used.
type HistoricRateFilter struct {
	Granularity Timeslice `json:"granularity"`
	End         Time      `json:"end"`
	Start       Time      `json:"start"`
}

func (h *HistoricRateFilter) Params() []string {
	params := []string{fmt.Sprintf("granularity=%d", h.Granularity)}
	if !h.End.Time().IsZero() {
		end := h.End.Time().Format(time.RFC3339Nano)
		params = append(params, fmt.Sprintf("end=%s", end))
	}
	if !h.Start.Time().IsZero() {
		start := h.Start.Time().Format(time.RFC3339Nano)
		params = append(params, fmt.Sprintf("start=%s", start))
	}
	return params
}

type TimesliceParam time.Duration

func (t TimesliceParam) Validate() error {
	return t.Timeslice().Valid()
}

func (t TimesliceParam) Timeslice() Timeslice {
	return Timeslice(int(math.Round(time.Duration(t).Seconds())))
}

func (t *TimesliceParam) UnmarshalJSON(b []byte) error {
	var s string
	// quote bytes so that marshaller properly scans a number followed by a string as a single string
	err := json.Unmarshal([]byte(fmt.Sprintf("%q", b)), &s)
	if err != nil {
		return err
	}
	d, err := time.ParseDuration(strings.ReplaceAll(s, "\"", ""))
	if err != nil {
		return err
	}
	*t = TimesliceParam(d)
	return nil
}

type Timeslice int

const (
	Timeslice1Minute   Timeslice = 60
	Timeslice5Minutes  Timeslice = 300
	Timeslice15Minutes Timeslice = 900
	Timeslice1Hour     Timeslice = 3600
	Timeslice6Hours    Timeslice = 21600
	Timeslice1Day      Timeslice = 86400
)

func (t Timeslice) Valid() error {
	switch t {
	case Timeslice1Minute, Timeslice5Minutes, Timeslice15Minutes, Timeslice1Hour, Timeslice6Hours, Timeslice1Day:
		return nil
	default:
		return fmt.Errorf("timeslice(%ds) is invalid", t)
	}
}

type HistoricRates struct {
	Candles []*Candle
}

func (h *HistoricRates) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &h.Candles)
}

// A Candle is a common representation of a historic rate.
type Candle struct {
	Close  decimal.Decimal `json:"close"`
	High   decimal.Decimal `json:"high"`
	Low    decimal.Decimal `json:"low"`
	Open   decimal.Decimal `json:"open"`
	Time   Time            `json:"time"`
	Volume decimal.Decimal `json:"volume"`
}

func (c *Candle) UnmarshalJSON(b []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	if len(tmp) != 6 {
		return fmt.Errorf("a Candle must have 6 elements, only found %d", len(tmp))
	}
	var rawTime int64
	if err := json.Unmarshal(tmp[0], &rawTime); err != nil {
		return err
	}
	c.Time = Time(time.Unix(rawTime, 0).UTC())
	if err := json.Unmarshal(tmp[1], &c.Low); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &c.High); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[3], &c.Open); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[4], &c.Close); err != nil {
		return err
	}
	return json.Unmarshal(tmp[5], &c.Volume)
}

// ProductTicker holds snapshot information about the last trade (tick), best bid/ask and 24h volume.
type ProductTicker struct {
	Ask     decimal.Decimal `json:"ask"`
	Bid     decimal.Decimal `json:"bid"`
	Price   decimal.Decimal `json:"price"`
	Size    decimal.Decimal `json:"size"`
	TradeID int             `json:"trade_id"`
	Time    Time            `json:"time"`
	Volume  decimal.Decimal `json:"volume"`
}

// ProductTrades represents the latest trades for a product
type ProductTrades struct {
	Trades []*ProductTrade `json:"trades"`
	Page   *Pagination     `json:"page"`
}

func (p *ProductTrades) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &p.Trades)
}

type ProductTrade struct {
	Price   decimal.Decimal `json:"price"`
	Side    Side            `json:"side"`
	Size    decimal.Decimal `json:"size"`
	Time    Time            `json:"time"`
	TradeID int             `json:"trade_id"`
}

// ProductStats represents a 24 hr stat for the product.
// Volume is in base currency units.
// Open, High, Low are in quote currency units.
type ProductStats struct {
	High        decimal.Decimal `json:"high"`
	Last        decimal.Decimal `json:"last"`
	Low         decimal.Decimal `json:"low"`
	Open        decimal.Decimal `json:"open"`
	Volume      decimal.Decimal `json:"volume"`
	Volume30Day decimal.Decimal `json:"volume_30day"`
}
