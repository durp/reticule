package coinbasepro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/durp/reticule/pkg/phizog"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
)

// Private

// ListAccounts retrieves the list of trading accounts belonging to the Profile of the API key. The list is not paginated.
func (c *Client) ListAccounts(ctx context.Context) ([]Account, error) {
	var accounts []Account
	return accounts, c.api.Get(ctx, "/accounts/", &accounts)
}

// GetAccount retrieves the detailed representation of a trading Account. The requested Account must belong to the current Profile.
func (c *Client) GetAccount(ctx context.Context, accountID string) (Account, error) {
	var account Account
	return account, c.api.Get(ctx, fmt.Sprintf("/accounts/%s", accountID), &account)
}

// GetLedger retrieves a paginated list of Account activity for the current Profile.
func (c *Client) GetLedger(ctx context.Context, accountID string, pagination PaginationParams) (Ledger, error) {
	if err := pagination.Validate(); err != nil {
		return Ledger{}, err
	}
	var ledger Ledger
	query := query(pagination.Params())
	return ledger, c.api.Get(ctx, fmt.Sprintf("/accounts/%s/ledger/%s", accountID, query), &ledger)
}

// GetHolds retrieves the list of Holds for the Account. The requested Account must belong to the current Profile.
func (c *Client) GetHolds(ctx context.Context, accountID string, pagination PaginationParams) (Holds, error) {
	if err := pagination.Validate(); err != nil {
		return Holds{}, err
	}
	var holds Holds
	query := query(pagination.Params())
	return holds, c.api.Get(ctx, fmt.Sprintf("/accounts/%s/holds/%s", accountID, query), &holds)
}

// CreateLimitOrder creates a LimitOrder to trade a Product with specified Price and Size limits.
func (c *Client) CreateLimitOrder(ctx context.Context, limitOrder LimitOrder) (Order, error) {
	if err := limitOrder.Validate(); err != nil {
		return Order{}, err
	}
	var order Order
	return order, c.api.Post(ctx, "/orders/", limitOrder, &order)
}

// CreateMarketOrder creates a MarketOrder with no pricing guarantees. A MarketOrder makes it easy to trade specific
// amounts of a Product without specifying prices.
func (c *Client) CreateMarketOrder(ctx context.Context, marketOrder MarketOrder) (Order, error) {
	if err := marketOrder.Validate(); err != nil {
		return Order{}, err
	}
	var order Order
	return order, c.api.Post(ctx, "/orders/", marketOrder, &order)
}

// CancelOrder cancels a previously placed order. orderID is mandatory, productID is optional but will make the request
// more performant. If the Order had no matches during its lifetime, it may be subject to purge and as a result will
// no longer available via GetOrder.
// Requires "trade" permission.
func (c *Client) CancelOrder(ctx context.Context, spec CancelOrderSpec) (map[string]interface{}, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	var resp map[string]interface{}
	return resp, c.api.Do(ctx, "DELETE", "/orders/"+spec.Path()+query(spec.Params()), nil, &resp)
}

// GetOrders retrieves a paginated list of the current open orders for the current Profile. Only open or un-settled
// orders are returned by default. An OrderFilter can be used to further refine the request.
func (c *Client) GetOrders(ctx context.Context, filter OrderFilter, pagination PaginationParams) (Orders, error) {
	params := append(filter.Params(), pagination.Params()...)
	var orders Orders
	return orders, c.api.Get(ctx, fmt.Sprintf("/orders/%s", query(params)), &orders)
}

// GetOrder retrieves the details of a single Order. The requested Order must belong to the current Profile.
func (c *Client) GetOrder(ctx context.Context, orderID string) (Order, error) {
	var order Order
	return order, c.api.Get(ctx, fmt.Sprintf("/orders/%s", orderID), &order)
}

// GetClientOrder retrieves the details of a single Order using a client-provided identifier.
// The requested Order must belong to the current Profile.
func (c *Client) GetClientOrder(ctx context.Context, clientID string) (Order, error) {
	var order Order
	return order, c.api.Get(ctx, fmt.Sprintf("/orders/client:%s", clientID), &order)
}

