package commands

import (
	"bytes"
	"io"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutput(t *testing.T) {
	t.Run("OutputType", func(t *testing.T) {
		var cli struct {
			Output
		}
		newParser := func(w io.Writer) *kong.Kong {
			cli.Output = Output{
				w: w,
			}
			return mustNew(t, &cli)
		}
		s := struct {
			String string `json:"string"`
		}{
			String: "string",
		}
		t.Run("JSON", func(t *testing.T) {
			var w bytes.Buffer
			k := newParser(&w)
			_, err := k.Parse([]string{`--output=json`})
			require.NoError(t, err)
			err = cli.Output.Encode(s)
			require.NoError(t, err)
			assert.JSONEq(t, `{"string":"string"}`, w.String())
		})
		t.Run("YAML", func(t *testing.T) {
			var w bytes.Buffer
			k := newParser(&w)
			_, err := k.Parse([]string{`--output=yaml`})
			require.NoError(t, err)
			err = cli.Output.Encode(s)
			require.NoError(t, err)
			assert.Equal(t, "string: string\n", w.String())
		})
		t.Run("NoOp", func(t *testing.T) {
			var w bytes.Buffer
			k := newParser(&w)
			_, err := k.Parse([]string{`--output=none`})
			require.NoError(t, err)
			err = cli.Output.Encode(s)
			require.NoError(t, err)
			assert.Equal(t, "", w.String())
		})
		t.Run("os.StdOut", func(t *testing.T) {
			k := newParser(nil)
			_, err := k.Parse([]string{})
			require.NoError(t, err)
			err = cli.Output.Encode(s)
			require.NoError(t, err)
		})
		t.Run("EncodeNoOutputTypePanic", func(t *testing.T) {
			assert.Panics(t, func() {
				cli.Output.Output = "blah"
				_ = cli.Output.Encode(s)
			})
		})
	})
}
