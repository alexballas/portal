//go:build integration
// +build integration

package portal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexballas/portal/memorymonitor"
	"github.com/alexballas/portal/networkmonitor"
	"github.com/alexballas/portal/settings"
	"github.com/godbus/dbus/v5"
)

func TestListenerDeliveryAndCancellation(t *testing.T) {
	conn, err := dbus.SessionBus()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, signal string
		body         []any
		listen       func(context.Context, func()) error
	}{
		{"memory", "org.freedesktop.portal.MemoryMonitor.LowMemoryWarning", []any{byte(123)}, func(ctx context.Context, received func()) error {
			return memorymonitor.OnSignalLowMemoryWarningContext(ctx, func(w memorymonitor.LowMemoryWarning) {
				if w.Level != 123 {
					t.Errorf("unexpected warning: %+v", w)
				}
				received()
			})
		}},
		{"network", "org.freedesktop.portal.NetworkMonitor.changed", nil, func(ctx context.Context, received func()) error {
			return networkmonitor.OnSignalChangedContext(ctx, received)
		}},
		{"settings", "org.freedesktop.portal.Settings.SettingChanged", []any{"org.freedesktop.appearance", "color-scheme", dbus.MakeVariant(uint32(1))}, func(ctx context.Context, received func()) error {
			return settings.OnSignalSettingChangedContext(ctx, func(c settings.Changed) {
				if c.Namespace != "org.freedesktop.appearance" || c.Key != "color-scheme" || c.Value != uint32(1) {
					t.Errorf("unexpected setting: %+v", c)
				}
				received()
			})
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			received := make(chan struct{}, 1)
			go func() {
				done <- tc.listen(ctx, func() {
					select {
					case received <- struct{}{}:
					default:
					}
				})
			}()
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			timeout := time.NewTimer(3 * time.Second)
			defer timeout.Stop()
		wait:
			for {
				select {
				case <-ticker.C:
					if err := conn.Emit("/org/freedesktop/portal/desktop", tc.signal, tc.body...); err != nil {
						t.Fatal(err)
					}
				case <-received:
					break wait
				case err := <-done:
					t.Fatalf("listener exited early: %v", err)
				case <-timeout.C:
					t.Fatal("signal delivery timed out")
				}
			}
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("got %v, want context.Canceled", err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("cancellation did not stop listener")
			}
		})
	}
}
