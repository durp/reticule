package commands

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/durp/reticule/pkg/coinbasepro"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type coinbaseCmd struct {
	Config          string    `kong:"name='config',short='f',type='path',default='~/.reticule/coinbasepro'"`
	Cancel          cancelCmd `kong:"cmd,name='cancel',help='cancel an order'"`
	Create          createCmd `kong:"cmd,name='create',help='create resources including deposits, orders, and withdrawals'"`
	Get             getCmd    `kong:"cmd,name='get',help='retrieve resource representations'"`
	Watch           watchCmd  `kong:"cmd,name='watch',help='watch the websocket feed'"`
	DevelopmentMode bool      `kong:"name='dev-mode',short='D',help='dev-mode collects API response shapes for inspection and comparison'"`

	Output
}

var _ coinbaser = (*coinbasepro.Client)(nil)

// AfterApply attempts to read the coinbase config and load the
// BaseURL and Auth required to interact with the coinbasepro API.
// If the config can be loaded, it creates the coinbase.Client and binds
// it into the kong.Context for use by other commands.
func (c *coinbaseCmd) AfterApply(ktx *kong.Context) error {
	_, err := os.Stat(c.Config)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no config found at %q: use 'coinbasepro create config' to create a config file", c.Config)
		// we are in "coinbasepro create config" command; we don't need a client and will create the config
	}
	if err != nil {
		return err
	}
	source, err := ioutil.ReadFile(c.Config)
	if err != nil {
		return err
	}
	var configSet coinbaseProConfigSet
	err = yaml.Unmarshal(source, &configSet)
	if err != nil {
		return err
	}
	if len(configSet.Configs) == 0 {
		return errors.New("no config exists, use the `config create` command to create a config")
	}
	current := configSet.Current
	if current == "" {
		return fmt.Errorf("no current config is set, use the `config use` command to set the config to use")
	}
	cfg, ok := configSet.Configs[current]
	if !ok {
		return fmt.Errorf("no config with name %q, use the `config use` command to set the config to use", current)
	}
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return err
	}
	feedURL, err := url.Parse(cfg.FeedURL)
	if err != nil {
		return err
	}
	client, err := coinbasepro.NewClient(baseURL, feedURL, cfg.Auth)
	if err != nil {
		return err
	}
	if c.DevelopmentMode {
		// DevelopmentMode is a nod to the fact that the Coinbase Pro API has some endpoints that either are
		// subject to change or return responses with schemaless maps. My hope is that DevelopmentMode makes
		// it easier to identify changes in the shape of data.
		coinbasepro.DevelopmentMode(client)
	}
	ktx.BindTo(client, (*coinbaser)(nil))
	return nil
}

// Run of the coinbaseCmd is a good place to tuck cleanup
// as it is called after any and all leaf commands.
func (c *coinbaseCmd) Run(cb coinbaser) error {
	return cb.Close()
}

type cancelCmd struct {
	Order cancelOrderCmd `kong:"cmd,name='order',help='cancel an order'"`
}

type createCmd struct {
	Order           createOrderCmd             `kong:"cmd,name='order',help='create an order'"`
	Conversion      createStablecoinConversion `kong:"cmd,name='conversion',help='convert crypto to stablecoin'"`
	Deposit         createDepositCmd           `kong:"cmd,name='deposit',help='create a deposit'"`
	DepositAddress  createDepositAddressCmd    `kong:"cmd,name='deposit-address',help='create a crypto deposit address'"`
	ProfileTransfer createProfileTransferCmd   `kong:"cmd,name='profile-transfer',help='transfer between profiles'"`
	Report          createReportCmd            `kong:"cmd,name='report',help='create a report of account or fill activity'"`
	Withdrawal      createWithdrawalCmd        `kong:"cmd,name='withdrawal',help='create a withdrawal of funds'"`
}

