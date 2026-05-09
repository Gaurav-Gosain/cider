package fm

import (
	"os"
	"testing"
)

// TestMain attempts to load the Foundation Models library so tests that
// need it can run. Tests that require the library should call requireLib;
// they are skipped when the dylib is not present (e.g. CI on Linux).
func TestMain(m *testing.M) {
	_ = Init()
	os.Exit(m.Run())
}

// requireLib skips a test if the Foundation Models library is not loaded.
func requireLib(t *testing.T) {
	t.Helper()
	if libHandle == 0 {
		t.Skip("Foundation Models library not available")
	}
}
