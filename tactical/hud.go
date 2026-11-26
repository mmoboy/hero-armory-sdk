package tactical

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

const (
	// DefaultHUDChannel is the default OpPluginPacket channel used by
	// tactical.Engine (RoleServer) to push BattleHUDSnapshot updates to
	// HandleHUDPacket (RoleClient).
	DefaultHUDChannel uint16 = 5050

	// BattleHUDOverlayID is the named client overlay ID mounted at top_center.
	BattleHUDOverlayID = "tactical_hud"

	// BattleTimelineOverlayID is the named client overlay ID mounted on the right
	// showing ordered turn combatants, HP/Shield, active turn highlight, and statuses.
	BattleTimelineOverlayID = "tactical_timeline"

	// EndTurnButtonID is the clickable End Turn button inside tactical_hud.
	EndTurnButtonID = "tactical_end_turn"

	// BattleSummaryOverlayID is the centered Victory/Defeat Loot Summary overlay.
	BattleSummaryOverlayID = "tactical_summary"

	// SummaryCloseButtonID is the clickable Continue button inside tactical_summary.
	SummaryCloseButtonID = "tactical_summary_close"
)

// LootDrop describes one item awarded at the end of a tactical battle.
type LootDrop struct {
	ItemID string `json:"item_id"`
	Name   string `json:"name"`
	Qty    int    `json:"qty"`
}

// BattleSummarySnapshot is sent on DefaultHUDChannel when a battle concludes.
type BattleSummarySnapshot struct {
	Won        bool       `json:"won"`
	EnemyName  string     `json:"enemy_name"`
	Difficulty string     `json:"difficulty,omitempty"`
	Turns      int        `json:"turns"`
	Gold       int64      `json:"gold,omitempty"`
	XP         int64      `json:"xp,omitempty"`
	Loot       []LootDrop `json:"loot,omitempty"`
}

// BattleHUDAction is sent from client to server on DefaultHUDChannel when the
// player clicks a tactical_hud action button (such as End Turn).
type BattleHUDAction struct {
	Action string `json:"action"`
}

// TimelineEntry describes one combatant card in the tactical Battle Timeline.
type TimelineEntry struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Team     string   `json:"team"` // "ally" or "enemy"
	HP       int      `json:"hp"`
	MaxHP    int      `json:"max_hp"`
	Shield   int      `json:"shield,omitempty"`
	Active   bool     `json:"active,omitempty"`
	Defeated bool     `json:"defeated,omitempty"`
	Statuses []string `json:"statuses,omitempty"`
}

// BattleHUDSnapshot is sent by tactical.Engine to the hero's client plugin
// whenever combat HUD, Battle Timeline, or end-of-battle Loot Summary state changes.
type BattleHUDSnapshot struct {
	Kind       string                 `json:"kind,omitempty"` // "" / "hud" or "summary"
	Active     bool                   `json:"active"`
	Round      int                    `json:"round,omitempty"`
	PlayerTurn bool                   `json:"player_turn,omitempty"`
	TurnLabel  string                 `json:"turn_label,omitempty"`
	HeroHP     int                    `json:"hero_hp,omitempty"`
	HeroMaxHP  int                    `json:"hero_max_hp,omitempty"`
	HeroShield int                    `json:"hero_shield,omitempty"`
	HeroAP     int                    `json:"hero_ap,omitempty"`
	HeroMaxAP  int                    `json:"hero_max_ap,omitempty"`
	HeroMP     float32                `json:"hero_mp,omitempty"`
	HeroMaxMP  float32                `json:"hero_max_mp,omitempty"`
	Statuses   []string               `json:"statuses,omitempty"`
	Prompt     string                 `json:"prompt,omitempty"`
	EnemyName  string                 `json:"enemy_name,omitempty"`
	Tier       string                 `json:"tier,omitempty"`
	Timeline   []TimelineEntry        `json:"timeline,omitempty"`
	Summary    *BattleSummarySnapshot `json:"summary,omitempty"`
}

