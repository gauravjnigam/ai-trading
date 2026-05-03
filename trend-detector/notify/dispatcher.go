package notify

import (
	"context"
	"log"

	"gnai/trend-detector/trend"
)

// Dispatcher fans out trend events to all registered Notifiers.
type Dispatcher struct {
	notifiers []Notifier
}

// NewDispatcher creates a Dispatcher with the provided notifiers, skipping nils.
func NewDispatcher(notifiers ...Notifier) *Dispatcher {
	d := &Dispatcher{}
	for _, n := range notifiers {
		if n != nil {
			d.notifiers = append(d.notifiers, n)
		}
	}
	return d
}

// Run listens on the manager's event channels and dispatches messages until
// ctx is cancelled.
func (d *Dispatcher) Run(ctx context.Context, mgr *trend.Manager) {
	for {
		select {
		case <-ctx.Done():
			return
		case st := <-mgr.ShiftEvents:
			d.dispatch(Message{
				Text:    FormatShiftAlert(st),
				IsShift: true,
				State:   st,
			})
		case st := <-mgr.DigestEvents:
			d.dispatch(Message{
				Text:    FormatDigest(st),
				IsShift: false,
				State:   st,
			})
		}
	}
}

func (d *Dispatcher) dispatch(msg Message) {
	for _, n := range d.notifiers {
		if err := n.Send(msg); err != nil {
			log.Printf("notify/%s: %v", n.Name(), err)
		}
	}
}