// GetFills retrieves a paginated list of recent Fills for the current Profile.
func (c *Client) GetFills(ctx context.Context, filter FillFilter, pagination PaginationParams) (Fills, error) {
	params := append(filter.Params(), pagination.Params()...)
	var fills Fills
	return fills, c.api.Get(ctx, fmt.Sprintf("/fills/%s", query(params)), &fills)
}

// GetLimits retrieves the payment method transfer limits and per currency buy/sell limits for the current Profile.
func (c *Client) GetLimits(ctx context.Context) (Limits, error) {
	var limits Limits
	return limits, c.api.Get(ctx, "/users/self/exchange-limits/", &limits)
}

// GetDeposits retrieves a paginated list of Deposits, in descending order by CreatedAt time.
func (c *Client) GetDeposits(ctx context.Context, filter DepositFilter, pagination PaginationParams) (Deposits, error) {
	params := append(filter.Params(), pagination.Params()...)
	var deposits Deposits
	err := c.api.Get(ctx, fmt.Sprintf("/transfers/%s", query(params)), &deposits)
	if err != nil {
		return Deposits{}, err
	}
	// Deposits are a flavor of Transfer and the coinbasepro API cannot filter by multiple types
	// TODO: this potentially screws up pagination
	transferDeposits := make([]*Deposit, 0, len(deposits.Deposits))
	for _, transfer := range deposits.Deposits {
		if transfer.Type == DepositTypeInternal || transfer.Type == DepositTypeDeposit {
			transferDeposits = append(transferDeposits, transfer)
		}
	}
	if len(transferDeposits) == 0 {
		deposits.Page = &Pagination{}
	}
	deposits.Deposits = transferDeposits
	return deposits, nil
}

// GetDeposit retrieves the details for a single Deposit. The Deposit must belong to the current Profile.
func (c *Client) GetDeposit(ctx context.Context, depositID string) (Deposit, error) {
	var deposit Deposit
	return deposit, c.api.Get(ctx, fmt.Sprintf("/transfers/%s", depositID), &deposit)
}

// CreatePaymentMethodDeposit creates a Deposit of funds from an external payment method. Use ListPaymentMethods to
// retrieve details of available PaymentMethods.
func (c *Client) CreatePaymentMethodDeposit(ctx context.Context, paymentMethodDeposit PaymentMethodDepositSpec) (Deposit, error) {
	result := struct {
		ID string `json:"id"`
	}{}
	err := c.api.Post(ctx, "/deposits/payment-method/", paymentMethodDeposit, &result)
	if err != nil {
		return Deposit{}, err
	}
	// POST coinbasepro response is partial; retrieve full representation
	return c.GetDeposit(ctx, result.ID)
}

// CreateCoinbaseAccountDeposit creates a Deposit of funds from a CoinbaseAccount. Funds can be moved between
// CoinbaseAccounts and Coinbase Pro trading Accounts within daily limits. Moving funds between Coinbase and Coinbase Pro
// is instant and free. Use ListCoinbaseAccounts to retrieve available Coinbase accounts.
func (c *Client) CreateCoinbaseAccountDeposit(ctx context.Context, coinbaseAccountDeposit CoinbaseAccountDeposit) (Deposit, error) {
	result := struct {
		ID string `json:"id"`
	}{}
	err := c.api.Post(ctx, "/deposits/coinbase-account/", coinbaseAccountDeposit, &result)
	if err != nil {
		return Deposit{}, err
	}
	// POST coinbasepro response is partial; retrieve full representation
	return c.GetDeposit(ctx, result.ID)
}

// CreateCryptoDepositAddress generates an address for crypto deposits into a CoinbaseAccount.
func (c *Client) CreateCryptoDepositAddress(ctx context.Context, coinbaseAccountID string) (CryptoDepositAddress, error) {
	var cryptoDepositAddress CryptoDepositAddress
	return cryptoDepositAddress, c.api.Post(ctx, fmt.Sprintf("/coinbase-accounts/%s/addresses/", coinbaseAccountID), nil, &cryptoDepositAddress)
}

