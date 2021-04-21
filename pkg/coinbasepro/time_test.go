package coinbasepro

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
		var tm Time
		assert.NoError(t, json.Unmarshal([]byte(`"2006-01-02T15:04:05.999999999Z"`), &tm))
		assert.Equal(t, "2006-01-02 15:04:05.999999999 +0000 UTC", tm.Time().UTC().String())
		fmt.Println(tm.Time().UnixNano())
		assert.NoError(t, json.Unmarshal([]byte(`"2006-01-02T15:04:05.999999999-07:00"`), &tm))
		assert.Equal(t, "2006-01-02 22:04:05.999999999 +0000 UTC", tm.Time().UTC().String())
		assert.NoError(t, json.Unmarshal([]byte(`"2006-01-02T15:04:05Z"`), &tm))
		assert.Equal(t, "2006-01-02 15:04:05 +0000 UTC", tm.Time().UTC().String())
		assert.NoError(t, json.Unmarshal([]byte(`"2006-01-02 15:04:05+00"`), &tm))
		assert.Equal(t, "2006-01-02 15:04:05 +0000 UTC", tm.Time().UTC().String())
		assert.NoError(t, json.Unmarshal([]byte(`"2006-01-02 15:04:05.999999"`), &tm))
		assert.Equal(t, "2006-01-02 15:04:05.999999 +0000 UTC", tm.Time().UTC().String())
		assert.NoError(t, json.Unmarshal([]byte(`"2006-01-02 15:04:05.999999+00"`), &tm))
		assert.Equal(t, "2006-01-02 15:04:05.999999 +0000 UTC", tm.Time().UTC().String())
		assert.NoError(t, json.Unmarshal([]byte(`null`), &tm))
		assert.True(t, tm.Time().IsZero())
		assert.Error(t, json.Unmarshal([]byte(`"unhandled format"`), &tm))
	})
	t.Run("Marshal", func(t *testing.T) {
		b, err := json.Marshal(Time(time.Unix(1136214245, 999999999).UTC()))
		require.NoError(t, err)
		assert.Equal(t, `"2006-01-02T15:04:05.999999999Z"`, string(b))
	})
}
