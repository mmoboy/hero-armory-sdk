package main

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
	"realm_bounty/proto"
)

const (
	viewClosed  = ""
	viewBounty  = "bounty"
	viewSandbox = "sandbox"
)

type bountyClient struct {
	sdk.BaseClientPlugin
	hero           sdk.HeroSnapshot
	view           string
	cheerNote      string
	probePath      string
	selectedOath   string
	inCreateScreen bool
	state          proto.BountySnapshot

	clientAudit *proto.SandboxProbeResp
	serverAudit *proto.SandboxProbeResp
	lastStatus  string
}

func newBountyClient() *bountyClient {
	return &bountyClient{
		probePath:    "/",
		selectedOath: "oath_greed",
		state: proto.BountySnapshot{
			Title:       "Chunk Explorer",
			Description: "Explore 3 different rooms around the realm.",
			Progress:    0,
			Target:      3,
			RankTitle:   "Initiate",
			LastEvent:   "Syncing with Realm Bounty Board...",
		},
	}
}

func (c *bountyClient) Init() {
	sdk.Log("Realm Bounty Board client plugin ready")
	registerCustomRealmMeshes()
	c.probePath = sdk.PrefGetString("sandbox.last_path", "/")
	c.cheerNote = sdk.PrefGetString("bounty.last_note", "")
	if savedOath := strings.TrimSpace(sdk.PrefGetString("bounty.selected_oath", "")); savedOath != "" {
		c.selectedOath = savedOath
	}
	c.runClientAudit(c.probePath, "explore")
}

func (c *bountyClient) savePathPref(path string) {
	if p := strings.TrimSpace(path); p != "" {
		_ = sdk.PrefSet("sandbox.last_path", p)
	}
}

func (c *bountyClient) runClientAudit(path, mode string) {
	if strings.TrimSpace(path) != "" {
		c.probePath = strings.TrimSpace(path)
	}
	if c.probePath == "" {
		c.probePath = "/"
	}
	rep := proto.RunSandboxProbes("client", c.probePath, mode)

	// Attempt server-only hostcalls from RoleClient: host must reject them all.
	rcMob := sdk.RegisterMonster(sdk.MonsterSpec{ID: "client_escalate", Name: "Hack Mob", MaxHP: 10})
	rcNotice := sdk.BroadcastNotice("warn", "Hack", "Should be rejected in RoleClient")
	rcItem := sdk.GiveItemToHero(c.hero.ID, "plug_bounty_ore", 99)
	kvErr := sdk.KVSet("client_escalate", []byte("1"))
	blockedHostcall := rcMob < 0 && rcNotice < 0 && rcItem < 0 && kvErr != nil
	detail := fmt.Sprintf("Rejected (mob=%d,notice=%d,item=%d,kv=%v): RoleClient blocked", rcMob, rcNotice, rcItem, kvErr != nil)
	if !blockedHostcall {
		detail = fmt.Sprintf("UNEXPECTED SERVER HOSTCALL SUCCEEDED (%d,%d,%d,kv=%v)", rcMob, rcNotice, rcItem, kvErr)
		rep.AllBlocked = false
	}
	rep.Checks = append(rep.Checks, proto.SandboxProbeCheck{
		Name:    "Server Hostcalls",
		Target:  "RegisterMob/Notice/Give/KVSet",
		Blocked: blockedHostcall,
		Detail:  detail,
	})
	c.clientAudit = &rep
	c.lastStatus = fmt.Sprintf("Probed client VM (%s): %d/%d blocked (Pref saved).", c.probePath, blockedCount(rep.Checks), len(rep.Checks))
}

func (c *bountyClient) requestServerAudit(path, mode string) {
	p := strings.TrimSpace(path)
	if p == "" {
		p = c.probePath
	}
	if p == "" {
		p = "/"
	}
	c.probePath = p
	sdk.SendJSONToServer(proto.ChannelSandbox, proto.SandboxProbeReq{
		Path: p,
		Mode: mode,
	})
}

