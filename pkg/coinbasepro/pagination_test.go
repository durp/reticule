package coinbasepro

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginationParams(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			p := PaginationParams{
				After: "after",
			}
			require.NoError(t, p.Validate())
			p = PaginationParams{
				Before: "before",
			}
			require.NoError(t, p.Validate())
			p = PaginationParams{
				Limit: 10,
			}
			require.NoError(t, p.Validate())
		})
		t.Run("Invalid", func(t *testing.T) {
			t.Run("Token", func(t *testing.T) {
				p := PaginationParams{
					After:  "after",
					Before: "before",
				}
				err := p.Validate()
				require.Error(t, err)
				assert.Regexp(t, "only one of", err.Error())
			})
			t.Run("Limit", func(t *testing.T) {
				p := PaginationParams{
					Limit: -1,
				}
				err := p.Validate()
				require.Error(t, err)
				assert.Regexp(t, ".*limit.* outside of allowed range", err.Error())
			})
			p := PaginationParams{
				Limit: 101,
			}
			err := p.Validate()
			require.Error(t, err)
			assert.Regexp(t, ".*limit.* outside of allowed range", err.Error())
		})
	})
	t.Run("Params", func(t *testing.T) {
		p := PaginationParams{
			After: "after",
			Limit: 1,
		}
		assert.ElementsMatch(t, []string{"after=after", "limit=1"}, p.Params())
		p = PaginationParams{
			Before: "before",
		}
		assert.ElementsMatch(t, []string{"before=before"}, p.Params())
	})
}

func TestPagination(t *testing.T) {
	p := Pagination{
		After:  "after",
		Before: "before",
	}
	assert.True(t, p.NotEmpty())
}