type getCmd struct {
	Accounts         accountsCmd         `kong:"cmd,name='accounts',help='get accounts and account details'"`
	CoinbaseAccounts coinbaseAccountsCmd `kong:"cmd,name='coinbase-accounts',help='get coinbase accounts and details'"`
	Currencies       currencyCmd         `kong:"cmd,name='currencies',help='get currencies and currency details'"`
	Deposits         depositsCmd         `kong:"cmd,name='deposits',help='get deposits and deposit details'"`
	Fees             feesCmd             `kong:"cmd,name='fees',help='get fees and fee details'"`
	Fills            fillsCmd            `kong:"cmd,name='fills',help='get fills and fill details'"`
	Holds            holdsCmd            `kong:"cmd,name='holds',help='get holds and hold details'"`
	Ledger           ledgerCmd           `kong:"cmd,name='ledger',help='get ledger of transactions and transaction details'"`
	Limits           limitsCmd           `kong:"cmd,name='limits',help='get limits and limit details'"`
	Orders           ordersCmd           `kong:"cmd,name='orders',help='get orders and order details'"`
	PaymentMethods   paymentMethodsCmd   `kong:"cmd,name='payment-methods',help='get payment methods and payment method details'"`
	Products         productsCmd         `kong:"cmd,name='products',help='get products and product details'"`
	ProductOrderBook productOrderBookCmd `kong:"cmd,name='product-book',help='get order book for a product'"`
	ProductHistory   historicRatesCmd    `kong:"cmd,name='product-history',help='get history for a product'"`
	ProductStats     productStatsCmd     `kong:"cmd,name='product-stats',help='get stats for a product'"`
	ProductTicker    productTickerCmd    `kong:"cmd,name='product-ticker',help='get current ticker for a product'"`
	ProductTrades    productTradesCmd    `kong:"cmd,name='product-trades',help='get trades for a product'"`
	Profiles         profilesCmd         `kong:"cmd,name='profiles',help='get profiles and profile details'"`
	Report           reportCmd           `kong:"cmd,name='report',help='get status of a report'"`
	Time             serverTimeCmd       `kong:"cmd,name='server-time',help='get current server time'"`
	Withdrawals      withdrawalsCmd      `kong:"cmd,name='withdrawals',help='get withdrawals and withdrawal details'"`
	WithdrawalFee    withdrawalFeeCmd    `kong:"cmd,name='withdrawal-fee',help='get estimated fee for a withdrawal'"`
}

type coinbaser interface {
	ListAccounts(ctx context.Context) ([]coinbasepro.Account, error)
	GetAccount(ctx context.Context, id string) (coinbasepro.Account, error)
	GetLedger(ctx context.Context, accountID string, pagination coinbasepro.PaginationParams) (coinbasepro.Ledger, error)
	GetHolds(ctx context.Context, accountID string, pagination coinbasepro.PaginationParams) (coinbasepro.Holds, error)

	CreateLimitOrder(ctx context.Context, limitOrder coinbasepro.LimitOrder) (coinbasepro.Order, error)
	CreateMarketOrder(ctx context.Context, marketOrder coinbasepro.MarketOrder) (coinbasepro.Order, error)
	CancelOrder(ctx context.Context, spec coinbasepro.CancelOrderSpec) (map[string]interface{}, error)
	GetOrder(ctx context.Context, orderID string) (coinbasepro.Order, error)
	GetClientOrder(ctx context.Context, clientID string) (coinbasepro.Order, error)
	GetOrders(ctx context.Context, filter coinbasepro.OrderFilter, pagination coinbasepro.PaginationParams) (coinbasepro.Orders, error)

	GetFills(ctx context.Context, filter coinbasepro.FillFilter, pagination coinbasepro.PaginationParams) (coinbasepro.Fills, error)

	GetLimits(ctx context.Context) (coinbasepro.Limits, error)

	GetDeposits(ctx context.Context, filter coinbasepro.DepositFilter, pagination coinbasepro.PaginationParams) (coinbasepro.Deposits, error)
	GetDeposit(ctx context.Context, depositID string) (coinbasepro.Deposit, error)
	CreatePaymentMethodDeposit(ctx context.Context, depositPaymentMethod coinbasepro.PaymentMethodDepositSpec) (coinbasepro.Deposit, error)
	CreateCoinbaseAccountDeposit(ctx context.Context, coinbaseAccountDeposit coinbasepro.CoinbaseAccountDeposit) (coinbasepro.Deposit, error)
	CreateCryptoDepositAddress(ctx context.Context, coinbaseAccountID string) (coinbasepro.CryptoDepositAddress, error)

	GetWithdrawals(ctx context.Context, filter coinbasepro.WithdrawalFilter, pagination coinbasepro.PaginationParams) (coinbasepro.Withdrawals, error)
	GetWithdrawal(ctx context.Context, withdrawalID string) (coinbasepro.Withdrawal, error)
	CreatePaymentMethodWithdrawal(ctx context.Context, paymentMethodWithdrawal coinbasepro.PaymentMethodWithdrawalSpec) (coinbasepro.Withdrawal, error)
	CreateCoinbaseAccountWithdrawal(ctx context.Context, coinbaseAccountWithdrawal coinbasepro.CoinbaseAccountWithdrawalSpec) (coinbasepro.Withdrawal, error)
	CreateCryptoAddressWithdrawal(ctx context.Context, cryptoAddressWithdrawal coinbasepro.CryptoAddressWithdrawalSpec) (coinbasepro.Withdrawal, error)
	GetWithdrawalFeeEstimate(ctx context.Context, cryptoAddress coinbasepro.CryptoAddress) (coinbasepro.WithdrawalFeeEstimate, error)
	CreateStablecoinConversion(ctx context.Context, stablecoinConversionSpec coinbasepro.StablecoinConversionSpec) (coinbasepro.StablecoinConversion, error)
	ListPaymentMethods(ctx context.Context) ([]coinbasepro.PaymentMethod, error)
	ListCoinbaseAccounts(ctx context.Context) ([]coinbasepro.CoinbaseAccount, error)
	GetFees(ctx context.Context) (coinbasepro.Fees, error)

	CreateReport(ctx context.Context, createReportSpec coinbasepro.ReportSpec) (coinbasepro.Report, error)
	GetReport(ctx context.Context, reportID string) (coinbasepro.Report, error)

	ListProfiles(ctx context.Context, filter coinbasepro.ProfileFilter) ([]coinbasepro.Profile, error)
	GetProfile(ctx context.Context, profileID string) (coinbasepro.Profile, error)
	CreateProfileTransfer(ctx context.Context, transferSpec coinbasepro.ProfileTransferSpec) (coinbasepro.ProfileTransfer, error)

	// Market Data

	ListProducts(ctx context.Context) ([]coinbasepro.Product, error)
	GetProduct(ctx context.Context, productID coinbasepro.ProductID) (coinbasepro.Product, error)
	GetAggregatedOrderBook(ctx context.Context, productID coinbasepro.ProductID, level coinbasepro.BookLevel) (coinbasepro.AggregatedOrderBook, error)
	GetOrderBook(ctx context.Context, productID coinbasepro.ProductID) (coinbasepro.OrderBook, error)
	GetProductTicker(ctx context.Context, productID coinbasepro.ProductID) (coinbasepro.ProductTicker, error)
	GetProductTrades(ctx context.Context, productID coinbasepro.ProductID, pagination coinbasepro.PaginationParams) (coinbasepro.ProductTrades, error)
	GetHistoricRates(ctx context.Context, productID coinbasepro.ProductID, params coinbasepro.HistoricRateFilter) (coinbasepro.HistoricRates, error)
	GetProductStats(ctx context.Context, productID coinbasepro.ProductID) (coinbasepro.ProductStats, error)

	ListCurrencies(ctx context.Context) ([]coinbasepro.Currency, error)
	GetCurrency(ctx context.Context, currencyName coinbasepro.CurrencyName) (coinbasepro.Currency, error)

	GetServerTime(ctx context.Context) (coinbasepro.ServerTime, error)

	// Watch websocket feed
	Watch(ctx context.Context, subscriptionRequest coinbasepro.SubscriptionRequest, feed coinbasepro.Feed) (capture error)

	Close() error
}