// BuildBattleHUDOverlay builds the top-center tactical battle HUD UIPanelSpec
// for snap.
func BuildBattleHUDOverlay(snap BattleHUDSnapshot) sdk.UIPanelSpec {
	turnColor := [4]float32{0.34, 0.94, 0.62, 1.0}
	if !snap.PlayerTurn {
		turnColor = [4]float32{1.0, 0.52, 0.34, 1.0}
	}
	hpText := fmt.Sprintf("HP %d/%d", snap.HeroHP, snap.HeroMaxHP)
	if snap.HeroShield > 0 {
		hpText = fmt.Sprintf("HP %d/%d (+%d)", snap.HeroHP, snap.HeroMaxHP, snap.HeroShield)
	}
	statsLine := fmt.Sprintf("%s  |  AP %d/%d  |  MP %.1f/%.1f",
		hpText, snap.HeroAP, snap.HeroMaxAP, snap.HeroMP, snap.HeroMaxMP)

	headerChildren := []sdk.UINodeSpec{
		sdk.LabelNode("tac_round", fmt.Sprintf("ROUND %d", snap.Round), 13, [4]float32{0.96, 0.84, 0.42, 1}, false),
		sdk.LabelNode("tac_turn", snap.TurnLabel, 13, turnColor, false),
	}
	if snap.PlayerTurn {
		headerChildren = append(headerChildren, sdk.ButtonNode(EndTurnButtonID, "End Turn (Space)", "active", 142, 24))
	}
	headerRow := sdk.RowNode("tac_hdr", 8, headerChildren...)
	children := []sdk.UINodeSpec{
		headerRow,
		sdk.LabelNode("tac_stats", statsLine, 12, [4]float32{0.90, 0.94, 1.0, 1}, false),
	}
	if len(snap.Statuses) > 0 {
		children = append(children, sdk.LabelNode(
			"tac_statuses",
			"Status: "+strings.Join(snap.Statuses, ", "),
			11,
			[4]float32{0.62, 0.88, 1.0, 1},
			false,
		))
	}
	if snap.Prompt != "" {
		children = append(children, sdk.LabelNode("tac_prompt", snap.Prompt, 11, [4]float32{0.78, 0.84, 0.94, 1}, false))
	}

	root := sdk.ColNode("tac_root", 4, children...)
	return sdk.UIPanelSpec{
		ID:          BattleHUDOverlayID,
		Visible:     true,
		Anchor:      "top_center",
		OffsetY:     12,
		Width:       400,
		Padding:     10,
		Gap:         4,
		BgColor:     [4]float32{0.07, 0.09, 0.14, 0.92},
		BorderColor: [4]float32{0.28, 0.78, 0.96, 0.85},
		BorderWidth: 1.5,
		Radius:      8,
		Root:        &root,
	}
}

