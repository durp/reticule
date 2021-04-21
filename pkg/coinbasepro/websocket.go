package coinbasepro

func NewFeed() Feed {
	return Feed{
		Subscriptions: make(chan SubscriptionRequest, 1),
		Messages:      make(chan interface{}),
	}
}

type Feed struct {
	Subscriptions chan SubscriptionRequest
	Messages      chan interface{}
}