type accountsCmd struct {
	Account string `kong:"short='a',help='id of account to retrieve'"`
}

func (a *accountsCmd) Run(ctx context.Context, cb coinbaser, enc encoder) error {
	if a.Account != "" {
		account, err := cb.GetAccount(ctx, a.Account)
		if err != nil {
			return err
		}
		return enc.Encode(account)
	}
	accounts, err := cb.ListAccounts(ctx)
	if err != nil {
		return err
	}
	return enc.Encode(accounts)
}

type ledgerCmd struct {
	Account string `kong:"name='account',short='a',help='id of account to retrieve',required"`
	Pagination
}

func (l *ledgerCmd) Run(ctx context.Context, cb coinbaser, enc encoder) error {
	params, err := l.Pagination.Params()
	if err != nil {
		return err
	}
	ledger, err := cb.GetLedger(ctx, l.Account, params)
	if err != nil {
		return err
	}
	return enc.Encode(ledger)
}

type holdsCmd struct {
	Account string `kong:"name:'account',short='a',help='id of account to retrieve',required"`
	Pagination
}

func (h *holdsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	params, err := h.Pagination.Params()
	if err != nil {
		return err
	}
	ledger, err := client.GetHolds(ctx, h.Account, params)
	if err != nil {
		return err
	}
	return enc.Encode(ledger)
}

// --- Orders ---
type ordersCmd struct {
	// ClientOrderID is the client-supplied UUID of the Order
	ClientOrderID string `kong:"name='client-oid',short='c',help='client order id to retrieve'"`
	// OrderID is the server-generated UUID for the Order
	OrderID string `kong:"name='order-id',short='o',help='order id to retrieve'"`

	// List filters
	// ProductID limits the list of Orders to those with the specified ProductID
	ProductID coinbasepro.ProductID `kong:"name='product-id',short='p',help='filter by product id'"`
	// Status limits list of Orders to the provided Statuses. The default, `all`, returns orders of all statuses
	Status []coinbasepro.OrderStatusParam `kong:"name='status',short='s',help='filter by status(es)'"`
	Pagination
}

func (o *ordersCmd) Validate() error {
	if o.ClientOrderID != "" && o.OrderID != "" {
		return errors.New("only one of 'client-oid' or 'order-id' can be provided")
	}
	if o.ClientOrderID != "" && (o.ProductID != "" || len(o.Status) > 0 || !o.Pagination.Empty()) {
		return errors.New("when 'client-oid' is provided, it must be the only flag")
	}
	if o.OrderID != "" && (o.ProductID != "" || len(o.Status) > 0 || !o.Pagination.Empty()) {
		return errors.New("when 'order-id' us provided, it must be the only flag")
	}
	return nil
}