// BuildBattleTimelineOverlay builds the right-side Battle Timeline panel
// showing turn order, active combatant highlight, HP/Shield bars, and statuses.
func BuildBattleTimelineOverlay(snap BattleHUDSnapshot) sdk.UIPanelSpec {
	children := make([]sdk.UINodeSpec, 0, len(snap.Timeline)+2)
	title := fmt.Sprintf("TURN TIMELINE (R%d)", snap.Round)
	children = append(children, sdk.LabelNode("tl_title", title, 12, [4]float32{0.96, 0.84, 0.42, 1}, false))

	maxEntries := len(snap.Timeline)
	if maxEntries > 6 {
		maxEntries = 6
	}
	for i := 0; i < maxEntries; i++ {
		e := snap.Timeline[i]
		prefix := "  "
		if e.Defeated {
			prefix = "[KO] "
		} else if e.Active {
			prefix = "> "
		}

		nameCol := [4]float32{0.96, 0.54, 0.46, 1.0}
		barCol := [4]float32{0.88, 0.30, 0.26, 1.0}
		cardBg := [4]float32{0.11, 0.12, 0.18, 0.90}
		cardBorder := [4]float32{0.34, 0.24, 0.30, 0.75}
		if e.Team == "ally" {
			nameCol = [4]float32{0.44, 0.94, 0.72, 1.0}
			barCol = [4]float32{0.26, 0.84, 0.56, 1.0}
			cardBorder = [4]float32{0.22, 0.58, 0.48, 0.80}
		}
		if e.Active && !e.Defeated {
			cardBg = [4]float32{0.14, 0.18, 0.26, 0.96}
			cardBorder = [4]float32{0.96, 0.82, 0.34, 0.98}
		}
		if e.Defeated {
			nameCol = [4]float32{0.54, 0.56, 0.62, 0.85}
			barCol = [4]float32{0.34, 0.36, 0.40, 0.80}
			cardBg = [4]float32{0.08, 0.09, 0.12, 0.72}
			cardBorder = [4]float32{0.24, 0.26, 0.32, 0.55}
		}

		hpStr := fmt.Sprintf("HP %d/%d", e.HP, e.MaxHP)
		if e.Shield > 0 && !e.Defeated {
			hpStr = fmt.Sprintf("HP %d/%d (+%d)", e.HP, e.MaxHP, e.Shield)
		}
		pct := float32(0)
		if e.MaxHP > 0 && e.HP > 0 {
			pct = float32(e.HP) / float32(e.MaxHP)
			if pct > 1 {
				pct = 1
			}
		}

		cardKids := []sdk.UINodeSpec{
			sdk.LabelNode(fmt.Sprintf("tl_n_%d", i), prefix+e.Name, 11, nameCol, false),
			sdk.ProgressNode(fmt.Sprintf("tl_b_%d", i), pct, 12, barCol),
			sdk.LabelNode(fmt.Sprintf("tl_h_%d", i), hpStr, 10, [4]float32{0.84, 0.88, 0.96, 1}, false),
		}
		if len(e.Statuses) > 0 && !e.Defeated {
			cardKids = append(cardKids, sdk.LabelNode(
				fmt.Sprintf("tl_s_%d", i),
				strings.Join(e.Statuses, " | "),
				10,
				[4]float32{0.64, 0.88, 1.0, 1},
				false,
			))
		}
		children = append(children, sdk.CardNode(fmt.Sprintf("tl_c_%d", i), 6, 3, cardBg, cardBorder, cardKids...))
	}

	root := sdk.ColNode("tl_root", 5, children...)
	return sdk.UIPanelSpec{
		ID:          BattleTimelineOverlayID,
		Visible:     true,
		Anchor:      "top_right",
		OffsetX:     16,
		OffsetY:     120,
		Width:       236,
		Padding:     8,
		Gap:         5,
		BgColor:     [4]float32{0.07, 0.09, 0.14, 0.90},
		BorderColor: [4]float32{0.34, 0.74, 0.96, 0.82},
		BorderWidth: 1.5,
		Radius:      8,
		Root:        &root,
	}
}

// BuildBattleSummaryOverlay builds the centered Victory / Defeat Loot Summary
// overlay panel shown at the end of a tactical battle.
func BuildBattleSummaryOverlay(sum BattleSummarySnapshot) sdk.UIPanelSpec {
	titleText := "VICTORY!"
	titleCol := [4]float32{0.38, 0.96, 0.58, 1.0}
	borderCol := [4]float32{0.96, 0.80, 0.30, 0.95}
	if !sum.Won {
		titleText = "DEFEAT"
		titleCol = [4]float32{1.0, 0.36, 0.30, 1.0}
		borderCol = [4]float32{0.92, 0.30, 0.26, 0.92}
	}
	enemy := sum.EnemyName
	if enemy == "" {
		enemy = "Encounter"
	}
	tier := strings.ToUpper(sum.Difficulty)
	if tier == "" {
		tier = "NORMAL"
	}
	sub := fmt.Sprintf("%s  |  %s  |  %d turns", enemy, tier, sum.Turns)

	children := []sdk.UINodeSpec{
		sdk.LabelNode("sum_title", titleText, 20, titleCol, false),
		sdk.LabelNode("sum_sub", sub, 12, [4]float32{0.82, 0.88, 0.96, 1}, false),
	}

	if sum.Won {
		rewKids := []sdk.UINodeSpec{
			sdk.LabelNode("sum_rew_hdr", "REWARDS", 11, [4]float32{0.96, 0.84, 0.42, 1}, false),
			sdk.LabelNode("sum_currency", fmt.Sprintf("Gold: +%dg    XP: +%d XP", sum.Gold, sum.XP), 13, [4]float32{0.98, 0.88, 0.48, 1}, false),
		}
		if len(sum.Loot) == 0 {
			rewKids = append(rewKids, sdk.LabelNode("sum_no_loot", "No item drops this battle.", 11, [4]float32{0.68, 0.74, 0.82, 1}, false))
		} else {
			maxDrops := len(sum.Loot)
			if maxDrops > 10 {
				maxDrops = 10
			}
			for i := 0; i < maxDrops; i++ {
				d := sum.Loot[i]
				name := d.Name
				if name == "" {
					name = d.ItemID
				}
				rewKids = append(rewKids, sdk.LabelNode(
					fmt.Sprintf("sum_loot_%d", i),
					fmt.Sprintf("- %s x%d", name, d.Qty),
					12,
					[4]float32{0.64, 0.94, 0.82, 1},
					false,
				))
			}
		}
		children = append(children, sdk.CardNode(
			"sum_rew_card", 10, 4,
			[4]float32{0.10, 0.13, 0.20, 0.95},
			[4]float32{0.34, 0.56, 0.78, 0.75},
			rewKids...,
		))
	} else {
		children = append(children, sdk.CardNode(
			"sum_defeat_card", 10, 4,
			[4]float32{0.14, 0.09, 0.12, 0.94},
			[4]float32{0.72, 0.28, 0.28, 0.75},
			sdk.LabelNode("sum_defeat_msg", "Regroup, adjust your strategy or /difficulty, and try again.", 12, [4]float32{0.94, 0.78, 0.78, 1}, true),
		))
	}

	children = append(children, sdk.RowNode("sum_btn_row", 8,
		sdk.ButtonNode(SummaryCloseButtonID, "Continue", "active", 160, 30),
	))

	root := sdk.ColNode("sum_root", 8, children...)
	return sdk.UIPanelSpec{
		ID:          BattleSummaryOverlayID,
		Visible:     true,
		Anchor:      "center",
		Width:       420,
		Padding:     14,
		Gap:         8,
		BgColor:     [4]float32{0.07, 0.09, 0.14, 0.96},
		BorderColor: borderCol,
		BorderWidth: 2.0,
		Radius:      10,
		Root:        &root,
	}
}

