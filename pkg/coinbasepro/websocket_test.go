package coinbasepro

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFeed(t *testing.T) {
	f := NewFeed()
	assert.NotNil(t, f.Messages)
	assert.NotNil(t, f.Subscriptions)
}
