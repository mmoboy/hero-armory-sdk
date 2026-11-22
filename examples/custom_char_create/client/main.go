// Client plugin for custom_char_create: renders a custom declarative creation
// overlay, drives the live 3D Sanctuary pedestal via SetCharacterCreatePreview,
// and submits final hero creation via SubmitCharacterCreate.
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

type originClientData struct {
	Trait string `json:"trait,omitempty"`
}

type originsClient struct {
	sdk.BaseClientPlugin
	cfg         sdk.CharacterCreateConfig
	serverName  string
	selectedID  string
	selectedIdx int
	trait       string
	faceFocus   bool
	hairStyle   int
	eyeHueIdx   int
}

var eyePalettes = [][4]float32{
	{0.20, 0.85, 0.48, 1},
	{0.26, 0.68, 0.98, 1},
	{0.92, 0.64, 0.22, 1},
	{0.76, 0.42, 0.96, 1},
}

func (c *originsClient) Init() {
	c.trait = "wanderer"
}

func (c *originsClient) OnEnterCharacterCreate(ctx sdk.CharacterCreateContext) {
	c.cfg = ctx.Config
	c.serverName = ctx.ServerName
	c.trait = "wanderer"
	c.selectedIdx = 0
	c.selectedID = ""
	if len(ctx.Config.Classes) > 0 {
		c.selectedID = ctx.Config.Classes[0].ID
	}
	c.pushPreview()
	c.renderCreateOverlay()
}

func (c *originsClient) OnEnterWorld(self sdk.HeroSnapshot) {
	sdk.ClearOverlay()
}

func (c *originsClient) OnUIEvent(ev sdk.UIClickEvent) {
	id := ev.ControlID
	switch {
	case strings.HasPrefix(id, "class_"):
		target := strings.TrimPrefix(id, "class_")
		for i, cls := range c.cfg.Classes {
			if cls.ID == target {
				c.selectedIdx = i
				c.selectedID = cls.ID
				break
			}
		}
		c.pushPreview()
		c.renderCreateOverlay()

	case strings.HasPrefix(id, "trait_"):
		c.trait = strings.TrimPrefix(id, "trait_")
		c.renderCreateOverlay()

	case id == "btn_toggle_cam":
		c.faceFocus = !c.faceFocus
		c.pushPreview()
		c.renderCreateOverlay()

	case id == "btn_cycle_look":
		c.hairStyle = (c.hairStyle + 1) % 6
		c.eyeHueIdx = (c.eyeHueIdx + 1) % len(eyePalettes)
		c.pushPreview()
		c.renderCreateOverlay()

	case id == "btn_forge_submit":
		name := strings.TrimSpace(ev.Inputs["origin_hero_name"])
		rawCustom, _ := json.Marshal(originClientData{Trait: c.trait})
		_ = sdk.SubmitCharacterCreate(sdk.CharacterCreateSubmission{
			Name:             name,
			ClassID:          c.selectedID,
			CustomData:       rawCustom,
			EnterImmediately: true,
		})
	}
}

func (c *originsClient) pushPreview() {
	focus := "body"
	if c.faceFocus {
		focus = "face"
	}
	var eq map[string]string
	if c.selectedIdx >= 0 && c.selectedIdx < len(c.cfg.Classes) {
		eq = c.cfg.Classes[c.selectedIdx].Equipped
	}
	_ = sdk.SetCharacterCreatePreview(sdk.CharacterCreatePreviewSpec{
		Equipped:      eq,
		ApplyEquipped: len(eq) > 0,
		UseAppearance: true,
		Appearance: sdk.HeroAppearanceSpec{
			HairStyle: c.hairStyle,
			IrisColor: eyePalettes[c.eyeHueIdx%len(eyePalettes)],
		},
		CameraFocus: focus,
	})
}

func (c *originsClient) renderCreateOverlay() {
	if c.cfg.Mode != sdk.CharCreateModePluginCustom {
		// In vanilla_extended mode the client keeps native customizer + presets + class cards.
		sdk.ClearOverlay()
		return
	}

	children := []sdk.UINodeSpec{
		sdk.LabelNode("orig_title", "SANCTUARY ORIGINS", 16, [4]float32{0.35, 0.92, 1, 1}, false),
		sdk.LabelNode("orig_sub", fmt.Sprintf("Realm: %s", c.serverName), 12, [4]float32{0.72, 0.80, 0.92, 1}, true),
		sdk.InputNode("origin_hero_name", "Enter Hero Name (3-16 chars)...", "", 16),
		sdk.LabelNode("orig_cls_hdr", "1. Select Starter Class:", 13, [4]float32{0.92, 0.95, 1, 1}, false),
	}

	for _, cls := range c.cfg.Classes {
		variant := "default"
		if cls.ID == c.selectedID {
			variant = "primary"
		}
		children = append(children,
			sdk.ButtonNode("class_"+cls.ID, cls.Name, variant, 290, 32),
		)
	}

	children = append(children,
		sdk.LabelNode("orig_trait_hdr", "2. Select Backstory Perk:", 13, [4]float32{0.92, 0.95, 1, 1}, false),
		sdk.RowNode("trait_row", 6,
			sdk.ButtonNode("trait_wanderer", "Wanderer", traitVariant(c.trait == "wanderer"), 92, 28),
			sdk.ButtonNode("trait_merchant", "Merchant", traitVariant(c.trait == "merchant"), 92, 28),
			sdk.ButtonNode("trait_alchemist", "Alchemist", traitVariant(c.trait == "alchemist"), 94, 28),
		),
		sdk.RowNode("preview_row", 8,
			sdk.ButtonNode("btn_cycle_look", "Cycle Hair/Eyes", "default", 140, 30),
			sdk.ButtonNode("btn_toggle_cam", camLabel(c.faceFocus), "default", 140, 30),
		),
		sdk.ButtonNode("btn_forge_submit", "FORGE & ENTER SANCTUARY", "emerald", 290, 36),
	)

	sdk.SetOverlay(sdk.UIPanelSpec{
		ID:      "custom_origins_panel",
		Visible: true,
		Anchor:  "top_right",
		OffsetX: 24,
		OffsetY: 74,
		Width:   330,
		Root:    ptrNode(sdk.ColNode("origins_col", 8, children...)),
	})
}

func traitVariant(active bool) string {
	if active {
		return "primary"
	}
	return "default"
}

func camLabel(face bool) string {
	if face {
		return "Camera: Full Body"
	}
	return "Camera: Face Zoom"
}

func ptrNode(n sdk.UINodeSpec) *sdk.UINodeSpec { return &n }

func init() {
	sdk.RegisterClient(&originsClient{})
}

func main() {}
