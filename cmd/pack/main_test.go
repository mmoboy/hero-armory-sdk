package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackDirectoryRejectsSymlinkAndInvalidWasm(t *testing.T) {
	dir := t.TempDir()
	manifest := `{"sdk_version":1,"id":"pack_test","name":"Pack Test","version":"1.0.0","entry_wasm":"client.wasm"}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	// Non-wasm bytes must be rejected.
	if err := os.WriteFile(filepath.Join(dir, "client.wasm"), []byte("not-a-wasm"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out.gtpack")
	if err := packDirectory(dir, out); err == nil || !strings.Contains(err.Error(), "WebAssembly") {
		t.Fatalf("expected invalid wasm error, got: %v", err)
	}

	// Write valid wasm header, then add a symlinked asset -> must be rejected.
	validWasm := []byte{0x00, 'a', 's', 'm', 0x01, 0x00, 0x00, 0x00}
	if err := os.WriteFile(filepath.Join(dir, "client.wasm"), validWasm, 0o644); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(outside, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(assetsDir, "leak.json")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := packDirectory(dir, out); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink rejection error, got: %v", err)
	}
}
