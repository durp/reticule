package coinbasepro

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAPIClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Golang Reticule v0.1", r.Header.Get("User-Agent"))
		assert.Equal(t, "k", r.Header.Get("CB-ACCESS-KEY"))
		assert.Equal(t, "p", r.Header.Get("CB-ACCESS-PASSPHRASE"))
		assert.Equal(t, "1", r.Header.Get("CB-ACCESS-TIMESTAMP"))
		assert.Equal(t, "2A+P3UPYN+cYs7ehkTJOLK/mkIHgHk1cahG1BCGDHRs=", r.Header.Get("CB-ACCESS-SIGN"))
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()
	baseURL, err := url.Parse(ts.URL)
	require.NoError(t, err)
	c := APIClient{
		auth: &Auth{
			Key:        "k",
			Passphrase: "p",
			Secret:     "zZ==",
		},
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		timestamp: func() string {
			return "1"
		},
	}
	ctx := context.Background()
	var result interface{}
	err = c.Get(ctx, "/test", &result)
	require.NoError(t, err)

	d, err := NewDevelopmentClient(&c, afero.NewMemMapFs())
	require.NoError(t, err)
	err = d.Get(ctx, "/test", &result)
	require.NoError(t, err)
}

func TestClient_Accounts(t *testing.T) {
	t.Run("ListAccounts", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/accounts/", mock.IsType(&[]Account{})).Return(nil)
		c := Client{api: &api}
		_, err := c.ListAccounts(context.Background())
		require.NoError(t, err)
	})
	t.Run("GetAccount", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/accounts/account-id", mock.IsType(&Account{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetAccount(context.Background(), "account-id")
		require.NoError(t, err)
	})
	t.Run("GetLedger", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		pagination := PaginationParams{
			After: "after",
			Limit: 2,
		}
		api.On("Get", "/accounts/account-id/ledger/?after=after&limit=2", mock.IsType(&Ledger{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetLedger(context.Background(), "account-id", pagination)
		require.NoError(t, err)
	})
	t.Run("GetHolds", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		pagination := PaginationParams{
			After: "after",
			Limit: 2,
		}
		api.On("Get", "/accounts/account-id/holds/?after=after&limit=2", mock.IsType(&Holds{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetHolds(context.Background(), "account-id", pagination)
		require.NoError(t, err)
	})
}

func TestClient_Orders(t *testing.T) {
	t.Run("CreateLimitOrder", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		limitOrder := LimitOrder{
			Price:     decimal.NewFromFloat(1.0),
			ProductID: "BTC-USD",
			Side:      "sell",
			Size:      decimal.NewFromFloat(1.0),
			Type:      "limit",
		}
		api.On("Post", "/orders/", limitOrder, mock.IsType(&Order{})).Return(nil)
		c := Client{api: &api}
		_, err := c.CreateLimitOrder(context.Background(), limitOrder)
		require.NoError(t, err)
	})
	t.Run("CreateMarketOrder", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		size := decimal.NewFromFloat(1.0)
		marketOrder := MarketOrder{
			ProductID: "BTC-USD",
			Side:      "sell",
			Size:      &size,
			Type:      "market",
		}
		api.On("Post", "/orders/", marketOrder, mock.IsType(&Order{})).Return(nil)
		c := Client{api: &api}
		_, err := c.CreateMarketOrder(context.Background(), marketOrder)
		require.NoError(t, err)
	})
	t.Run("CancelOrder", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		spec := CancelOrderSpec{
			OrderID:   "order-id",
			ProductID: "BTC-USD",
		}
		api.On("Do", "DELETE", "/orders/order-id?product_id=BTC-USD", nil, mock.IsType(&map[string]interface{}{})).Return(nil)
		c := Client{api: &api}
		_, err := c.CancelOrder(context.Background(), spec)
		require.NoError(t, err)
	})
	t.Run("GetOrders", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		filter := OrderFilter{
			ProductID: "BTC-USD",
			Status:    []OrderStatusParam{"active"},
		}
		pagination := PaginationParams{
			After: "after",
			Limit: 2,
		}
		api.On("Get", "/orders/?product_id=BTC-USD&status=active&after=after&limit=2", mock.IsType(&Orders{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetOrders(context.Background(), filter, pagination)
		require.NoError(t, err)
	})
	t.Run("GetOrder", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/orders/order-id", mock.IsType(&Order{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetOrder(context.Background(), "order-id")
		require.NoError(t, err)
	})
	t.Run("GetClientOrder", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/orders/client:client-oid", mock.IsType(&Order{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetClientOrder(context.Background(), "client-oid")
		require.NoError(t, err)
	})
}

