package commands

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCoinbaseConfigCmd(t *testing.T) {
	var cli struct {
		createCoinbaseConfigCmd
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{
		`--name=horatioHornblower`,
		`--base-url=https://api-public.sandbox.pro.coinbase.com`,
		`--feed-url=wss://ws-feed-public.sandbox.pro.coinbase.com`,
		`--key=key`,
		`--passphrase=passphrase`,
		`--secret=secret`,
	})
	fs := afero.NewMemMapFs()
	require.NoError(t, err)
	err = cli.Run(fs)
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	f, err := fs.Open(home + "/.reticule/coinbasepro")
	require.NoError(t, err)
	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(`
current: horatioHornblower
configs:
    horatioHornblower:
        baseurl: https://api-public.sandbox.pro.coinbase.com
        feedurl: wss://ws-feed-public.sandbox.pro.coinbase.com
        auth:
            key: key
            passphrase: passphrase
            secret: secret
`),
		strings.TrimSpace(string(b)))
}

func TestUpdateCoinbaseConfigCmd(t *testing.T) {
	var cli struct {
		updateCoinbaseConfigCmd
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{
		`--name=horatioHornblower`,
		`--base-url=https://api-monkeys.sandbox.pro.coinbase.com`,
		`--feed-url=wss://ws-feed-monkeys-public.sandbox.pro.coinbase.com`,
		`--key=new-key`,
		`--passphrase=new-passphrase`,
		`--rename=new-name`,
		`--secret=new-secret`,
	})
	fs := afero.NewMemMapFs()
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	configPath := home + "/.reticule/coinbasepro"
	f, err := fs.Create(configPath)
	require.NoError(t, err)
	_, err = f.WriteString(`
current: horatioHornblower
configs:
    horatioHornblower:
        baseurl: https://api-public.sandbox.pro.coinbase.com
        feedurl: wss://ws-feed-public.sandbox.pro.coinbase.com
        auth:
            passphrase: passphrase
            secret: secret
`)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)
	err = cli.Run(fs)
	require.NoError(t, err)
	f, err = fs.Open(home + "/.reticule/coinbasepro")
	require.NoError(t, err)
	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(`
current: horatioHornblower
configs:
    new-name:
        baseurl: https://api-monkeys.sandbox.pro.coinbase.com
        feedurl: wss://ws-feed-monkeys-public.sandbox.pro.coinbase.com
        auth:
            key: new-key
            passphrase: new-passphrase
            secret: new-secret
`),
		strings.TrimSpace(string(b)))
}

func mustNew(t *testing.T, cli interface{}, options ...kong.Option) *kong.Kong {
	t.Helper()
	options = append([]kong.Option{
		kong.Name("test"),
		kong.Exit(func(int) {
			t.Helper()
			t.Fatalf("unexpected exit()")
		}),
	}, options...)
	parser, err := kong.New(cli, options...)
	require.NoError(t, err)
	return parser
}
