package commands

import (
	"github.com/sirupsen/logrus"
)

type CLI struct {
	Config   configCmd   `kong:"cmd,name='config',help='create, update or use a config'"`
	Coinbase coinbaseCmd `kong:"cmd,name='coinbase',aliases='cb',help='use the coinbasepro api'"`
	LogLevel logLevel    `kong:"name='log-level',short='v',default='info',help='set level of log, one of [ panic, fatal, error, warn, info, debug, trace ]'"`
}

type logLevel string

func (l logLevel) AfterApply() error {
	lvl, err := logrus.ParseLevel(string(l))
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	return nil
}