// GetWithdrawals retrieves a paginated list of Withdrawals for the current Profile, in descending order by CreatedAt time.
func (c *Client) GetWithdrawals(ctx context.Context, filter WithdrawalFilter, pagination PaginationParams) (Withdrawals, error) {
	params := append(filter.Params(), pagination.Params()...)
	var withdrawals Withdrawals
	err := c.api.Get(ctx, fmt.Sprintf("/transfers/%s", query(params)), &withdrawals)
	if err != nil {
		return Withdrawals{}, err
	}
	// Withdrawals are a flavor of Transfer and the coinbasepro API cannot filter by multiple types
	// TODO: this potentially screws up pagination
	transferWithdrawals := make([]*Withdrawal, 0, len(withdrawals.Withdrawals))
	for _, transfer := range withdrawals.Withdrawals {
		if transfer.Type == WithdrawalTypeInternal || transfer.Type == WithdrawalTypeWithdraw {
			transferWithdrawals = append(transferWithdrawals, transfer)
		}
	}
	if len(transferWithdrawals) == 0 {
		withdrawals.Page = &Pagination{}
	}
	withdrawals.Withdrawals = transferWithdrawals
	return withdrawals, nil
}

// GetWithdrawal retrieves the details of a single Withdrawal. The Withdrawal must belong to the current Profile.
func (c *Client) GetWithdrawal(ctx context.Context, withdrawalID string) (Withdrawal, error) {
	var withdrawal Withdrawal
	return withdrawal, c.api.Get(ctx, fmt.Sprintf("/transfers/%s", withdrawalID), &withdrawal)
}

// CreatePaymentMethodWithdrawal creates a Withdrawal of funds to an external PaymentMethod. Use ListPaymentMethods to
// retrieve details of available PaymentMethods.
func (c *Client) CreatePaymentMethodWithdrawal(ctx context.Context, paymentMethodWithdrawal PaymentMethodWithdrawalSpec) (Withdrawal, error) {
	result := struct {
		ID string `json:"id"`
	}{}
	err := c.api.Post(ctx, "/withdrawals/payment-method/", paymentMethodWithdrawal, &result)
	if err != nil {
		return Withdrawal{}, err
	}
	// POST coinbasepro response is partial; retrieve full representation
	return c.GetWithdrawal(ctx, result.ID)
}

// CreateCoinbaseAccountWithdrawal creates a Withdrawal of funds to a CoinbaseAccount. Funds can be moved between
// CoinbaseAccounts and Coinbase Pro trading Accounts within daily limits. Moving funds between Coinbase and Coinbase Pro
// is instant and free. Use ListCoinbaseAccounts to retrieve available Coinbase accounts.
func (c *Client) CreateCoinbaseAccountWithdrawal(ctx context.Context, coinbaseAccountWithdrawal CoinbaseAccountWithdrawalSpec) (Withdrawal, error) {
	result := struct {
		ID string `json:"id"`
	}{}
	err := c.api.Post(ctx, "/withdrawals/coinbase-account/", coinbaseAccountWithdrawal, &result)
	if err != nil {
		return Withdrawal{}, err
	}
	// POST coinbasepro response is partial; retrieve full representation
	return c.GetWithdrawal(ctx, result.ID)
}

// CreateCryptoAddressWithdrawal creates a Withdrawal of funds to a crypto address.
func (c *Client) CreateCryptoAddressWithdrawal(ctx context.Context, cryptoAddressWithdrawal CryptoAddressWithdrawalSpec) (Withdrawal, error) {
	result := struct {
		ID string `json:"id"`
	}{}
	err := c.api.Post(ctx, "/withdrawals/crypto/", cryptoAddressWithdrawal, &result)
	if err != nil {
		return Withdrawal{}, err
	}
	// POST coinbasepro response is partial; retrieve full representation
	return c.GetWithdrawal(ctx, result.ID)
}