func blockedCount(checks []proto.SandboxProbeCheck) int {
	n := 0
	for _, ch := range checks {
		if ch.Blocked {
			n++
		}
	}
	return n
}

func (c *bountyClient) allAuditsSafe() bool {
	if c.clientAudit != nil && !c.clientAudit.AllBlocked {
		return false
	}
	if c.serverAudit != nil && !c.serverAudit.AllBlocked {
		return false
	}
	return true
}

func (c *bountyClient) syncHUDButton() {
	badge := fmt.Sprintf("%d/%d", c.state.Progress, c.state.Target)
	if c.state.Target <= 0 {
		badge = ""
	}
	if c.state.Completed {
		badge = "DONE"
	}
	secBadge := "SAFE"
	if !c.allAuditsSafe() {
		secBadge = "LEAK"
	}
	sdk.SetHUDActions([]sdk.HUDActionSpec{
		{
			ID:      "bounty_toggle",
			Tooltip: "Realm Bounty Board",
			Icon:    "assets/bounty_icon.png",
			Badge:   badge,
			Active:  c.view == viewBounty,
		},
		{
			ID:      "sandbox_toggle",
			Tooltip: "Wasm File Explorer & Sandbox Audit",
			Icon:    "folder",
			Badge:   secBadge,
			Active:  c.view == viewSandbox,
		},
	})
}

func (c *bountyClient) syncOverlay() {
	switch c.view {
	case viewBounty:
		c.renderBountyOverlay()
	case viewSandbox:
		c.renderSandboxOverlay()
	default:
		sdk.ClearOverlay()
	}
}

func (c *bountyClient) renderBountyOverlay() {
	pct := float32(0)
	if c.state.Target > 0 {
		pct = float32(c.state.Progress) / float32(c.state.Target)
		if pct > 1 {
			pct = 1
		}
	}
	fillCol := [4]float32{0.22, 0.72, 0.96, 0.96}
	if c.state.Completed {
		fillCol = [4]float32{0.22, 0.84, 0.56, 0.96}
	}

	origin := c.state.OriginName
	if origin == "" {
		origin = "Freelance"
	}
	oath := c.state.OathName
	if oath == "" {
		oath = "None"
	}
	statsLine := fmt.Sprintf("Origin: %s  |  Oath: %s  |  Progress: %d/%d  |  Done: %d (%s)  |  Cheers: %d",
		origin, oath, c.state.Progress, c.state.Target, c.state.CompletedCount, c.state.RankTitle, c.state.RealmCheers)

	sdk.SetOverlay(sdk.UIPanelSpec{
		ID:          "bounty_board",
		Visible:     true,
		Anchor:      "top_right",
		OffsetX:     14,
		OffsetY:     82,
		Width:       380,
		Padding:     12,
		Gap:         8,
		BgColor:     [4]float32{0.05, 0.08, 0.14, 0.95},
		BorderColor: [4]float32{0.92, 0.72, 0.24, 0.88},
		BorderWidth: 1.5,
		Radius:      10.0,
		Root: &sdk.UINodeSpec{
			Kind: "col",
			Gap:  8,
			Children: []sdk.UINodeSpec{
				{
					Kind: "row",
					Gap:  8,
					Children: []sdk.UINodeSpec{
						{Kind: "image", AssetPath: "assets/bounty_icon.png", Width: 24, Height: 24},
						{Kind: "label", Text: "REALM BOUNTY BOARD", TextSize: 13.5, FlexGrow: 1, Color: [4]float32{0.98, 0.84, 0.42, 1.0}},
						{Kind: "button", ID: "bounty_codex", Text: "Codex", Variant: "default", Height: 24, TextSize: 10.0},
						{Kind: "button", ID: "bounty_close", Text: "Close", Variant: "default", Height: 24, TextSize: 10.5},
					},
				},
				{
					Kind:        "panel",
					Padding:     9,
					Gap:         5,
					Color:       [4]float32{0.08, 0.12, 0.20, 0.92},
					BorderColor: [4]float32{0.26, 0.46, 0.72, 0.65},
					Radius:      7.0,
					Children: []sdk.UINodeSpec{
						{Kind: "label", Text: c.state.Title, TextSize: 13.0, Color: [4]float32{0.96, 0.98, 1.0, 1.0}},
						{Kind: "label", Text: c.state.Description, TextSize: 11.0, Wrap: true, Color: [4]float32{0.78, 0.86, 0.96, 1.0}},
						{Kind: "progress", Progress: pct, Height: 12, Color: fillCol},
						{Kind: "label", Text: statsLine, TextSize: 10.5, Wrap: true, Color: [4]float32{0.68, 0.86, 0.96, 1.0}},
					},
				},
				{Kind: "label", Text: "Status: " + c.state.LastEvent, TextSize: 10.5, Wrap: true, Color: [4]float32{0.66, 0.92, 0.74, 1.0}},
				{Kind: "input", ID: "bounty_note", Text: c.cheerNote, Placeholder: "Optional realm cheer shoutout...", Height: 28},
				{
					Kind: "row",
					Gap:  6,
					Children: []sdk.UINodeSpec{
						{Kind: "button", ID: "bounty_reroll", Text: "Reroll Bounty", Variant: "default", FlexGrow: 1, Height: 28},
						{Kind: "button", ID: "bounty_summon", Text: "Summon Stalker", Variant: "primary", FlexGrow: 1, Height: 28},
						{Kind: "button", ID: "bounty_cheer", Text: "Realm Cheer", Variant: "emerald", FlexGrow: 1, Height: 28},
					},
				},
			},
		},
	})
}

