package watchman

import "github.com/sjansen/watchman/protocol"

// A Subscription represents a request to receive notification of changes to a watched root.
type Subscription struct {
	client       *Client
	name         string
	root         string
	relativeRoot string
}

// Unsubscribe cancels a subscription.
func (s *Subscription) Unsubscribe() (err error) {
	req := &protocol.UnsubscribeRequest{
		Name: s.name,
		Root: s.root,
	}
	_, err = s.client.send(req)

	return
}
