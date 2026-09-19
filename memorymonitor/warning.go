package memorymonitor

import (
	"context"

	"github.com/alexballas/portal/internal/apis"
)

// LowMemoryWarning contains the information given in such a warning.
type LowMemoryWarning struct {
	Level byte // Representing the level of low memory warning.
}

// OnSignalLowMemoryWarning listens for the LowMemoryWarning signal.
// Signal is emitted when a particular low memory situation happens,
// with 0 being the lowest level of memory availability warning,
// and 255 being the highest.
//
// This function blocks for the lifetime of the subscription; the
// subscription is released only when the process exits.
//
// Deprecated: Use OnSignalLowMemoryWarningContext to allow cancellation.
func OnSignalLowMemoryWarning(callback func(warning LowMemoryWarning)) error {
	return OnSignalLowMemoryWarningContext(context.Background(), callback)
}

// OnSignalLowMemoryWarningContext listens until ctx is cancelled and releases the subscription.
// It returns ctx.Err() on cancellation. Callbacks run synchronously and must
// return before cancellation can finish. A nil context means context.Background().
func OnSignalLowMemoryWarningContext(ctx context.Context, callback func(warning LowMemoryWarning)) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	signal, cleanup, err := apis.ListenOnSignal(interfaceName, "LowMemoryWarning")
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
		case sig, ok := <-signal:
			if !ok {
				return nil
			}
			if len(sig.Body) == 0 {
				continue
			}

			level, ok := sig.Body[0].(byte)
			if !ok {
				continue
			}

			callback(LowMemoryWarning{Level: level})
		}
	}
}
