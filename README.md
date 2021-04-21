# Reticule
A Golang Crypto toolkit featuring a Coinbase Pro REST API client

## Quickstart
To get the full `reticule` toolkit:

`go get -u github.com/durp/reticule`

or to just install the `coinbasepro` client:

`go get -u github.com/durp/pkg/coinbasepro`

## Features
A suite of Go clients and tools for the [Coinbase Pro REST API](https://pro.coinbase.com)
- The `coinbasepro.APIClient`
  A low-level, bare-bones, Coinbase Pro REST client that [signs](https://docs.pro.coinbase.com/#signing-a-message) requests 
  and decodes Coinbase Pro REST responses.
- The `coinbasepro.Client`
  A high-level, full-featured Coinbase Pro REST client that provides structured interaction for all endpoints described in
  the [Coinbase Pro](https://docs.pro.coinbase.com) documentation.
- The `reticule` CLI, a command-line interface for the Coinbase Pro API. `reticule` makes it easy to jumpstart interactions 
  with Coinbase Pro, with no coding required.
  
Props due to [go-coinbasepro](https://github.com/preichenberger/go-coinbasepro) for providing guidance and a test bed 
for building the low-level client.

## Usage

### Coinbase Pro prerequisites
1. Open a [Coinbase](https://www.coinbase.com/join/4ty6) account[*](#support-open-source-development)

1. Open a [Coinbase Pro Sandbox](https://public.sandbox.pro.coinbase.com) account, if you just want to play around

1. (Optional) Open a [Coinbase Pro](https://pro.coinbase.com) account, if you are ready to trade

1. [Create API keys](https://help.coinbase.com/en/pro/other-topics/api/how-do-i-create-an-api-key-for-coinbase-pro)


### Using the `reticule` CLI

Using the `reticule` CLI and the sandbox is a great way to learn your way around the Coinbase Pro API. Once you have
opened your accounts and created an API key, just use `reticule` to create a Sandbox config:

`reticule config create coinbase --name <name> --key <key>> --passphrase <passphrase> --secret <secret>`

Reticule makes it easy to interact with Coinbase Pro API

####Commands:
All commands listed below follow the `reticule` root command. To delete a coinbase config, for example, the full command
would be:
`reticule config delete coinbase --name <name>`

|Command |Description  |
--- | --- |
`config create coinbase`                          | create a new coinbasepro config
`config delete coinbase`                          | delete an existing coinbasepro config
`config update coinbase`                          | update an existing coinbasepro config
`coinbase cancel order`                           | cancel an order
`coinbase create order limit`                     | create a limit order (default)
`coinbase create order market`                    | create a market order
`coinbase create conversion`                      | convert crypto to stablecoin
`coinbase create deposit payment-method`          | create payment method deposit
`coinbase create deposit coinbasepro-account`     | create coinbase account deposit
`coinbase create deposit-address`                 | create a crypto deposit address
`coinbase create profile-transfer`                | transfer between profiles
`coinbase create report`                          | create a report of account or fill activity
`coinbase create withdrawal coinbasepro-account`  | create withdrawal from coinbasepro account
`coinbase create withdrawal crypto-address`       | create withdrawal from crypto address
`coinbase create withdrawal payment-method`       | create withdrawal from payment method
`coinbase get accounts `                          | get accounts and account details
`coinbase get coinbase-accounts`                  | get coinbase accounts and details
`coinbase get currencies`                         | get currencies and currency details
`coinbase get deposits`                           | get deposits and deposit details
`coinbase get fees`                               | get fees and fee details
`coinbase get fills`                              | get fills and fill details
`coinbase get holds`                              | get holds and hold details
`coinbase get ledger`                             | get ledger of transactions and transaction details
`coinbase get limits`                             | get limits and limit details
`coinbase get orders`                             | get orders and order details
`coinbase get payment-methods`                    | get payment methods and payment method details
`coinbase get products`                           | get products and product details
`coinbase get product-book`                       | get order book for a product
`coinbase get product-history`                    | get history for a product
`coinbase get product-stats`                      | get stats for a product
`coinbase get product-ticker`                     | get current ticker for a product
`coinbase get product-trades `                    | get trades for a product
`coinbase get profiles `                          | get profiles and profile details
`coinbase get report `                            | get status of a report
`coinbase get server-time`                        | get current server time
`coinbase get withdrawals`                        | get withdrawals and withdrawal details
`coinbase get withdrawal-fee`                     | get estimated fee for a withdrawal
`coinbase watch`                                  | watch the websocket feed

#### Create commands
Most create commands require a JSON spec describing the resource to be created. The command help, where possible, provides
an example representation of the minimum required fields.

For example:

`coinbase create order limit --help` 

yields: 

`-o, --order=LIMIT-ORDER    json {"size": "0.01","price": "0.100","side": "buy","product_id": "BTC-USD"}`

And the full command to create this limit order would be:
```
reticule cb create order limit -o `{"size": "0.01","price": "0.100","side": "buy","product_id": "BTC-USD"}`
```

### Using the coinbasepro.Client

Make a new Client:
```
  client, err := coinbasepro.NewClient(baseURL, feedURL,
    &coinbasepro.Auth{
      Key: "key",
      Passphrase: "passphrase",
      Secret: "secret"},
  )
```

Then use it to interact with Coinbase Pro:
```
  accounts, _ := cb.ListAccounts(ctx)
  for _, account := range accounts {
    fmt.Println("%+v", account)
  }
  account, _ := cb.GetAccounts(ctx, "account-id")
  fmt.Println("%+v", account)
```

### Support Open Source Development
`*` Full disclosure, if you use this [link to open a Coinbase account](https://www.coinbase.com/join/4ty6)
and spend $100, I get $10. It's a nice, no cost  way to support `reticule` development.

Or feel free to just throw some monetary support towards the project in the form of a [crypto donation](https://commerce.coinbase.com/checkout/69d34200-9a06-4d31-849a-07001fdbe0c4)