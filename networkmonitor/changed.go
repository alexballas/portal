package networkmonitor

import (
	"context"

	"github.com/alexballas/portal/internal/apis"
)

// OnSignalChanged calls the passed function when the network configuration changes.
//
// This function blocks for the lifetime of the subscription; the subscription
// is released only when the process exits.
//
// Deprecated: Use OnSignalChangedContext to allow cancellation.
func OnSignalChanged(callback func()) error {
	return OnSignalChangedContext(context.Background(), callback)
}

// OnSignalChangedContext listens until ctx is cancelled and releases the subscription.
// It returns ctx.Err() on cancellation. Callbacks run synchronously and must
// return before cancellation can finish. A nil context means context.Background().
func OnSignalChangedContext(ctx context.Context, callback func()) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	signal, cleanup, err := apis.ListenOnSignal(interfaceName, "changed")
	if err != nil {
		return err
	}

	defer cleanup()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-signal:
			if !ok {
				return nil
			}
			callback()
		}
	}
}
