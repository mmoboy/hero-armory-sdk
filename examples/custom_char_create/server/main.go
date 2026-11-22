// Server plugin for custom_char_create: demonstrates Realm-authoritative
// character creation rules, starter classes, validation, and KV persistence.
package main

import (
	"encoding/json"
	"errors"
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
	"github.com/mmoboy/hero-armory-sdk/vanilla"
)

// To preserve the built-in Vanilla face/body customizer and saved look presets
// alongside realm starter classes, switch creationMode to sdk.CharCreateModeVanillaExtended.
const creationMode = sdk.CharCreateModePluginCustom

type originSubmissionData struct {
	Trait string `json:"trait,omitempty"`
}

type originsServer struct {
	sdk.BaseServerPlugin
}

func (originsServer) Init() {
	vanilla.RegisterAll()
	sdk.Log("Sanctuary Origins server plugin initialized")
}

func (originsServer) OnBeginCharacterCreate(ev sdk.BeginCharacterCreateEvent) (sdk.CharacterCreateConfig, bool) {
	return sdk.CharacterCreateConfig{
		Mode:          creationMode,
		Title:         "SANCTUARY ORIGINS",
		Subtitle:      "Select your origin and starting relic, " + ev.DisplayName,
		Classes:       vanilla.SanctuaryStarterClasses(),
		HideNameInput: creationMode == sdk.CharCreateModePluginCustom,
	}, true
}

func (originsServer) OnFinalizeCharacterCreate(ev sdk.FinalizeCharacterCreateEvent) (sdk.CharacterCreateResult, error) {
	name := strings.TrimSpace(ev.HeroName)
	if len(name) < 3 {
		return sdk.CharacterCreateResult{}, errors.New("hero name must be at least 3 characters")
	}
	lower := strings.ToLower(name)
	if lower == "admin" || lower == "system" || lower == "sanctuary" {
		return sdk.CharacterCreateResult{}, errors.New("reserved name rejected by Sanctuary Origins")
	}

	var extra originSubmissionData
	if len(ev.CustomData) > 0 {
		_ = json.Unmarshal(ev.CustomData, &extra)
	}
	trait := strings.TrimSpace(extra.Trait)
	if trait == "" {
		trait = "wanderer"
	}

	classes := vanilla.SanctuaryStarterClasses()
	chosen := classes[0]
	found := false
	for _, cls := range classes {
		if cls.ID == ev.ClassID {
			chosen = cls
			found = true
			break
		}
	}
	if !found && ev.ClassID != "" {
		return sdk.CharacterCreateResult{}, errors.New("unknown origin class: " + ev.ClassID)
	}

	bag := []sdk.ItemStackSpec{
		{ItemID: "cons_wakfu_potion", Quantity: 5},
		{ItemID: "cons_red_heart_elixir", Quantity: 2},
	}
	gold := int64(350)
	if trait == "merchant" {
		gold += 250
	} else if trait == "alchemist" {
		bag = append(bag, sdk.ItemStackSpec{ItemID: "mat_wakfu_crystal", Quantity: 8})
	}

	return sdk.CharacterCreateResult{
		HeroName:            name,
		Equipped:            chosen.Equipped,
		OverrideEquipped:    true,
		StarterBag:          bag,
		OverrideStarterBag:  true,
		StarterGold:         gold,
		OverrideStarterGold: true,
		BaseHP:              100 + chosen.BonusHP,
		BaseAP:              6 + chosen.BonusAP,
		BaseMP:              6 + chosen.BonusMP,
		CustomKV: map[string][]byte{
			"hero:{hero_id}:origin": []byte(chosen.ID + ":" + trait),
		},
		WelcomeNotice: "Entered Sanctuary as " + chosen.Name + " (" + trait + ").",
	}, nil
}

func init() {
	sdk.RegisterServer(originsServer{})
}

func main() {}