func (c *bountyClient) renderSandboxOverlay() {
	if c.clientAudit == nil {
		c.runClientAudit(c.probePath, "explore")
	}

	totalChecks := 0
	totalBlocked := 0
	if c.clientAudit != nil {
		totalChecks += len(c.clientAudit.Checks)
		totalBlocked += blockedCount(c.clientAudit.Checks)
	}
	if c.serverAudit != nil {
		totalChecks += len(c.serverAudit.Checks)
		totalBlocked += blockedCount(c.serverAudit.Checks)
	}

	summaryText := fmt.Sprintf("SANDBOX VERIFIED: %d/%d PROBES BLOCKED (0 host files, 0 sockets, 0 exec)", totalBlocked, totalChecks)
	summaryCol := [4]float32{0.38, 0.96, 0.62, 1.0}
	borderCol := [4]float32{0.24, 0.78, 0.56, 0.88}
	if !c.allAuditsSafe() {
		summaryText = "WARNING: A sandbox probe was NOT blocked!"
		summaryCol = [4]float32{1.0, 0.44, 0.42, 1.0}
		borderCol = [4]float32{0.96, 0.34, 0.32, 0.92}
	}

	var scrollChildren []sdk.UINodeSpec

	if c.clientAudit != nil {
		scrollChildren = append(scrollChildren, sdk.UINodeSpec{
			Kind:         "label",
			Text:         fmt.Sprintf("CLIENT WASM VM (%d/%d blocked, ReadDir entries: %d)", blockedCount(c.clientAudit.Checks), len(c.clientAudit.Checks), c.clientAudit.EntriesFound),
			TextSize:     10.5,
			PercentWidth: 100,
			Color:        [4]float32{0.42, 0.88, 1.0, 1.0},
		})
		scrollChildren = append(scrollChildren, buildAuditCheckCards("cli", c.clientAudit.Checks)...)
	}

	if c.serverAudit != nil {
		scrollChildren = append(scrollChildren, sdk.UINodeSpec{
			Kind:         "label",
			Text:         fmt.Sprintf("SERVER WASM VM (%d/%d blocked, ReadDir entries: %d)", blockedCount(c.serverAudit.Checks), len(c.serverAudit.Checks), c.serverAudit.EntriesFound),
			TextSize:     10.5,
			PercentWidth: 100,
			Color:        [4]float32{0.98, 0.82, 0.40, 1.0},
		})
		scrollChildren = append(scrollChildren, buildAuditCheckCards("srv", c.serverAudit.Checks)...)
	} else {
		scrollChildren = append(scrollChildren, sdk.UINodeSpec{
			Kind:         "label",
			Text:         "SERVER WASM VM: Waiting for server sandbox report (or click Audit Server VM)...",
			TextSize:     10.0,
			Wrap:         true,
			PercentWidth: 100,
			Color:        [4]float32{0.72, 0.80, 0.92, 1.0},
		})
	}

	sdk.SetOverlay(sdk.UIPanelSpec{
		ID:          "sandbox_explorer",
		Visible:     true,
		Anchor:      "top_right",
		OffsetX:     14,
		OffsetY:     82,
		Width:       500,
		MaxHeight:   590,
		Padding:     12,
		Gap:         7,
		BgColor:     [4]float32{0.05, 0.08, 0.14, 0.96},
		BorderColor: borderCol,
		BorderWidth: 1.5,
		Radius:      10.0,
		Root: &sdk.UINodeSpec{
			Kind: "col",
			Gap:  7,
			Children: []sdk.UINodeSpec{
				{
					Kind: "row",
					Gap:  6,
					Children: []sdk.UINodeSpec{
						{Kind: "icon", Icon: "folder", Width: 20, Height: 20},
						{Kind: "label", Text: "WASM FILE EXPLORER & SECURITY", TextSize: 13.0, FlexGrow: 1, Color: [4]float32{0.44, 0.94, 0.72, 1.0}},
						{Kind: "button", ID: "sandbox_close", Text: "Close", Variant: "default", Height: 24, TextSize: 10.0},
					},
				},
				{
					Kind:         "panel",
					Padding:      7,
					Gap:          4,
					PercentWidth: 100,
					Color:        [4]float32{0.06, 0.11, 0.16, 0.94},
					BorderColor:  borderCol,
					Radius:       6.0,
					Children: []sdk.UINodeSpec{
						{Kind: "label", ID: "sandbox_summary", Text: summaryText, TextSize: 11.0, Wrap: true, PercentWidth: 100, Color: summaryCol},
						{Kind: "label", Text: "Status: " + c.lastStatus, TextSize: 9.5, Wrap: true, PercentWidth: 100, Color: [4]float32{0.78, 0.88, 0.96, 1.0}},
					},
				},
				{
					Kind: "row",
					Gap:  5,
					Children: []sdk.UINodeSpec{
						{Kind: "input", ID: "sandbox_path", Text: c.probePath, Placeholder: "Host path (e.g. /, /etc/passwd, /home)...", FlexGrow: 1, Height: 28},
						{Kind: "button", ID: "sandbox_explore", Text: "ReadDir", Variant: "primary", Height: 28, TextSize: 10.0},
						{Kind: "button", ID: "sandbox_read", Text: "ReadFile", Variant: "default", Height: 28, TextSize: 10.0},
						{Kind: "button", ID: "sandbox_write", Text: "WriteTest", Variant: "danger", Height: 28, TextSize: 10.0},
					},
				},
				{
					Kind: "row",
					Gap:  4,
					Children: []sdk.UINodeSpec{
						{Kind: "button", ID: "sandbox_preset_root", Text: "/", Variant: "default", FlexGrow: 1, Height: 23, TextSize: 9.5},
						{Kind: "button", ID: "sandbox_preset_passwd", Text: "/etc/passwd", Variant: "default", FlexGrow: 1, Height: 23, TextSize: 9.5},
						{Kind: "button", ID: "sandbox_preset_home", Text: "/home", Variant: "default", FlexGrow: 1, Height: 23, TextSize: 9.5},
						{Kind: "button", ID: "sandbox_preset_cwd", Text: ".", Variant: "default", FlexGrow: 1, Height: 23, TextSize: 9.5},
						{Kind: "button", ID: "sandbox_preset_mod", Text: "../../go.mod", Variant: "default", FlexGrow: 1, Height: 23, TextSize: 9.5},
					},
				},
				{
					Kind: "row",
					Gap:  6,
					Children: []sdk.UINodeSpec{
						{Kind: "button", ID: "sandbox_audit_client", Text: "Re-Audit Client VM", Variant: "emerald", FlexGrow: 1, Height: 26, TextSize: 10.0},
						{Kind: "button", ID: "sandbox_audit_server", Text: "Audit Server VM", Variant: "active", FlexGrow: 1, Height: 26, TextSize: 10.0},
					},
				},
				{
					Kind:     "scroll",
					ID:       "sandbox_scroll",
					Height:   270,
					Gap:      5,
					Children: scrollChildren,
				},
			},
		},
	})
}

