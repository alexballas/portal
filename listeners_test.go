package portal_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexballas/portal/memorymonitor"
	"github.com/alexballas/portal/networkmonitor"
	"github.com/alexballas/portal/settings"
)

func TestListenersAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for name, listen := range map[string]func() error{
		"memory": func() error {
			return memorymonitor.OnSignalLowMemoryWarningContext(ctx, func(memorymonitor.LowMemoryWarning) { t.Error("unexpected callback") })
		},
		"network": func() error {
			return networkmonitor.OnSignalChangedContext(ctx, func() { t.Error("unexpected callback") })
		},
		"settings": func() error {
			return settings.OnSignalSettingChangedContext(ctx, func(settings.Changed) { t.Error("unexpected callback") })
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := listen(); !errors.Is(err, context.Canceled) {
				t.Fatalf("got %v, want context.Canceled", err)
			}
		})
	}
}
