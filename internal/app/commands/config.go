package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/durp/reticule/pkg/coinbasepro"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type configCmd struct {
	Create createConfigCmd `kong:"cmd,name='create',help='create a new config'"`
	Delete deleteConfigCmd `kong:"cmd,name='delete',help='delete a config'"`
	Update updateConfigCmd `kong:"cmd,name='update',help='update an existing config'"`
}

type createConfigCmd struct {
	Coinbase createCoinbaseConfigCmd `kong:"cmd,name='coinbase',alias='cb',help='create a new coinbasepro config'"`
}

type deleteConfigCmd struct {
	Coinbase deleteCoinbaseConfigCmd `kong:"cmd,name='coinbase',alias='cb',help='delete an existing coinbasepro config'"`
}

type updateConfigCmd struct {
	Coinbase updateCoinbaseConfigCmd `kong:"cmd,name='coinbase',alias='cb',help='update an existing coinbasepro config'"`
}

type coinbaseProConfigSet struct {
	Current string
	Configs map[string]coinbaseProConfig
}

type coinbaseProConfig struct {
	BaseURL string
	FeedURL string
	Auth    *coinbasepro.Auth
}

type createCoinbaseConfigCmd struct {
	Name       string   `kong:"name='name',short='n',help='name of config',required"`
	BaseURL    *url.URL `kong:"name='base-url',short='b',default='https://api-public.sandbox.pro.coinbase.com',help='url of coinbasepro api that provided key'"`
	FeedURL    *url.URL `kong:"name='feed-url',short='f',default='wss://ws-feed-public.sandbox.pro.coinbase.com',help='url of websocket feed'"`
	Key        string   `king:"name='key',short='k',help='coinbasepro provided api key'"`
	Passphrase string   `king:"name='passphrase',short='p',help='coinbasepro api passphrase'"`
	Secret     string   `king:"name='secret',short='s',help='coinbasepro provided api secret'"`
	Use        bool     `king:"name='use',short='s',help='set as config to use'"`
}

func (c *createCoinbaseConfigCmd) Run(fs afero.Fs) (capture error) {
	configPath, err := configPath()
	if err != nil {
		return err
	}
	var configSet coinbaseProConfigSet
	f, err := fs.OpenFile(configPath, os.O_RDWR|os.O_CREATE, 0755)
	switch {
	case errors.Is(err, os.ErrNotExist):
		cfgDir := path.Dir(configPath)
		err = fs.MkdirAll(cfgDir, os.ModePerm)
		if err != nil {
			return err
		}
		fmt.Printf("creating config %q\n", configPath)
		f, err = fs.Create(configPath)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		defer func() { coinbasepro.Capture(&capture, f.Close()) }()
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(b, &configSet)
		if err != nil {
			return err
		}
		if _, ok := configSet.Configs[c.Name]; ok {
			return fmt.Errorf("coinbase config %q already exists, use `config update coinbase` to modify an existing config", c.Name)
		}
	}
	cfg := coinbaseProConfig{
		BaseURL: c.BaseURL.String(),
		FeedURL: c.FeedURL.String(),
		Auth: coinbasepro.NewAuth(
			c.Key,
			c.Passphrase,
			c.Secret),
	}
	if configSet.Configs == nil {
		configSet.Configs = make(map[string]coinbaseProConfig)
		configSet.Current = c.Name
	}
	if c.Use {
		configSet.Current = c.Name
	}
	configSet.Configs[c.Name] = cfg
	enc := yaml.NewEncoder(f)
	err = enc.Encode(&configSet)
	if err != nil {
		return err
	}
	return enc.Close()
}

type deleteCoinbaseConfigCmd struct {
	Name string `kong:"name='name',short='n',help='name of config',required"`
}

func (d *deleteCoinbaseConfigCmd) Run(fs afero.Fs) error {
	configPath, err := configPath()
	if err != nil {
		return err
	}
	configSet, err := readConfigSet(fs, configPath)
	if err != nil {
		return err
	}
	delete(configSet.Configs, d.Name)
	return writeConfigSet(fs, configPath, configSet)
}

type updateCoinbaseConfigCmd struct {
	Name       string   `kong:"name='name',short='n',help='name of config',required"`
	BaseURL    *url.URL `kong:"name='base-url',short='b',help='url of coinbasepro api that provided key'"`
	FeedURL    *url.URL `kong:"name='feed-url',short='f',help='url of websocket feed'"`
	Key        string   `king:"name='key',short='k',help='coinbasepro provided api key'"`
	Passphrase string   `king:"name='passphrase',short='p',help='coinbasepro api passphrase'"`
	Rename     string   `kong:"name='rename',short='r',help='new name for config'"`
	Secret     string   `king:"name='secret',short='s',help='coinbasepro provided api secret'"`
	Use        bool     `king:"name='use',short='s',help='set as config to use'"`
}

func (c *updateCoinbaseConfigCmd) Run(fs afero.Fs) error {
	configPath, err := configPath()
	if err != nil {
		return err
	}
	configSet, err := readConfigSet(fs, configPath)
	if err != nil {
		return err
	}
	cfg, ok := configSet.Configs[c.Name]
	if !ok {
		return fmt.Errorf("coinbase config %q does not exists, use `config create coinbase` to create a new config", c.Name)
	}
	if c.BaseURL != nil {
		cfg.BaseURL = c.BaseURL.String()
	}
	if c.FeedURL != nil {
		cfg.FeedURL = c.FeedURL.String()
	}
	if c.Key != "" {
		cfg.Auth.Key = c.Key
	}
	if c.Passphrase != "" {
		cfg.Auth.Passphrase = c.Passphrase
	}
	if c.Secret != "" {
		cfg.Auth.Secret = c.Secret
	}
	if c.Rename != "" {
		delete(configSet.Configs, c.Name)
		c.Name = c.Rename
	}
	if c.Use {
		configSet.Current = c.Name
	}
	configSet.Configs[c.Name] = cfg
	return writeConfigSet(fs, configPath, configSet)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("no user home directory defined")
	}
	return path.Join(home, ".reticule", "coinbasepro"), nil
}

func readConfigSet(fs afero.Fs, configPath string) (coinbaseProConfigSet, error) {
	var configSet coinbaseProConfigSet
	f, err := fs.Open(configPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return coinbaseProConfigSet{}, fmt.Errorf("config file %q does not exist, create a new config with `create config coinbase`", configPath)
	case err != nil:
		return coinbaseProConfigSet{}, err
	default:
		defer func() { _ = f.Close() }()
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return coinbaseProConfigSet{}, err
		}
		err = yaml.Unmarshal(b, &configSet)
		if err != nil {
			return coinbaseProConfigSet{}, err
		}
		return configSet, nil
	}
}

func writeConfigSet(fs afero.Fs, configPath string, configSet coinbaseProConfigSet) (capture error) {
	f, err := fs.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() { coinbasepro.Capture(&capture, f.Close()) }()
	enc := yaml.NewEncoder(f)
	err = enc.Encode(&configSet)
	if err != nil {
		return err
	}
	return enc.Close()
}