// GetWithdrawalFeeEstimate retrieves the estimated network fees that would apply when sending to the given address.
func (c *Client) GetWithdrawalFeeEstimate(ctx context.Context, cryptoAddress CryptoAddress) (WithdrawalFeeEstimate, error) {
	var withdrawalFeeEstimate WithdrawalFeeEstimate
	return withdrawalFeeEstimate, c.api.Get(ctx, fmt.Sprintf("/withdrawals/fee-estimate/%s", query(cryptoAddress.Params())), &withdrawalFeeEstimate)
}

// CreateStablecoinConversion creates a conversion from a crypto Currency a stablecoin Currency.
func (c *Client) CreateStablecoinConversion(ctx context.Context, stablecoinConversionSpec StablecoinConversionSpec) (StablecoinConversion, error) {
	var result StablecoinConversion
	return result, c.api.Post(ctx, "/conversions/", stablecoinConversionSpec, &result)
}

// ListPaymentMethods retrieves the list of PaymentMethods available for the current Profile. The list is not paginated.
func (c *Client) ListPaymentMethods(ctx context.Context) ([]PaymentMethod, error) {
	var paymentMethods []PaymentMethod
	return paymentMethods, c.api.Get(ctx, "/payment-methods/", &paymentMethods)
}

// ListCoinbaseAccounts retrieves the list of CoinbaseAccounts available for the current Profile. The list is not paginated.
func (c *Client) ListCoinbaseAccounts(ctx context.Context) ([]CoinbaseAccount, error) {
	var coinbaseAccounts []CoinbaseAccount
	return coinbaseAccounts, c.api.Get(ctx, "/coinbase-accounts/", &coinbaseAccounts)
}

// GetFees returns current maker & taker fee rates, as well as the 30-day trailing volume. GetFees is plural, but returns
// a single object. Perhaps there is a better name.
func (c *Client) GetFees(ctx context.Context) (Fees, error) {
	var fees Fees
	return fees, c.api.Get(ctx, "/fees/", &fees)
}

// CreateReport creates request for batches of historic Profile information in various human and machine readable forms.
// Reports will be generated when resources are available. Report status can be queried using GetReport.
func (c *Client) CreateReport(ctx context.Context, createReportSpec ReportSpec) (Report, error) {
	result := struct {
		ID string `json:"id"`
	}{}
	err := c.api.Post(ctx, "/reports/", createReportSpec, &result)
	if err != nil {
		return Report{}, err
	}
	// POST coinbasepro response is partial; retrieve full representation
	return c.GetReport(ctx, result.ID)
}

// GetReport retrieves the status of the processing of a Report request. When the ReportStatus is 'ready',
// the Report will be available for download at the FileURL.
func (c *Client) GetReport(ctx context.Context, reportID string) (Report, error) {
	var report Report
	return report, c.api.Get(ctx, fmt.Sprintf("/reports/%s", reportID), &report)
}

// ListProfiles retrieves a list of Profiles (portfolio equivalents). A given user can have a maximum of 10 profiles.
// The list is not paginated.
func (c *Client) ListProfiles(ctx context.Context, filter ProfileFilter) ([]Profile, error) {
	var profiles []Profile
	return profiles, c.api.Get(ctx, fmt.Sprintf("/profiles/%s", query(filter.Params())), &profiles)
}

// GetProfile retrieves the details of a single Profile.
func (c *Client) GetProfile(ctx context.Context, profileID string) (Profile, error) {
	var profile Profile
	return profile, c.api.Get(ctx, fmt.Sprintf("/profiles/%s", profileID), &profile)
}

// CreateProfileTransfer transfers funds between user Profiles.
func (c *Client) CreateProfileTransfer(ctx context.Context, transferSpec ProfileTransferSpec) (ProfileTransfer, error) {
	var transfer ProfileTransfer
	return transfer, c.api.Post(ctx, "/profiles/transfer", transferSpec, &transfer)
}

// Market Data

// ListProducts retrieves the list Currency pairs available for trading. The list is not paginated.
func (c *Client) ListProducts(ctx context.Context) ([]Product, error) {
	var products []Product
	return products, c.api.Get(ctx, "/products/", &products)
}

// GetProduct retrieves the details of a single Currency pair.
func (c *Client) GetProduct(ctx context.Context, productID ProductID) (Product, error) {
	var product Product
	return product, c.api.Get(ctx, fmt.Sprintf("/products/%s", productID), &product)
}