func (o *ordersCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	if o.OrderID != "" {
		order, err := client.GetOrder(ctx, o.OrderID)
		if err != nil {
			return err
		}
		return enc.Encode(order)
	}
	if o.ClientOrderID != "" {
		order, err := client.GetClientOrder(ctx, o.ClientOrderID)
		if err != nil {
			return err
		}
		return enc.Encode(order)
	}
	pagination, err := o.Pagination.Params()
	if err != nil {
		return err
	}
	orders, err := client.GetOrders(ctx, o.Filter(), pagination)
	if err != nil {
		return err
	}
	return enc.Encode(orders)
}

func (o *ordersCmd) Filter() coinbasepro.OrderFilter {
	return coinbasepro.OrderFilter{
		ProductID: o.ProductID,
		Status:    o.Status,
	}
}

type createOrderCmd struct {
	Limit  limitOrderCmd  `kong:"cmd,name='limit',default:'1',help='create a limit order (default)'"`
	Market marketOrderCmd `kong:"cmd,name='market',help='create a market order'"`
}

type limitOrderCmd struct {
	Order coinbasepro.LimitOrder `kong:"name='order',short='o',help='json {\"size\": \"0.01\",\"price\": \"0.100\",\"side\": \"buy\",\"product_id\": \"BTC-USD\"}',required"`
}

func (l *limitOrderCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	order, err := client.CreateLimitOrder(ctx, l.Order)
	if err != nil {
		return err
	}
	return enc.Encode(order)
}

type marketOrderCmd struct {
	Order coinbasepro.MarketOrder `kong:"name='order',short='o',help='json {\"size\": \"0.01\",\"side\": \"buy\",\"product_id\": \"BTC-USD\"}',required"`
}

func (m *marketOrderCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	order, err := client.CreateMarketOrder(ctx, m.Order)
	if err != nil {
		return err
	}
	return enc.Encode(order)
}

type cancelOrderCmd struct {
	OrderID       string                `kong:"name='order-id',short='o',help='id of order to cancel'"`
	ClientOrderID string                `kong:"name='client-oid',short='c',help='client order id to cancel'"`
	ProductID     coinbasepro.ProductID `kong:"name='product-id',short='p',help='product id to cancel (optional, but recommended)'"`
}

func (c *cancelOrderCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	spec := coinbasepro.CancelOrderSpec{
		OrderID:       c.OrderID,
		ClientOrderID: c.ClientOrderID,
		ProductID:     c.ProductID,
	}
	if err := spec.Validate(); err != nil {
		return err
	}
	resp, err := client.CancelOrder(ctx, spec)
	if err != nil {
		return err
	}
	return enc.Encode(resp)
}

type productsCmd struct {
	ProductID coinbasepro.ProductID `kong:"name='product-id',short='p',help='id of product to retrieve'"`
}

func (p *productsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	if p.ProductID != "" {
		product, err := client.GetProduct(ctx, p.ProductID)
		if err != nil {
			return err
		}
		return enc.Encode(product)
	}
	products, err := client.ListProducts(ctx)
	if err != nil {
		return err
	}
	return enc.Encode(products)
}

type fillsCmd struct {
	// OrderID limits the list of Fills to those with the specified OrderID
	OrderID string `kong:"name='order-id',short='o',help='filter retrieval by order id'"`
	// ProductID limits the list of Fills to those with the specified ProductID
	ProductID coinbasepro.ProductID `kong:"name='product-id',short='p',help='filter retrieval by product id'"`
	Pagination
}

func (f *fillsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	pagination, err := f.Pagination.Params()
	if err != nil {
		return err
	}
	fills, err := client.GetFills(ctx, f.Filter(), pagination)
	if err != nil {
		return err
	}
	return enc.Encode(fills)
}

func (f *fillsCmd) Filter() coinbasepro.FillFilter {
	return coinbasepro.FillFilter{
		OrderID:   f.OrderID,
		ProductID: f.ProductID,
	}
}

type limitsCmd struct{}

func (l *limitsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	limits, err := client.GetLimits(ctx)
	if err != nil {
		return err
	}
	return enc.Encode(limits)
}

type depositsCmd struct {
	DepositID string `kong:"arg,optional,name='deposit-id',help='id of deposit to retrieve'"`

	DepositFilter
	TimestampFilterPagination
}

func (d *depositsCmd) Validate() error {
	if d.DepositID != "" && !(d.DepositFilter == (DepositFilter{}) && d.TimestampFilterPagination == (TimestampFilterPagination{})) {
		return errors.New("deposit-id cannot be combined with filters")
	}
	return nil
}

type DepositFilter struct {
	// ProfileID limits the list of Deposit to the ProfileID
	ProfileID string `kong:"name='profile',short='p',help='filter by profile id'"`
	// Type identifies the type of the Deposit (`deposit` or `internal_deposit`)
	Type coinbasepro.DepositType `kong:"name='type',short='t',enum=',deposit,internal_deposit',help='filter by deposit type, one of [deposit,internal_deposit]'"`
}

func (d DepositFilter) Filter() coinbasepro.DepositFilter {
	return coinbasepro.DepositFilter{
		Type:      d.Type,
		ProfileID: d.ProfileID,
	}
}