func buildAuditCheckCards(prefix string, checks []proto.SandboxProbeCheck) []sdk.UINodeSpec {
	cards := make([]sdk.UINodeSpec, 0, len(checks))
	for i, ch := range checks {
		status := "[BLOCKED]"
		stCol := [4]float32{0.38, 0.96, 0.62, 1.0}
		bCol := [4]float32{0.20, 0.46, 0.44, 0.72}
		if !ch.Blocked {
			status = "[LEAK!]"
			stCol = [4]float32{1.0, 0.42, 0.40, 1.0}
			bCol = [4]float32{0.92, 0.30, 0.30, 0.90}
		}
		cards = append(cards, sdk.UINodeSpec{
			Kind:         "panel",
			ID:           fmt.Sprintf("sbx_%s_%d", prefix, i),
			Padding:      6,
			Gap:          2,
			PercentWidth: 100,
			Color:        [4]float32{0.07, 0.10, 0.17, 0.92},
			BorderColor:  bCol,
			Radius:       5.0,
			Children: []sdk.UINodeSpec{
				{
					Kind:         "label",
					Text:         fmt.Sprintf("%s %s  (%s)", status, ch.Name, ch.Target),
					TextSize:     9.8,
					Wrap:         true,
					PercentWidth: 100,
					Color:        stCol,
				},
				{
					Kind:         "label",
					Text:         ch.Detail,
					TextSize:     9.0,
					Wrap:         true,
					PercentWidth: 100,
					Color:        [4]float32{0.76, 0.86, 0.96, 0.95},
				},
			},
		})
	}
	return cards
}