// HandleHUDPacket decodes a BattleHUDSnapshot on DefaultHUDChannel (RoleClient)
// and mounts or clears `tactical_hud`, `tactical_timeline`, and `tactical_summary`.
// Returns true when channel matched DefaultHUDChannel.
func HandleHUDPacket(channel uint16, data []byte) bool {
	return HandleHUDPacketOnChannel(DefaultHUDChannel, channel, data)
}

// HandleHUDPacketOnChannel is like HandleHUDPacket for a custom packet channel.
func HandleHUDPacketOnChannel(expectedChannel, channel uint16, data []byte) bool {
	if channel != expectedChannel {
		return false
	}
	var snap BattleHUDSnapshot
	if json.Unmarshal(data, &snap) != nil {
		return true
	}
	if snap.Kind == "summary" && snap.Summary != nil {
		_ = sdk.ClearOverlayID(BattleHUDOverlayID)
		_ = sdk.ClearOverlayID(BattleTimelineOverlayID)
		_ = sdk.SetOverlay(BuildBattleSummaryOverlay(*snap.Summary))
		return true
	}
	if !snap.Active {
		_ = sdk.ClearOverlayID(BattleHUDOverlayID)
		_ = sdk.ClearOverlayID(BattleTimelineOverlayID)
		_ = sdk.ClearOverlayID(BattleSummaryOverlayID)
		return true
	}
	_ = sdk.ClearOverlayID(BattleSummaryOverlayID)
	_ = sdk.SetOverlay(BuildBattleHUDOverlay(snap))
	if len(snap.Timeline) > 0 {
		_ = sdk.SetOverlay(BuildBattleTimelineOverlay(snap))
	} else {
		_ = sdk.ClearOverlayID(BattleTimelineOverlayID)
	}
	return true
}

// HandleUIEvent handles clicks on tactical_hud and tactical_summary controls
// (such as End Turn and Continue) inside ClientPlugin.OnUIEvent.
// Returns true when handled.
func HandleUIEvent(ev sdk.UIClickEvent) bool {
	return HandleUIEventOnChannel(DefaultHUDChannel, ev)
}

// HandleUIEventOnChannel is like HandleUIEvent for a custom packet channel.
func HandleUIEventOnChannel(channel uint16, ev sdk.UIClickEvent) bool {
	switch ev.ControlID {
	case EndTurnButtonID:
		_ = sdk.SendJSONToServer(channel, BattleHUDAction{Action: EndTurnActionID})
		return true
	case SummaryCloseButtonID:
		_ = sdk.ClearOverlayID(BattleSummaryOverlayID)
		return true
	default:
		return false
	}
}