func TestClient_Fills(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	filter := FillFilter{
		ProductID: "BTC-USD",
	}
	pagination := PaginationParams{
		After: "after",
		Limit: 2,
	}
	api.On("Get", "/fills/?product_id=BTC-USD&after=after&limit=2", mock.IsType(&Fills{})).Return(nil)
	c := Client{api: &api}
	_, err := c.GetFills(context.Background(), filter, pagination)
	require.NoError(t, err)
}

func TestClient_Limits(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	api.On("Get", "/users/self/exchange-limits/", mock.IsType(&Limits{})).Return(nil)
	c := Client{api: &api}
	_, err := c.GetLimits(context.Background())
	require.NoError(t, err)
}

func TestClient_Deposits(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		t.Run("GetDeposits", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			filter := DepositFilter{
				ProfileID: "profile_id",
				Type:      "deposit",
			}
			pagination := PaginationParams{
				After: "after",
				Limit: 2,
			}
			api.On("Get", "/transfers/?profile_id=profile_id&type=deposit&after=after&limit=2", mock.IsType(&Deposits{})).Return(nil)
			c := Client{api: &api}
			_, err := c.GetDeposits(context.Background(), filter, pagination)
			require.NoError(t, err)
		})
		t.Run("GetDeposit", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			api.On("Get", "/transfers/deposit_id", mock.IsType(&Deposit{})).Return(nil)
			c := Client{api: &api}
			_, err := c.GetDeposit(context.Background(), "deposit_id")
			require.NoError(t, err)
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("CreatePaymentMethodDeposit", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			spec := PaymentMethodDepositSpec{
				Amount:          decimal.NewFromFloat(1.0),
				Currency:        "USD",
				PaymentMethodID: "payment_method_id",
			}
			result := struct {
				ID string `json:"id"`
			}{}
			api.On("Post", "/deposits/payment-method/", spec, &result).Return(nil)
			api.On("Get", mock.MatchedBy(func(s string) bool {
				return strings.HasPrefix(s, "/transfers/")
			}), mock.IsType(&Deposit{})).Return(nil)
			c := Client{api: &api}
			_, err := c.CreatePaymentMethodDeposit(context.Background(), spec)
			require.NoError(t, err)
		})
		t.Run("CreateCoinbaseAccountDeposit", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			spec := CoinbaseAccountDeposit{
				Amount:            decimal.NewFromFloat(1.0),
				Currency:          "USD",
				CoinbaseAccountID: "coinbase-account-id",
			}
			result := struct {
				ID string `json:"id"`
			}{}
			api.On("Post", "/deposits/coinbase-account/", spec, &result).Return(nil)
			api.On("Get", mock.MatchedBy(func(s string) bool {
				return strings.HasPrefix(s, "/transfers/")
			}), mock.IsType(&Deposit{})).Return(nil)
			c := Client{api: &api}
			_, err := c.CreateCoinbaseAccountDeposit(context.Background(), spec)
			require.NoError(t, err)
		})
		t.Run("CreateCryptoDepositAddress", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			api.On("Post", "/coinbase-accounts/coinbase-account-id/addresses/", nil, mock.IsType(&CryptoDepositAddress{})).Return(nil)
			c := Client{api: &api}
			_, err := c.CreateCryptoDepositAddress(context.Background(), "coinbase-account-id")
			require.NoError(t, err)
		})
	})
}

