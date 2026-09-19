//go:build integration
// +build integration

package request

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestSendRequestEntropyFailure(t *testing.T) {
	if _, err := dbus.SessionBus(); err != nil {
		t.Fatal(err)
	}
	original := rand.Reader
	t.Cleanup(func() { rand.Reader = original })
	want := errors.New("entropy unavailable")
	rand.Reader = failingReader{want}
	called := false
	response, err := SendRequest(context.Background(), "", "com.example.Call", func(string) []any {
		called = true
		return nil
	})
	if !errors.Is(err, want) || response.Status != Ended || called {
		t.Fatalf("response=%+v error=%v buildArgs called=%v", response, err, called)
	}
}