// GetAggregatedOrderBook retrieves an aggregated, BookLevelBest (1) and BookLevelTop50 (2), representation of a Product
// OrderBook. Aggregated levels return only one Size for each active Price (as if there was only a single Order for that Size at the level).
func (c *Client) GetAggregatedOrderBook(ctx context.Context, productID ProductID, level BookLevel) (AggregatedOrderBook, error) {
	var aggregatedBook AggregatedOrderBook
	return aggregatedBook, c.api.Get(ctx, fmt.Sprintf("/products/%s/book/%s", productID, query(level.Params())), &aggregatedBook)
}

// GetOrderBook retrieves the full, un-aggregated OrderBook for a Product.
func (c *Client) GetOrderBook(ctx context.Context, productID ProductID) (OrderBook, error) {
	var book OrderBook
	return book, c.api.Get(ctx, fmt.Sprintf("/products/%s/book/?level=3", productID), &book)
}

// GetProductTicker retrieves snapshot information about the last trade (tick), best bid/ask and 24h volume of a Product.
func (c *Client) GetProductTicker(ctx context.Context, productID ProductID) (ProductTicker, error) {
	var ticker ProductTicker
	return ticker, c.api.Get(ctx, fmt.Sprintf("/products/%s/ticker", productID), &ticker)
}

// GetProductTrades retrieves a paginated list of the last trades of a Product.
func (c *Client) GetProductTrades(ctx context.Context, productID ProductID, pagination PaginationParams) (ProductTrades, error) {
	var trades ProductTrades
	return trades, c.api.Get(ctx, fmt.Sprintf("/products/%s/trades/%s", productID, query(pagination.Params())), &trades)
}

// GetHistoricRates retrieves historic rates, as Candles, for a Product. Rates grouped buckets based on requested Granularity.
// If either one of the Start or End fields are not provided then both fields will be ignored.
// The Granularity is limited to a set of supported Timeslices, one of:
//   one minute, five minutes, fifteen minutes, one hour, six hours, or one day.
func (c *Client) GetHistoricRates(ctx context.Context, productID ProductID, filter HistoricRateFilter) (HistoricRates, error) {
	var history HistoricRates
	return history, c.api.Get(ctx, fmt.Sprintf("/products/%s/candles/%s", productID, query(filter.Params())), &history)
}

// GetProductStats retrieves the 24hr stats for a Product. Volume is in base Currency units. Open, High, and Low are in quote Currency units.
func (c *Client) GetProductStats(ctx context.Context, productID ProductID) (ProductStats, error) {
	var stats ProductStats
	return stats, c.api.Get(ctx, fmt.Sprintf("/products/%s/stats", productID), &stats)
}

// ListCurrencies retrieves the list of known Currencies. Not all Currencies may be available for trading.
func (c *Client) ListCurrencies(ctx context.Context) ([]Currency, error) {
	var currencies []Currency
	return currencies, c.api.Get(ctx, "/currencies/", &currencies)
}

// GetCurrency retrieves the details of a specific Currency.
func (c *Client) GetCurrency(ctx context.Context, currencyName CurrencyName) (Currency, error) {
	var currency Currency
	return currency, c.api.Get(ctx, fmt.Sprintf("/currencies/%s", currencyName), &currency)
}

// GetServerTime retrieves the Coinbase Pro API server time.
func (c *Client) GetServerTime(ctx context.Context) (ServerTime, error) {
	var serverTime ServerTime
	return serverTime, c.api.Get(ctx, "/time", &serverTime)
}

// Watch provides a feed of real-time market data updates for orders and trades.
func (c *Client) Watch(ctx context.Context, subscriptionRequest SubscriptionRequest, feed Feed) (capture error) {
	wsConn, err := c.dialer.Dial()
	if err != nil {
		return err
	}
	// subscription request must be sent within 5 seconds of open or socket will auto-close
	err = wsConn.WriteJSON(subscriptionRequest)
	if err != nil {
		return err
	}
	return c.watch(ctx, wsConn, feed)
}

