package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/durp/reticule/internal/app/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func main() {
	ctx := context.Background()
	var cli commands.CLI
	ktx := kong.Parse(&cli,
		kong.Name("reticule"),
		kong.Description("crypto coin toolkit"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": "0.0.1",
		},
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.BindTo(afero.NewOsFs(), (*afero.Fs)(nil)),
	)
	err := ktx.Run()
	if err != nil {
		logrus.Errorf("%+v", err)
	}
}
