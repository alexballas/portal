package settings

import (
	"context"

	"github.com/alexballas/portal/internal/apis"
	"github.com/godbus/dbus/v5"
)

// Changed is the result given when a setting changes its value.
type Changed struct {
	Namespace string // Namespace of changed setting.
	Key       string // The key of changed setting.
	Value     any    // The new value.
}

// OnSignalSettingChanged listens for the SettingChanged signal.
// This signal is emitted when a setting changes.
//
// This function blocks for the lifetime of the subscription; the subscription
// is released only when the process exits.
//
// Deprecated: Use OnSignalSettingChangedContext to allow cancellation.
func OnSignalSettingChanged(callback func(changed Changed)) error {
	return OnSignalSettingChangedContext(context.Background(), callback)
}

// OnSignalSettingChangedContext listens until ctx is cancelled and releases the subscription.
// It returns ctx.Err() on cancellation. Callbacks run synchronously and must
// return before cancellation can finish. A nil context means context.Background().
func OnSignalSettingChangedContext(ctx context.Context, callback func(changed Changed)) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	signal, cleanup, err := apis.ListenOnSignal(interfaceName, "SettingChanged")
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

			namespace, ok := sig.Body[0].(string)
			if !ok {
				continue // We sometimes get responses from other portals.
			}

			changed := Changed{Namespace: namespace}

			if len(sig.Body) > 1 {
				key, ok := sig.Body[1].(string)
				if !ok {
					continue // Avoid crashing if the response is unexpected.
				}

				changed.Key = key
			}

			if len(sig.Body) > 2 {
				changed.Value = sig.Body[2]
				variant, ok := changed.Value.(dbus.Variant)
				if ok {
					changed.Value = variant.Value()
				}
			}

			callback(changed)
		}
	}
}
