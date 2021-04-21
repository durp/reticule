package coinbasepro

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportSpec(t *testing.T) {
	var timestamp Time
	require.NoError(t, timestamp.UnmarshalJSON([]byte("2021-04-09T19:04:58.964459Z")))
	validReportSpec := func() *ReportSpec {
		r := ReportSpec{
			AccountID: "account_id",
			EndDate:   timestamp,
			Email:     "email@example.com",
			Format:    ReportFormatCSV,
			// ProductID: "BTC-USD",
			StartDate: timestamp,
			Type:      ReportTypeAccount,
		}
		require.NoError(t, r.Validate())
		return &r
	}
	t.Run("Unmarshal", func(t *testing.T) {
		raw := `
  {
    "account_id": "account_id",
    "end_date": "2021-04-09T19:04:58.964459Z",
    "email": "email@example.com",
    "format": "csv",
    "start_date": "2021-04-09T19:04:58.964459Z",
    "type": "account"
  }
`
		var r ReportSpec
		require.NoError(t, json.Unmarshal([]byte(raw), &r))
		assert.Equal(t, validReportSpec(), &r)
	})

	t.Run("Validate", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			assert.NoError(t, validReportSpec().Validate())
		})
		t.Run("TypeInvalid", func(t *testing.T) {
			r := validReportSpec()
			r.Type = "blah"
			err := r.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*type.* must be .*account.* .*fills.*", err.Error())
		})

		t.Run("AccountTypeRequiresAccountID", func(t *testing.T) {
			r := validReportSpec()
			r.Type = ReportTypeAccount
			r.AccountID = ""
			err := r.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*account_id.* required", err.Error())
		})
		t.Run("FillsTypeRequiresProductID", func(t *testing.T) {
			r := validReportSpec()
			r.Type = ReportTypeFills
			r.AccountID = ""
			r.ProductID = ""
			err := r.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*product_id.* required", err.Error())
		})
		t.Run("FormatInvalid", func(t *testing.T) {
			r := validReportSpec()
			r.Format = "blah"
			err := r.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*format.* must be .*pdf.* or .*csv.*", err.Error())
		})
		t.Run("EndDateRequired", func(t *testing.T) {
			r := validReportSpec()
			r.EndDate = Time{}
			err := r.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*end_date.* required", err.Error())
		})
		t.Run("StartDateRequired", func(t *testing.T) {
			r := validReportSpec()
			r.StartDate = Time{}
			err := r.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*start_date.* required", err.Error())
		})
	})
}
