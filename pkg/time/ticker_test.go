package time

import (
	"testing"
	"time"
)

func TestNewTicker(t *testing.T) {
	t.Parallel()

	t.Run("immediately disabled", func(t *testing.T) {
		t.Parallel()

		ticker := NewTicker(5*time.Second, false)

		select {
		case <-time.After(6 * time.Second):
			t.Fatal("does not work as std ticker")
		case <-ticker.C:
			// ok
		}
	})

	t.Run("immediately enabled", func(t *testing.T) {
		t.Parallel()

		ticker := NewTicker(5*time.Second, true)

		select {
		case <-time.After(1 * time.Millisecond):
			t.Fatal("should emit immediately")
		case <-ticker.C:
			// ok
		}

		select {
		case <-time.After(6 * time.Second):
			t.Fatal("should emit after duration")
		case <-ticker.C:
			// ok
		}
	})
}