func (c *bountyClient) renderGuildOathOverlay() {
	oathDesc := "Earn +1 bonus Gilded Ore on hero initiation and +1 extra Gilded Ore on every completed bounty."
	switch c.selectedOath {
	case "oath_hunt":
		oathDesc = "Begin your hunt with an extra Stalker Draft tonic to challenge the Gilded Stalker immediately."
	case "oath_vanguard":
		oathDesc = "Receive +2 Red Heart Elixirs in your starter satchel for sustained frontline dungeon delves."
	}
	varFor := func(id string) string {
		if c.selectedOath == id {
			return "emerald"
		}
		return "default"
	}

	sdk.SetOverlay(sdk.UIPanelSpec{
		ID:          "bounty_oath_panel",
		Visible:     true,
		Anchor:      "top_left",
		OffsetX:     16,
		OffsetY:     82,
		Width:       292,
		Padding:     11,
		Gap:         7,
		BgColor:     [4]float32{0.05, 0.08, 0.14, 0.94},
		BorderColor: [4]float32{0.92, 0.72, 0.24, 0.86},
		BorderWidth: 1.5,
		Radius:      10.0,
		Root: &sdk.UINodeSpec{
			Kind: "col",
			Gap:  7,
			Children: []sdk.UINodeSpec{
				{
					Kind: "row",
					Gap:  7,
					Children: []sdk.UINodeSpec{
						{Kind: "image", AssetPath: "assets/bounty_icon.png", Width: 22, Height: 22},
						{Kind: "label", Text: "BOUNTY GUILD OATH", TextSize: 12.8, FlexGrow: 1, Color: [4]float32{0.98, 0.86, 0.44, 1.0}},
					},
				},
				{
					Kind:     "label",
					Text:     "Swear a Guild Oath alongside your Origin. Tip: Pick 'Gilded Automaton' on the right to preview the custom 3D Wakfu construct!",
					TextSize: 10.2,
					Wrap:     true,
					Color:    [4]float32{0.76, 0.86, 0.96, 1.0},
				},
				{Kind: "button", ID: "bounty_oath_greed", Text: "Oath of Greed (+Bonus Ore)", Variant: varFor("oath_greed"), Height: 28, TextSize: 10.8},
				{Kind: "button", ID: "bounty_oath_hunt", Text: "Oath of the Hunt (+Stalker Draft)", Variant: varFor("oath_hunt"), Height: 28, TextSize: 10.8},
				{Kind: "button", ID: "bounty_oath_vanguard", Text: "Oath of the Vanguard (+2 Elixirs)", Variant: varFor("oath_vanguard"), Height: 28, TextSize: 10.8},
				{
					Kind:        "panel",
					Padding:     8,
					Color:       [4]float32{0.08, 0.12, 0.20, 0.92},
					BorderColor: [4]float32{0.26, 0.56, 0.78, 0.65},
					Radius:      6.0,
					Children: []sdk.UINodeSpec{
						{Kind: "label", Text: oathDesc, TextSize: 10.2, Wrap: true, Color: [4]float32{0.66, 0.94, 0.78, 1.0}},
					},
				},
			},
		},
	})
}