func (d *depositsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	if d.DepositID != "" {
		deposit, err := client.GetDeposit(ctx, d.DepositID)
		if err != nil {
			return err
		}
		return enc.Encode(deposit)
	}
	deposits, err := client.GetDeposits(ctx, d.DepositFilter.Filter(), d.TimestampFilterPagination.Params())
	if err != nil {
		return err
	}
	return enc.Encode(deposits)
}

type createDepositCmd struct {
	PaymentMethod   paymentMethodDepositCmd   `kong:"cmd,name='payment-method',help='create payment method deposit'"`
	CoinbaseAccount coinbaseAccountDepositCmd `kong:"cmd,name='coinbasepro-account',help='create coinbase account deposit'"`
}

type paymentMethodDepositCmd struct {
	Deposit coinbasepro.PaymentMethodDepositSpec `kong:"name='deposit',short='d',help='json {\"amount\":10.00,\"currency\":\"USD\",\"payment_method_id\":\"bc677162-d934-5f1a-968c-a496b1c1270b\"}',required"`
}

func (d *paymentMethodDepositCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	deposit, err := client.CreatePaymentMethodDeposit(ctx, d.Deposit)
	if err != nil {
		return err
	}
	return enc.Encode(deposit)
}

type coinbaseAccountDepositCmd struct {
	Deposit coinbasepro.CoinbaseAccountDeposit `kong:"name='deposit',short='d',help='json {\"amount\":\"1000\",\"currency\":\"USD\",\"coinbase_account_id\":\"bcdd4c40-df40-5d76-810c-74aab722b223\"}',required"`
}

func (c *coinbaseAccountDepositCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	deposit, err := client.CreateCoinbaseAccountDeposit(ctx, c.Deposit)
	if err != nil {
		return err
	}
	return enc.Encode(deposit)
}

type createDepositAddressCmd struct {
	CoinbaseAccountID string `kong:"name='coinbasepro-account-id',short='c',help='id of coinbasepro account',required"`
}

func (c *createDepositAddressCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	cryptoDepositAddress, err := client.CreateCryptoDepositAddress(ctx, c.CoinbaseAccountID)
	if err != nil {
		return err
	}
	return enc.Encode(cryptoDepositAddress)
}

type withdrawalsCmd struct {
	WithdrawalID string `kong:"arg,optional,name='withdrawal-id',help='withdrawal id to retrieve'"`

	WithdrawalFilter
	TimestampFilterPagination
}

func (w withdrawalsCmd) Validate() error {
	if w.WithdrawalID != "" && !(w.WithdrawalFilter == (WithdrawalFilter{}) && w.TimestampFilterPagination == (TimestampFilterPagination{})) {
		return errors.New("withdrawal-id cannot be combined with filters")
	}
	return nil
}

type WithdrawalFilter struct {
	// Type identifies the type of the Withdrawal (`withdraw` or `internal_withdraw`)
	Type coinbasepro.WithdrawalType `kong:"name='type',short='t',enum=',withdraw,internal_withdraw',help='filter by withdrawal type, one of [withdraw,internal_withdraw]'"`
	// ProfileID limits the list of Withdrawals to the ProfileID
	ProfileID string `kong:"name='profile',short='p',help='filter retrieval by profile id'"`
}

func (w WithdrawalFilter) Filter() coinbasepro.WithdrawalFilter {
	return coinbasepro.WithdrawalFilter{
		Type:      w.Type,
		ProfileID: w.ProfileID,
	}
}
func (w *withdrawalsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	if w.WithdrawalID != "" {
		deposit, err := client.GetWithdrawal(ctx, w.WithdrawalID)
		if err != nil {
			return err
		}
		return enc.Encode(deposit)
	}
	withdrawals, err := client.GetWithdrawals(ctx, w.WithdrawalFilter.Filter(), w.TimestampFilterPagination.Params())
	if err != nil {
		return err
	}
	return enc.Encode(withdrawals)
}

type createWithdrawalCmd struct {
	CoinbaseAccount coinbaseAccountWithdrawalCmd `kong:"cmd,name='coinbasepro-account',help='create withdrawal from coinbasepro account'"`
	CryptoAddress   cryptoAddressWithdrawalCmd   `kong:"cmd,name='crypto-address',help='create withdrawal from crypto address'"`
	PaymentMethod   paymentMethodWithdrawalCmd   `kong:"cmd,name='payment-method',help='create withdrawal from payment method'"`
}

type paymentMethodWithdrawalCmd struct {
	Withdrawal coinbasepro.PaymentMethodWithdrawalSpec `kong:"name='withdrawal',short='w',help='json {\"amount\":10.00,\"currency\":\"USD\",\"id\":\"bc677162-d934-5f1a-968c-a496b1c1270b\"}',required"`
}

func (p *paymentMethodWithdrawalCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	withdrawal, err := client.CreatePaymentMethodWithdrawal(ctx, p.Withdrawal)
	if err != nil {
		return err
	}
	return enc.Encode(withdrawal)
}

