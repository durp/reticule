package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v2"
)

type Output struct {
	Output OutputType `kong:"name='output',short='o',default='json',enum='json,yaml,none'"`
	w      io.Writer
}

func (o *Output) BeforeApply(ktx *kong.Context) error {
	ktx.BindTo(o, (*encoder)(nil))
	if o.w != nil {
		return nil
	}
	o.w = os.Stdout
	return nil
}

type encoder interface {
	Encode(v interface{}) error
}

type OutputType string

const (
	// OutputTypeNone suppresses all output
	OutputTypeNone OutputType = "none"
	// OutputTypeJSON encodes all output to json (default)
	OutputTypeJSON OutputType = "json"
	// OutputTypeYAML encodes all output to yaml
	OutputTypeYAML OutputType = "yaml"
)

func (o *Output) Encode(value interface{}) (capture error) {
	var encoder encoder
	switch o.Output {
	case OutputTypeNone:
		encoder = noopEncoder{}
	case OutputTypeJSON:
		jsonEncoder := json.NewEncoder(o.w)
		jsonEncoder.SetIndent("", "  ")
		encoder = jsonEncoder
	case OutputTypeYAML:
		encoder = yaml.NewEncoder(o.w)
	default:
		panic(fmt.Sprintf("no encoder defined for output %s", o.Output))
	}
	return encoder.Encode(value)
}

type noopEncoder struct{}

func (n noopEncoder) Encode(_ interface{}) error {
	return nil
}