func (c *bountyClient) OnEnterCharacterCreate(ctx sdk.CharacterCreateContext) {
	c.inCreateScreen = true
	if c.selectedOath == "" {
		c.selectedOath = "oath_greed"
	}
	c.renderGuildOathOverlay()
}

func (c *bountyClient) OnEnterWorld(self sdk.HeroSnapshot) {
	c.inCreateScreen = false
	c.hero = self
	c.view = viewClosed
	sdk.ClearOverlay()
	c.runClientAudit(c.probePath, "explore")
	c.syncHUDButton()
	if c.selectedOath != "" {
		sdk.SendJSONToServer(proto.ChannelBounty, proto.ClientAction{
			Action: "claim_oath",
			Note:   c.selectedOath,
		})
	} else {
		sdk.SendJSONToServer(proto.ChannelBounty, proto.ClientAction{Action: "sync"})
	}
	c.requestServerAudit(c.probePath, "explore")
}

func (c *bountyClient) OnPacket(channel uint16, data []byte) {
	switch channel {
	case proto.ChannelBounty:
		var snap proto.BountySnapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			return
		}
		c.state = snap
		c.syncHUDButton()
		if c.view == viewBounty {
			c.syncOverlay()
		}

	case proto.ChannelSandbox:
		var resp proto.SandboxProbeResp
		if err := json.Unmarshal(data, &resp); err != nil {
			return
		}
		c.serverAudit = &resp
		c.lastStatus = fmt.Sprintf("Server VM probe (%s): %d/%d blocked.", resp.Path, blockedCount(resp.Checks), len(resp.Checks))
		c.syncHUDButton()
		if c.view == viewSandbox {
			c.syncOverlay()
		}
	}
}

var sandboxPresets = map[string]struct {
	path string
	mode string
}{
	"sandbox_preset_root":   {path: "/", mode: "explore"},
	"sandbox_preset_passwd": {path: "/etc/passwd", mode: "read"},
	"sandbox_preset_home":   {path: "/home", mode: "explore"},
	"sandbox_preset_cwd":    {path: ".", mode: "explore"},
	"sandbox_preset_mod":    {path: "../../go.mod", mode: "read"},
}