type coinbaseAccountWithdrawalCmd struct {
	Withdrawal coinbasepro.CoinbaseAccountWithdrawalSpec `kong:"name='withdrawal',short='w',help='json {\"amount\":10.00,\"currency\":\"USD\",\"coinbase_account_id\":\"bc677162-d934-5f1a-968c-a496b1c1270b\"}',required"`
}

func (c *coinbaseAccountWithdrawalCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	withdrawal, err := client.CreateCoinbaseAccountWithdrawal(ctx, c.Withdrawal)
	if err != nil {
		return err
	}
	return enc.Encode(withdrawal)
}

type cryptoAddressWithdrawalCmd struct {
	Withdrawal coinbasepro.CryptoAddressWithdrawalSpec `kong:"name='withdrawal',short='w',help='json {\"amount\":10.00,\"currency\":\"BTC\",\"crypto_address\":\"0x5ad5769cd04681FeD900BCE3DDc877B50E83d469\"}',required"`
}

func (c *cryptoAddressWithdrawalCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	withdrawal, err := client.CreateCryptoAddressWithdrawal(ctx, c.Withdrawal)
	if err != nil {
		return err
	}
	return enc.Encode(withdrawal)
}

type withdrawalFeeCmd struct {
	CryptoAddress coinbasepro.CryptoAddress `kong:"name='crypto-address',short='c',help='json {\"currency\":\"BTC\",\"crypto_address\":\"0x5ad5769cd04681FeD900BCE3DDc877B50E83d469\"}',required"`
}

func (w *withdrawalFeeCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	feeEstimate, err := client.GetWithdrawalFeeEstimate(ctx, w.CryptoAddress)
	if err != nil {
		return err
	}
	return enc.Encode(feeEstimate)
}

type createStablecoinConversion struct {
	StablecoinConversion coinbasepro.StablecoinConversionSpec `kong:"name='stablecoin',help='json: {\"from\":\"BTC\",\"to\":\"USD\",\"amount\":\"1.0\"}',required"`
}

func (c *createStablecoinConversion) Run(ctx context.Context, client coinbaser, enc encoder) error {
	conversion, err := client.CreateStablecoinConversion(ctx, c.StablecoinConversion)
	if err != nil {
		return err
	}
	return enc.Encode(conversion)
}

type paymentMethodsCmd struct{}

func (p *paymentMethodsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	paymentMethods, err := client.ListPaymentMethods(ctx)
	if err != nil {
		return err
	}
	return enc.Encode(paymentMethods)
}

type coinbaseAccountsCmd struct{}

func (c *coinbaseAccountsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	coinbaseAccounts, err := client.ListCoinbaseAccounts(ctx)
	if err != nil {
		return err
	}
	return enc.Encode(coinbaseAccounts)
}

type feesCmd struct{}

func (f *feesCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	fees, err := client.GetFees(ctx)
	if err != nil {
		return err
	}
	return enc.Encode(fees)
}

type createReportCmd struct {
	Spec coinbasepro.ReportSpec `kong:"name='spec',short='s',help='json {\"type\": \"fills\",\"start_date\": \"2014-11-01T00:00:00.000Z\",\"end_date\": \"2014-11-30T23:59:59.000Z\"}'"`
}

func (c *createReportCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	report, err := client.CreateReport(ctx, c.Spec)
	if err != nil {
		return err
	}
	return enc.Encode(report)
}

type reportCmd struct {
	ReportID string `kong:"name='report-id',short='r',help='id of report to retrieve',required"`
}

func (r *reportCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	report, err := client.GetReport(ctx, r.ReportID)
	if err != nil {
		return err
	}
	return enc.Encode(report)
}

type profilesCmd struct {
	Active    bool   `kong:"name='active',short='a',help='if set, only return active profiles'"`
	ProfileID string `kong:"name='profile-id',short='p',help='id of profile to retrieve'"`
}

func (p *profilesCmd) Validate() error {
	if p.Active && p.ProfileID != "" {
		return errors.New("only one of 'active' or 'profile-id' allowed")
	}
	return nil
}

func (p *profilesCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	if p.ProfileID != "" {
		profile, err := client.GetProfile(ctx, p.ProfileID)
		if err != nil {
			return err
		}
		return enc.Encode(profile)
	}
	profiles, err := client.ListProfiles(ctx, coinbasepro.ProfileFilter{
		Active: p.Active,
	})
	if err != nil {
		return err
	}
	return enc.Encode(profiles)
}

type createProfileTransferCmd struct {
	Transfer coinbasepro.ProfileTransferSpec `kong:"name='transfer',short='t',help='json: {\"from\":\"86602c68-306a-4500-ac73-4ce56a91d83c\",\"to\":\"e87429d3-f0a7-4f28-8dff-8dd93d383de1\",\"currency\":\"BTC\",\"amount\":\"1000.00\"}',required"`
}

