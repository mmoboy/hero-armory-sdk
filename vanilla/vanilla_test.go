package vanilla

import (
	"testing"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

func TestSanctuaryStarterClassesAndConfig(t *testing.T) {
	classes := SanctuaryStarterClasses()
	if len(classes) != 4 {
		t.Fatalf("expected 4 sanctuary starter classes, got %d", len(classes))
	}
	seen := map[string]bool{}
	for _, cls := range classes {
		if !sdk.IsValidContentID(cls.ID) {
			t.Fatalf("invalid starter class id %q", cls.ID)
		}
		if seen[cls.ID] {
			t.Fatalf("duplicate starter class id %q", cls.ID)
		}
		seen[cls.ID] = true
		if cls.Name == "" || cls.Description == "" || len(cls.Equipped) == 0 {
			t.Fatalf("starter class %q missing fields: %+v", cls.ID, cls)
		}
		if cls.BonusHP == 0 && cls.BonusAP == 0 && cls.BonusMP == 0 {
			t.Fatalf("expected stat bonuses on starter class %q", cls.ID)
		}
	}

	cfg := SanctuaryStarterConfig()
	if cfg.BaseHP <= 0 || cfg.BaseAP <= 0 || cfg.StarterGold <= 0 || len(cfg.StarterEquipped) == 0 {
		t.Fatalf("expected populated SanctuaryStarterConfig, got %+v", cfg)
	}
}
