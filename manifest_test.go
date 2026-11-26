package sdk

import "testing"

func TestValidateManifestAndPaths(t *testing.T) {
	m := PluginManifest{
		ID:          "realm_bounty",
		Name:        "Realm Bounty Board",
		Version:     "1.0.0",
		Author:      "Community",
		Description: "Adds a HUD button and live bounty challenges.",
		Required:    true,
	}
	if err := ValidateManifest(&m); err != nil {
		t.Fatalf("ValidateManifest valid failed: %v", err)
	}
	if m.EntryWasm != "client.wasm" || m.SDKVersion != SDKVersion {
		t.Fatalf("expected default entry/sdk_version, got %+v", m)
	}

	for _, bad := range []string{"../secret.wasm", "/etc/passwd", "hack.exe", "a/../../b.png", "C:evil.wasm", "a:b.png", "a\x00b.png", "a\x07b.png"} {
		if _, err := CleanBundlePath(bad); err == nil {
			t.Fatalf("expected CleanBundlePath(%q) to fail", bad)
		}
	}
	if got := SanitizeASCII("\n\n\n\n\n\n\n\n\n\nabc", 5); len(got) > 5 {
		t.Fatalf("expected SanitizeASCII newline cap <= 5, got len=%d (%q)", len(got), got)
	}
	if got, err := CleanBundlePath("assets/trophy.png"); err != nil || got != "assets/trophy.png" {
		t.Fatalf("expected valid asset path, got %q err=%v", got, err)
	}

	badID := m
	badID.ID = "INVALID SPACE"
	if err := ValidateManifest(&badID); err == nil {
		t.Fatal("expected invalid plugin ID to fail")
	}
}