func (c *createProfileTransferCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	transfer, err := client.CreateProfileTransfer(ctx, c.Transfer)
	if err != nil {
		return err
	}
	return enc.Encode(transfer)
}

type productOrderBookCmd struct {
	ProductID coinbasepro.ProductID `kong:"name='product-id',short='p',help='id of product to retrieve',required"`
	// Level
	// By default, only the inside (i.e. best) bid and ask are returned. This is equivalent to a book depth of 1 level.
	// To see a larger order book, specify the level query parameter.
	// If a level is not aggregated, then all of the orders at each price will be returned.
	// Aggregated levels return only one size for each active price (as if there was only a single order for that size at the level).
	Level coinbasepro.BookLevel `kong:"name='level',short='l',help='detail level of order book to retrieve, one of [ 1, 2, 3] corresponding to [ best, top50, full ]'"`
}

func (p *productOrderBookCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	switch p.Level {
	case coinbasepro.BookLevelUndefined, coinbasepro.BookLevelBest, coinbasepro.BookLevelTop50:
		orderBook, err := client.GetAggregatedOrderBook(ctx, p.ProductID, p.Level)
		if err != nil {
			return err
		}
		return enc.Encode(orderBook)
	case coinbasepro.BookLevelFull:
		orderBook, err := client.GetOrderBook(ctx, p.ProductID)
		if err != nil {
			return err
		}
		return enc.Encode(orderBook)
	default:
		return fmt.Errorf("level %d is not a valid level, valid levels are 1-3", p.Level)
	}
}

type productTickerCmd struct {
	ProductID coinbasepro.ProductID `kong:"name='product-id',short='p',required"`
}

func (p *productTickerCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	ticker, err := client.GetProductTicker(ctx, p.ProductID)
	if err != nil {
		return err
	}
	return enc.Encode(ticker)
}

type productTradesCmd struct {
	ProductID coinbasepro.ProductID `kong:"name='product-id',short='p',required"`
	Pagination
}

func (p *productTradesCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	params, err := p.Pagination.Params()
	if err != nil {
		return err
	}
	trades, err := client.GetProductTrades(ctx, p.ProductID, params)
	if err != nil {
		return err
	}
	return enc.Encode(trades)
}

type historicRatesCmd struct {
	HistoricRateParams

	ProductID coinbasepro.ProductID `kong:"name='product-id',short='p',required"`
}

type HistoricRateParams struct {
	Granularity coinbasepro.TimesliceParam `kong:"name='granularity',short='g',help='accepts any duration string that resolves to 1 minute (60s, 1m), 5 minutes (300s, 5m), 15 minutes (900s, 15m), 1 hour (3600s, 60m, 1h), 6 hours (21600s, 360m, 6h), or 1 day (86400s, 1440m, 24h)',required"`
	End         coinbasepro.Time           `kong:"name='end',short='e',help='end time in RFC3339 compatible format (21-04-14T12:35:00Z)'"`
	Start       coinbasepro.Time           `kong:"name='start',short='s',help='start time in RFC3339 compatible format (21-04-14T12:35:00Z)'"`
}

func (h *HistoricRateParams) Params() coinbasepro.HistoricRateFilter {
	return coinbasepro.HistoricRateFilter{
		Granularity: h.Granularity.Timeslice(),
		End:         h.End,
		Start:       h.Start,
	}
}

func (h *historicRatesCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	historicRates, err := client.GetHistoricRates(ctx, h.ProductID, h.HistoricRateParams.Params())
	if err != nil {
		return err
	}
	return enc.Encode(historicRates)
}

func (h *historicRatesCmd) Validate() error {
	if !h.Start.Time().IsZero() && !h.End.Time().IsZero() {
		return errors.New("if 'start' or 'end' time is provided, both 'start' and 'end' times must be provided")
	}
	return h.Granularity.Validate()
}

type productStatsCmd struct {
	ProductID coinbasepro.ProductID `kong:"name='product-id',short='p',required"`
}

func (p *productStatsCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	stats, err := client.GetProductStats(ctx, p.ProductID)
	if err != nil {
		return err
	}
	return enc.Encode(stats)
}

type currencyCmd struct {
	Currency coinbasepro.CurrencyName `kong:"name='currency',short='c',help='name of currency to retrieve'"`
}

func (c *currencyCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	if c.Currency != "" {
		currency, err := client.GetCurrency(ctx, c.Currency)
		if err != nil {
			return err
		}
		return enc.Encode(currency)
	}
	currencies, err := client.ListCurrencies(ctx)
	if err != nil {
		return err
	}
	return enc.Encode(currencies)
}

type serverTimeCmd struct{}

func (t *serverTimeCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	serverTime, err := client.GetServerTime(ctx)
	if err != nil {
		return err
	}
	return enc.Encode(serverTime)
}