func TestClient_Withdrawals(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		t.Run("GetWithdrawals", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			filter := WithdrawalFilter{
				ProfileID: "profile_id",
				Type:      "withdraw",
			}
			pagination := PaginationParams{
				After: "after",
				Limit: 2,
			}
			api.On("Get", "/transfers/?profile_id=profile_id&type=withdraw&after=after&limit=2", mock.IsType(&Withdrawals{})).Return(nil)
			c := Client{api: &api}
			_, err := c.GetWithdrawals(context.Background(), filter, pagination)
			require.NoError(t, err)
		})
		t.Run("GetWithdrawal", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			api.On("Get", "/transfers/withdrawal_id", mock.IsType(&Withdrawal{})).Return(nil)
			c := Client{api: &api}
			_, err := c.GetWithdrawal(context.Background(), "withdrawal_id")
			require.NoError(t, err)
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("CreatePaymentMethodWithdrawal", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			spec := PaymentMethodWithdrawalSpec{
				Amount:          decimal.NewFromFloat(1.01),
				Currency:        "USD",
				PaymentMethodID: "payment_method_id",
			}
			result := struct {
				ID string `json:"id"`
			}{}
			api.On("Post", "/withdrawals/payment-method/", spec, &result).Return(nil)
			api.On("Get", mock.MatchedBy(func(s string) bool {
				return strings.HasPrefix(s, "/transfers/")
			}), mock.IsType(&Withdrawal{})).Return(nil)
			c := Client{api: &api}
			_, err := c.CreatePaymentMethodWithdrawal(context.Background(), spec)
			require.NoError(t, err)
		})
		t.Run("CreateCoinbaseAccountWithdrawal", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			spec := CoinbaseAccountWithdrawalSpec{
				Amount:            decimal.NewFromFloat(1.01),
				Currency:          "USD",
				CoinbaseAccountID: "coinbase_account_id",
			}
			result := struct {
				ID string `json:"id"`
			}{}
			api.On("Post", "/withdrawals/coinbase-account/", spec, &result).Return(nil)
			api.On("Get", mock.MatchedBy(func(s string) bool {
				return strings.HasPrefix(s, "/transfers/")
			}), mock.IsType(&Withdrawal{})).Return(nil)
			c := Client{api: &api}
			_, err := c.CreateCoinbaseAccountWithdrawal(context.Background(), spec)
			require.NoError(t, err)
		})
		t.Run("CreateCryptoAddressWithdrawal", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			spec := CryptoAddressWithdrawalSpec{
				AddNetworkFeeToTotal: true,
				Amount:               decimal.NewFromFloat(1.01),
				CryptoAddress:        "crypto_address",
				Currency:             "USD",
				DestinationTag:       "destination_tag",
				NoDestinationTag:     false,
			}
			result := struct {
				ID string `json:"id"`
			}{}
			api.On("Post", "/withdrawals/crypto/", spec, &result).Return(nil)
			api.On("Get", mock.MatchedBy(func(s string) bool {
				return strings.HasPrefix(s, "/transfers/")
			}), mock.IsType(&Withdrawal{})).Return(nil)
			c := Client{api: &api}
			_, err := c.CreateCryptoAddressWithdrawal(context.Background(), spec)
			require.NoError(t, err)
		})

		t.Run("FeeEstimate", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			cryptoAddress := CryptoAddress{
				Currency:      "USD",
				CryptoAddress: "crypto_address",
			}
			api.On("Get", "/withdrawals/fee-estimate/?currency=USD&crypto_address=crypto_address", mock.IsType(&WithdrawalFeeEstimate{})).Return(nil)
			c := Client{api: &api}
			_, err := c.GetWithdrawalFeeEstimate(context.Background(), cryptoAddress)
			require.NoError(t, err)
		})
	})
}

func TestClient_CreateStablecoinConversion(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	spec := StablecoinConversionSpec{
		Amount: decimal.NewFromFloat(1.01),
		From:   "BTC",
		To:     "USD",
	}
	api.On("Post", "/conversions/", spec, mock.IsType(&StablecoinConversion{})).Return(nil)
	c := Client{api: &api}
	_, err := c.CreateStablecoinConversion(context.Background(), spec)
	require.NoError(t, err)
}

func TestClient_ListPaymentMethods(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	api.On("Get", "/payment-methods/", mock.IsType(&[]PaymentMethod{})).Return(nil)
	c := Client{api: &api}
	_, err := c.ListPaymentMethods(context.Background())
	require.NoError(t, err)
}