func (c *bountyClient) runAndPersistSandbox(path, mode string, includeServer bool) {
	c.runClientAudit(path, mode)
	c.savePathPref(c.probePath)
	if includeServer {
		c.requestServerAudit(c.probePath, mode)
	}
	c.syncHUDButton()
	c.syncOverlay()
}

func (c *bountyClient) OnUIEvent(ev sdk.UIClickEvent) {
	if ev.Inputs != nil {
		if note, ok := ev.Inputs["bounty_note"]; ok {
			c.cheerNote = note
		}
		if p, ok := ev.Inputs["sandbox_path"]; ok && strings.TrimSpace(p) != "" {
			c.probePath = strings.TrimSpace(p)
		}
	}

	if preset, ok := sandboxPresets[ev.ControlID]; ok {
		c.runAndPersistSandbox(preset.path, preset.mode, true)
		return
	}

	switch ev.ControlID {
	case "bounty_oath_greed", "bounty_oath_hunt", "bounty_oath_vanguard":
		c.selectedOath = strings.TrimPrefix(ev.ControlID, "bounty_")
		_ = sdk.PrefSet("bounty.selected_oath", c.selectedOath)
		if c.inCreateScreen {
			c.renderGuildOathOverlay()
		}
	case "bounty_toggle":
		if c.view == viewBounty {
			c.view = viewClosed
		} else {
			c.view = viewBounty
		}
		c.syncHUDButton()
		c.syncOverlay()
	case "bounty_close":
		c.view = viewClosed
		c.syncHUDButton()
		c.syncOverlay()
	case "bounty_codex":
		sdk.OpenCodex("Gilded")
	case "bounty_note":
		c.cheerNote = ev.Value
		if strings.TrimSpace(c.cheerNote) == "" {
			_ = sdk.PrefDelete("bounty.last_note")
		} else {
			_ = sdk.PrefSet("bounty.last_note", c.cheerNote)
		}
	case "bounty_reroll":
		sdk.SendJSONToServer(proto.ChannelBounty, proto.ClientAction{Action: "reroll"})
	case "bounty_summon":
		sdk.SendJSONToServer(proto.ChannelBounty, proto.ClientAction{Action: "summon_boss"})
	case "bounty_cheer":
		if ev.Value != "" {
			c.cheerNote = ev.Value
		}
		if strings.TrimSpace(c.cheerNote) == "" {
			_ = sdk.PrefDelete("bounty.last_note")
		} else {
			_ = sdk.PrefSet("bounty.last_note", c.cheerNote)
		}
		sdk.SendJSONToServer(proto.ChannelBounty, proto.ClientAction{
			Action: "cheer",
			Note:   c.cheerNote,
		})

	case "sandbox_toggle":
		if c.view == viewSandbox {
			c.view = viewClosed
			c.syncHUDButton()
			c.syncOverlay()
		} else {
			c.view = viewSandbox
			c.runAndPersistSandbox(c.probePath, "explore", true)
		}
	case "sandbox_close":
		c.view = viewClosed
		c.syncHUDButton()
		c.syncOverlay()
	case "sandbox_path":
		if strings.TrimSpace(ev.Value) != "" {
			c.probePath = strings.TrimSpace(ev.Value)
		}
		c.runAndPersistSandbox(c.probePath, "explore", true)
	case "sandbox_explore", "sandbox_audit_client":
		c.runAndPersistSandbox(c.probePath, "explore", false)
	case "sandbox_read":
		c.runAndPersistSandbox(c.probePath, "read", true)
	case "sandbox_write":
		c.runAndPersistSandbox(c.probePath, "write", true)
	case "sandbox_audit_server":
		c.savePathPref(c.probePath)
		c.requestServerAudit(c.probePath, "explore")
		c.lastStatus = "Sent sandbox audit request to Server Wasm VM..."
		c.syncOverlay()
	}
}

func init() {
	sdk.RegisterClient(newBountyClient())
}

func main() {}
