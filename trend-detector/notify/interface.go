package notify

import "gnai/trend-detector/trend"

// Notifier sends a message for a given trend state event.
type Notifier interface {
	Name() string
	Send(msg Message) error
}

// Message is the rendered payload passed to each Notifier.
type Message struct {
	Text    string
	IsShift bool // true = shift alert, false = periodic digest
	State   trend.State
}