func TestClient_ListCoinbaseAccounts(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	api.On("Get", "/coinbase-accounts/", mock.IsType(&[]CoinbaseAccount{})).Return(nil)
	c := Client{api: &api}
	_, err := c.ListCoinbaseAccounts(context.Background())
	require.NoError(t, err)
}

func TestClient_GetFees(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	api.On("Get", "/fees/", mock.IsType(&Fees{})).Return(nil)
	c := Client{api: &api}
	_, err := c.GetFees(context.Background())
	require.NoError(t, err)
}

func TestClient_Report(t *testing.T) {
	t.Run("CreateReport", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		rawTime, err := time.Parse(time.RFC3339Nano, "2021-04-09T19:04:58.964459Z")
		require.NoError(t, err)
		timestamp := Time(rawTime)
		spec := ReportSpec{
			AccountID: "account_id",
			EndDate:   timestamp,
			Email:     "bob.dobbs@subgenius.com",
			Format:    "pdf",
			ProductID: "product_id",
			StartDate: timestamp,
			Type:      "account",
		}
		result := struct {
			ID string `json:"id"`
		}{}
		api.On("Post", "/reports/", spec, &result).Return(nil)
		api.On("Get", mock.MatchedBy(func(s string) bool {
			return strings.HasPrefix(s, "/reports/")
		}), mock.IsType(&Report{})).Return(nil)
		c := Client{api: &api}
		_, err = c.CreateReport(context.Background(), spec)
		require.NoError(t, err)
	})
	t.Run("GetReport", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/reports/report_id", mock.IsType(&Report{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetReport(context.Background(), "report_id")
		require.NoError(t, err)
	})
}

func TestClient_Profile(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		t.Run("ListProfiles", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			f := ProfileFilter{
				Active: true,
			}
			api.On("Get", "/profiles/?active", mock.IsType(&[]Profile{})).Return(nil)
			c := Client{api: &api}
			_, err := c.ListProfiles(context.Background(), f)
			require.NoError(t, err)
		})
		t.Run("GetProfile", func(t *testing.T) {
			var api mockAPI
			defer api.AssertExpectations(t)
			api.On("Get", "/profiles/profile_id", mock.IsType(&Profile{})).Return(nil)
			c := Client{api: &api}
			_, err := c.GetProfile(context.Background(), "profile_id")
			require.NoError(t, err)
		})
	})
	t.Run("CreateProfileTransfer", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		spec := ProfileTransferSpec{
			Amount:   decimal.NewFromFloat(1.01),
			Currency: "USD",
			From:     "profile_id",
			To:       "other_profile_id",
		}
		api.On("Post", "/profiles/transfer", spec, mock.IsType(&ProfileTransfer{})).Return(nil)
		c := Client{api: &api}
		_, err := c.CreateProfileTransfer(context.Background(), spec)
		require.NoError(t, err)
	})
}

func TestClient_ListProducts(t *testing.T) {
	t.Run("ListProducts", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/products/", mock.IsType(&[]Product{})).Return(nil)
		c := Client{api: &api}
		_, err := c.ListProducts(context.Background())
		require.NoError(t, err)
	})
	t.Run("GetProduct", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/products/product_id", mock.IsType(&Product{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetProduct(context.Background(), "product_id")
		require.NoError(t, err)
	})
}

func TestClient_OrderBook(t *testing.T) {
	t.Run("GetAggregatedOrderBook", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/products/product_id/book/?level=1", mock.IsType(&AggregatedOrderBook{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetAggregatedOrderBook(context.Background(), "product_id", 1)
		require.NoError(t, err)
	})
	t.Run("GetOrderBook", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/products/product_id/book/?level=3", mock.IsType(&OrderBook{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetOrderBook(context.Background(), "product_id")
		require.NoError(t, err)
	})
}

func TestClient_GetProductTicker(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	api.On("Get", "/products/product_id/ticker", mock.IsType(&ProductTicker{})).Return(nil)
	c := Client{api: &api}
	_, err := c.GetProductTicker(context.Background(), "product_id")
	require.NoError(t, err)
}