type websocketFeedDialer struct {
	FeedURL string
}

// Dial returns a connection to the FeedURL websocket.
func (w websocketFeedDialer) Dial() (*websocket.Conn, error) {
	var wsDialer websocket.Dialer
	wsConn, resp, err := wsDialer.Dial(w.FeedURL, nil)
	if err != nil {
		return nil, err
	}
	_, _ = ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return wsConn, nil
}

type jsonReader interface {
	ReadJSON(v interface{}) error
}

func (c *Client) watch(ctx context.Context, r jsonReader, feed Feed) (capture error) {
	messages := make(chan interface{})
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		defer close(messages)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
				// TODO: Does this prevent blocking?
			case messages <- func() interface{} {
				logrus.Debug("receive message on socket")
				// TODO: Does message have a real structure
				var message interface{}
				err := r.ReadJSON(&message)
				if err != nil {
					return err
				}
				return message
			}():
			}
		}
	})

	wg.Go(func() error {
		for message := range messages {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case feed.Messages <- message:
				logrus.Debug("publish message on channel")
			default:
			}
		}
		return nil
	})
	return wg.Wait()
}

func (c *Client) Close() error {
	return c.api.Close()
}

// NewClient creates a high-level Coinbase Pro API client.
func NewClient(baseURL *url.URL, feedURL *url.URL, auth *Auth) (*Client, error) {
	apiClient, err := NewAPIClient(baseURL, feedURL, auth)
	if err != nil {
		return nil, err
	}
	return &Client{
		api: apiClient,
		dialer: websocketFeedDialer{
			FeedURL: feedURL.String(),
		},
	}, nil
}

func DevelopmentMode(client *Client) {
	devClient, err := NewDevelopmentClient(client.api.(*APIClient), afero.NewOsFs())
	if err != nil {
		panic(err)
	}
	client.api = devClient
}

type apier interface {
	Get(ctx context.Context, relativePath string, result interface{}) error
	Do(ctx context.Context, method string, relativePath string, content interface{}, result interface{}) (capture error)
	Post(ctx context.Context, relativePath string, content interface{}, result interface{}) error
	Close() error
}

type dialer interface {
	Dial() (*websocket.Conn, error)
}

type Client struct {
	api    apier
	dialer dialer
}

func query(params []string) string {
	if len(params) == 0 {
		return ""
	}
	return "?" + strings.Join(params, "&")
}

func NewAPIClient(baseURL *url.URL, feedURL *url.URL, auth *Auth) (*APIClient, error) {
	apiClient := APIClient{
		auth:       auth,
		baseURL:    baseURL,
		feedURL:    feedURL,
		httpClient: http.DefaultClient,
		timestamp: func() string {
			return strconv.FormatInt(time.Now().Unix(), 10)
		},
	}
	return &apiClient, nil
}

type APIClient struct {
	auth       *Auth
	baseURL    *url.URL
	feedURL    *url.URL
	httpClient *http.Client
	timestamp  func() string
}

func (a *APIClient) Get(ctx context.Context, relativePath string, result interface{}) error {
	return a.Do(ctx, "GET", relativePath, nil, result)
}

func (a *APIClient) Do(ctx context.Context, method string, relativePath string, content interface{}, result interface{}) (capture error) {
	resp, err := a.do(ctx, method, relativePath, content, result)
	if err != nil {
		return err
	}
	if isPaged(resp) {
		err = paginate(resp, result)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *APIClient) do(ctx context.Context, method string, relativePath string, content interface{}, result interface{}) (resp *http.Response, capture error) {
	uri, err := a.baseURL.Parse(relativePath)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("%s %s", method, relativePath)
	var b bytes.Buffer
	if content != nil {
		err = json.NewEncoder(&b).Encode(content)
		if err != nil {
			return nil, err
		}
	}
	timestamp := a.timestamp()
	msg := fmt.Sprintf("%s%s%s%s", timestamp, method, relativePath, b.Bytes())
	signature, err := a.auth.Sign(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), &b)
	if err != nil {
		return nil, err
	}
	a.addHeaders(req, timestamp, signature)
	resp, err = a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		coinbaseErr := Error{StatusCode: resp.StatusCode}
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&coinbaseErr); err != nil {
			return nil, err
		}
		return nil, coinbaseErr
	}
	defer func() { Capture(&capture, resp.Body.Close()) }()
	if result != nil {
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(result); err != nil {
			return nil, err
		}
	}
	return resp, err
}