// -- websocket feed --
type watchCmd struct {
	ProductIDs []coinbasepro.ProductID   `kong:"name='product-ids',short='p',help='product ids to add to all feeds'"`
	Channels   []coinbasepro.ChannelName `kong:"name='channels',short='c',help='specific channel name and product ids to watch'"`
	Heartbeat  []coinbasepro.ProductID   `kong:"name='heartbeat',short='b',help='watch heartbeat channel of product ids'"`
	Status     []coinbasepro.ProductID   `kong:"name='status',short='s',help='watch status channel of product ids'"`
	Ticker     []coinbasepro.ProductID   `kong:"name='ticker',short='t',help='watch ticker channel of product ids'"`
	Level2     []coinbasepro.ProductID   `kong:"name='level2',short='l',help='watch level2 channel of product ids'"`
	Full       []coinbasepro.ProductID   `kong:"name='full',short='f',help='watch full channel of product ids'"`
	User       []coinbasepro.ProductID   `kong:"name='user',short='u',help='watch user channel of product ids'"`
	Matches    []coinbasepro.ProductID   `kong:"name='matches',short='m',help='watch match channel of product ids'"`
}

func (w *watchCmd) Run(ctx context.Context, client coinbaser, enc encoder) error {
	feed := coinbasepro.NewFeed()

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return client.Watch(ctx, w.subscriptionRequest(), feed)
	})

	wg.Go(func() error {
		for message := range feed.Messages {
			logrus.Debug("receive message on channel")
			err := enc.Encode(message)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return wg.Wait()
}

func (w *watchCmd) subscriptionRequest() coinbasepro.SubscriptionRequest {
	var channels []coinbasepro.Channel
	if len(w.Heartbeat) > 0 {
		channels = append(channels, coinbasepro.Channel{
			Name:       coinbasepro.ChannelNameHeartbeat,
			ProductIDs: w.Heartbeat,
		})
	}
	if len(w.Status) > 0 {
		channels = append(channels, coinbasepro.Channel{
			Name:       coinbasepro.ChannelNameStatus,
			ProductIDs: w.Status,
		})
	}
	if len(w.Ticker) > 0 {
		channels = append(channels, coinbasepro.Channel{
			Name:       coinbasepro.ChannelNameTicker,
			ProductIDs: w.Ticker,
		})
	}
	if len(w.Level2) > 0 {
		channels = append(channels, coinbasepro.Channel{
			Name:       coinbasepro.ChannelNameLevel2,
			ProductIDs: w.Level2,
		})
	}
	if len(w.Full) > 0 {
		channels = append(channels, coinbasepro.Channel{
			Name:       coinbasepro.ChannelNameFull,
			ProductIDs: w.Full,
		})
	}
	if len(w.User) > 0 {
		channels = append(channels, coinbasepro.Channel{
			Name:       coinbasepro.ChannelNameUser,
			ProductIDs: w.User,
		})
	}
	if len(w.Matches) > 0 {
		channels = append(channels, coinbasepro.Channel{
			Name:       coinbasepro.ChannelNameMatches,
			ProductIDs: w.Matches,
		})
	}
	return coinbasepro.NewSubscriptionRequest(w.ProductIDs, w.Channels, channels)
}

type TimestampFilterPagination struct {
	// After limits the list of Deposits to those created before the after timestamp, sorted by newest
	After coinbasepro.Time `kong:"name='after',short='a',help='filter retrieval to deposits created before the after timestamp'"`
	// Before limits the list of Deposits to those created after the before timestamp, sorted by oldest creation date.
	Before coinbasepro.Time `kong:"name='before',short='b',help='filter retrieval to deposits created after the before timestamp'"`
	// Limit limits the number of Deposit returned to the Limit. Max 100. Default 100.
	Limit int `kong:"name='limit',short='l',help='limit retrieval to at most this many deposits, max 100, default 100'"`
}

func (d TimestampFilterPagination) Params() coinbasepro.PaginationParams {
	return coinbasepro.PaginationParams{
		After:  d.After.Time().Format(time.RFC3339),
		Before: d.Before.Time().Format(time.RFC3339),
		Limit:  d.Limit,
	}
}

func (d *TimestampFilterPagination) Validate() error {
	if d.Limit < 0 {
		return errors.New("limit must not be negative")
	}
	return nil
}

type Pagination struct {
	Before string `kong:"name='prev',short='p',help='retrieve the page before this token'"`
	After  string `kong:"name='next',short='n',help='retrieve the page after this token'"`
	Limit  int    `kong:"name='limit',short='l',help='limit retrieval to at most this many results, max 100, default 100'"`
}

func (p Pagination) Validate() error {
	if p.Before != "" && p.After != "" {
		return errors.New("only one of 'before' or 'after' allowed")
	}
	if p.Limit < 0 {
		return errors.New("'limit' must be a positive number")
	}
	return nil
}

func (p Pagination) Params() (coinbasepro.PaginationParams, error) {
	var paginationParams coinbasepro.PaginationParams
	return paginationParams, mapstructure.Decode(p, &paginationParams)
}

func (p Pagination) Empty() bool {
	return p.After == "" && p.Before == "" && p.Limit == 0
}