func TestClient_GetProductTrades(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	pagination := PaginationParams{
		After: "after",
		Limit: 2,
	}
	api.On("Get", "/products/product_id/trades/?after=after&limit=2", mock.IsType(&ProductTrades{})).Return(nil)
	c := Client{api: &api}
	_, err := c.GetProductTrades(context.Background(), "product_id", pagination)
	require.NoError(t, err)
}

func TestClient_GetHistoricRates(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	rawTime, err := time.Parse(time.RFC3339Nano, "2021-04-09T19:04:58.964459Z")
	require.NoError(t, err)
	timestamp := Time(rawTime)
	filter := HistoricRateFilter{
		Granularity: 300,
		End:         timestamp,
		Start:       timestamp,
	}
	api.On("Get", "/products/product_id/candles/?granularity=300&end=2021-04-09T19:04:58.964459Z&start=2021-04-09T19:04:58.964459Z", mock.IsType(&HistoricRates{})).Return(nil)
	c := Client{api: &api}
	_, err = c.GetHistoricRates(context.Background(), "product_id", filter)
	require.NoError(t, err)
}

func TestClient_GetProductStats(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	api.On("Get", "/products/product_id/stats", mock.IsType(&ProductStats{})).Return(nil)
	c := Client{api: &api}
	_, err := c.GetProductStats(context.Background(), "product_id")
	require.NoError(t, err)
}

func TestClient_Currencies(t *testing.T) {
	t.Run("ListCurrencies", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/currencies/", mock.IsType(&[]Currency{})).Return(nil)
		c := Client{api: &api}
		_, err := c.ListCurrencies(context.Background())
		require.NoError(t, err)
	})
	t.Run("GetCurrencies", func(t *testing.T) {
		var api mockAPI
		defer api.AssertExpectations(t)
		api.On("Get", "/currencies/currency_name", mock.IsType(&Currency{})).Return(nil)
		c := Client{api: &api}
		_, err := c.GetCurrency(context.Background(), "currency_name")
		require.NoError(t, err)
	})
}

func TestClient_ServerTime(t *testing.T) {
	var api mockAPI
	defer api.AssertExpectations(t)
	api.On("Get", "/time", mock.IsType(&ServerTime{})).Return(nil)
	c := Client{api: &api}
	_, err := c.GetServerTime(context.Background())
	require.NoError(t, err)
}

func TestWebsocket_watch(t *testing.T) {
	var c Client
	var r mockJSONReader
	defer r.AssertExpectations(t)
	r.On("ReadJSON", mock.Anything).Return([]byte(`{"key":"k","value":"v"}`), nil)
	f := NewFeed()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		read := <-f.Messages
		assert.Equal(t, map[string]interface{}{"key": "k", "value": "v"}, read)
		cancel()
	}()

	err := c.watch(ctx, &r, f)
	assert.True(t, errors.Is(err, context.Canceled))
}

type mockJSONReader struct {
	mock.Mock
}

func (m *mockJSONReader) ReadJSON(v interface{}) error {
	args := m.Called(v)
	err := json.Unmarshal(args.Get(0).([]byte), v)
	if err != nil {
		panic(err)
	}
	return args.Error(1)
}

func TestClient_paginate(t *testing.T) {
	resp := http.Response{
		Header: http.Header{
			"Cb-After":  []string{"after"},
			"Cb-Before": []string{"before"},
		},
	}
	paged := struct {
		Page *Pagination
	}{}
	require.NoError(t, paginate(&resp, &paged))
	assert.True(t, paged.Page.NotEmpty())
	assert.Equal(t, "after", paged.Page.After)
	assert.Equal(t, "before", paged.Page.Before)
}

type mockAPI struct {
	mock.Mock
}

func (m *mockAPI) Get(_ context.Context, relativePath string, result interface{}) error {
	args := m.Called(relativePath, result)
	return args.Error(0)
}

func (m *mockAPI) Do(_ context.Context, method string, relativePath string, content interface{}, result interface{}) (capture error) {
	args := m.Called(method, relativePath, content, result)
	return args.Error(0)
}

func (m *mockAPI) Post(_ context.Context, relativePath string, content interface{}, result interface{}) error {
	args := m.Called(relativePath, content, result)
	return args.Error(0)
}

func (m *mockAPI) Close() error {
	args := m.Called()
	return args.Error(0)
}
