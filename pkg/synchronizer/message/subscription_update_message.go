package message

import "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer/subscription"

type SubscriptionUpdateMessage struct {
	Added   []subscription.Subscription
	Removed []subscription.Subscription
}
