package coinbasepro

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSubscriptionRequest(t *testing.T) {
	req := NewSubscriptionRequest([]ProductID{"BTC-ETH"}, []ChannelName{
		ChannelNameStatus,
		ChannelNameTicker,
		ChannelNameLevel2,
		ChannelNameFull,
		ChannelNameUser,
		ChannelNameMatches,
	}, []Channel{{
		Name:       ChannelNameHeartbeat,
		ProductIDs: []ProductID{"BTC-USD"},
	}})
	assert.Equal(t, MessageTypeSubscribe, req.Type)
}