func isPaged(resp *http.Response) bool {
	return resp.Header.Get("CB-BEFORE") != "" && resp.Header.Get("CB-AFTER") != ""
}

func paginate(resp *http.Response, result interface{}) error {
	paginated := struct {
		Page *Pagination
	}{
		&Pagination{
			Before: resp.Header.Get("CB-BEFORE"),
			After:  resp.Header.Get("CB-AFTER"),
		},
	}
	if _, ok := result.(*json.RawMessage); ok {
		// pagination is never present in raw result, just skip
		return nil
	}
	return mapstructure.Decode(paginated, result)
}

func (a *APIClient) addHeaders(req *http.Request, timestamp string, signature string) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Golang Reticule v0.1")
	req.Header.Add("CB-ACCESS-KEY", a.auth.Key)
	req.Header.Add("CB-ACCESS-PASSPHRASE", a.auth.Passphrase)
	req.Header.Add("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("CB-ACCESS-SIGN", signature)
}

func (a *APIClient) Post(ctx context.Context, relativePath string, content interface{}, result interface{}) error {
	return a.Do(ctx, "POST", relativePath, content, result)
}

func (a *APIClient) Close() error { return nil }

func NewDevelopmentClient(client *APIClient, fs afero.Fs) (d *DevelopmentClient, capture error) {
	file, err := fs.OpenFile("store.json", os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer func() { Capture(&capture, file.Close()) }()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return &DevelopmentClient{
			api:   client,
			store: phizog.NewStore(),
			fs:    fs,
		}, nil
	}
	var raw map[string]phizog.Occurrence
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return nil, err
	}
	var store phizog.Store
	err = store.Load(raw)
	if err != nil {
		return nil, err
	}
	return &DevelopmentClient{
		api:   client,
		store: &store,
		fs:    fs,
	}, nil
}

func (d *DevelopmentClient) Close() (capture error) {
	file, err := d.fs.OpenFile("store.json", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { capture = file.Close() }()
	return d.store.Write(file)
}

type DevelopmentClient struct {
	api   *APIClient
	store *phizog.Store
	fs    afero.Fs
}

func (d *DevelopmentClient) Get(ctx context.Context, relativePath string, result interface{}) error {
	return d.Do(ctx, "GET", relativePath, nil, result)
}

func (d *DevelopmentClient) Post(ctx context.Context, relativePath string, content interface{}, result interface{}) error {
	return d.Do(ctx, "POST", relativePath, content, result)
}
func (d *DevelopmentClient) Do(ctx context.Context, method string, relativePath string, content interface{}, result interface{}) (capture error) {
	logrus.Debugf("DeveloperMode %s %s", method, relativePath)
	var rawMessage json.RawMessage
	resp, err := d.api.do(ctx, method, relativePath, content, &rawMessage)
	if err != nil {
		return err
	}
	// unmarshal in order to normalize for shape store
	var raw interface{}
	err = json.Unmarshal(rawMessage, &raw)
	if err != nil {
		panic(err)
	}
	basePath := path.Dir(relativePath)
	err = d.store.AddShape(basePath, raw)
	if err != nil {
		return err
	}
	// unmarshal in order to provide "real" result
	err = json.Unmarshal(rawMessage, &result)
	if err != nil {
		panic(err)
	}
	if isPaged(resp) {
		err = paginate(resp, &result)
		if err != nil {
			return err
		}
	}
	b, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	// unmarshal in order to normalize for shape store and detect any extra or unmapped fields
	var rawResult interface{}
	err = json.Unmarshal(b, &rawResult)
	if err != nil {
		panic(err)
	}
	typed := fmt.Sprintf("%s (%T)", basePath, result)
	err = d.store.AddShape(typed, rawResult)
	if err != nil {
		return err
	}

	return nil
}
