package vanilla

import (
	"fmt"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

func npcGearWithWeapon(base map[string]string, weaponID string) map[string]string {
	out := make(map[string]string, len(base))
	for k, v := range base {
		out[k] = v
	}
	out["weapon"] = weaponID
	return out
}

// NPCs returns all 7 classic Sanctuary NPCs.
func NPCs() []sdk.NPCSpec {
	return []sdk.NPCSpec{
		{
			ID: "npc_steward_aldric", Name: "Steward Aldric", Role: "Sanctuary Guide",
			RoomX: 0, RoomY: 0, PosX: -3.0, PosZ: -2.0, FacingYaw: 0.92, Pose: "heroic",
			Outfit:    npcGearWithWeapon(KokiriGear(), "weapon_guardian_halberd"),
			HairColor: [3]float32{0.82, 0.85, 0.90}, EyeColor: [3]float32{0.25, 0.85, 0.98},
			SkinRGB: [3]float32{0.94, 0.77, 0.64}, AccentRGB: [3]float32{0.18, 0.36, 0.56},
			HairStyle: 2, FacialHair: 1,
			Greeting: "Welcome to Sanctuary Courtyard! Walk onto any glowing cardinal portal pad to explore the 25 realm zones, or press [M] to open the World Map.",
			Options: []sdk.NPCDialogOptionSpec{
				{ID: "lore_sanctuary", Label: "Tell me about the portals", Action: "lore"},
				{ID: "lore_scouts", Label: "Who are the roaming scouts?", Action: "lore"},
				{ID: "emote_heroic", Label: "Salute with Aldric", Action: "emote"},
			},
		},
		{
			ID: "npc_armorer_lyra", Name: "Forgekeeper Lyra", Role: "Master Armorer",
			RoomX: 0, RoomY: 0, PosX: 3.0, PosZ: -2.0, FacingYaw: -0.92, Pose: "idle",
			Outfit:    npcGearWithWeapon(TwilightGear(), "weapon_xelor_staff"),
			HairColor: [3]float32{0.94, 0.42, 0.24}, EyeColor: [3]float32{0.96, 0.74, 0.22},
			SkinRGB: [3]float32{0.97, 0.80, 0.68}, AccentRGB: [3]float32{0.26, 0.20, 0.38},
			HairStyle: 4, FacialHair: 0,
			Greeting: "Need a fresh battlefield silhouette? Pick a relic blessing below and your Armory gear will synchronize across the room immediately.",
			Options: []sdk.NPCDialogOptionSpec{
				{ID: "equip_kokiri", Label: "Equip Kokiri Guardian Set", Action: "equip_set"},
				{ID: "equip_gobball", Label: "Equip Gobball Warchief Set", Action: "equip_set"},
				{ID: "equip_wakfu", Label: "Equip Wakfu Chronomancer Set", Action: "equip_set"},
				{ID: "equip_twilight", Label: "Equip Twilight Sanctuary Set", Action: "equip_set"},
			},
		},
		{
			ID: "npc_merchant_orion", Name: "Merchant Orion", Role: "Grand Merchant",
			RoomX: 0, RoomY: 0, PosX: 0.0, PosZ: -4.6, FacingYaw: 0.0, Pose: "heroic",
			Outfit:    npcGearWithWeapon(WakfuGear(), "weapon_xelor_staff"),
			HairColor: [3]float32{0.96, 0.80, 0.28}, EyeColor: [3]float32{0.24, 0.92, 1.00},
			SkinRGB: [3]float32{0.95, 0.78, 0.65}, AccentRGB: [3]float32{0.18, 0.28, 0.54},
			HairStyle: 3, FacialHair: 0,
			ShopEnabled: true,
			Greeting:    "Looking for gear, elixirs, materials, or Black Hole Portal Styles? I carry every item in the realm in exchange for Gold.",
			Options: []sdk.NPCDialogOptionSpec{
				{ID: "open_merchant", Label: "Browse All Game Items (Shop)", Action: "shop", OpenMerchant: true},
				{ID: "lore_merchant", Label: "How do I use Gold & Portal Styles?", Action: "lore"},
			},
		},
		{
			ID: "npc_botanist_sylvie", Name: "Botanist Sylvie", Role: "Glade Keeper",
			RoomX: 0, RoomY: 1, PosX: -2.6, PosZ: -2.2, FacingYaw: 0.75, Pose: "channel",
			Outfit:    npcGearWithWeapon(KokiriGear(), "weapon_master_blade"),
			HairColor: [3]float32{0.22, 0.76, 0.36}, EyeColor: [3]float32{0.30, 0.94, 0.48},
			SkinRGB: [3]float32{0.95, 0.80, 0.67}, AccentRGB: [3]float32{0.16, 0.42, 0.24},
			HairStyle: 1, FacialHair: 0,
			Greeting: "The Emerald Glade hums cleanly when visitors arrive through the northern gate.",
			Options: []sdk.NPCDialogOptionSpec{
				{ID: "equip_kokiri", Label: "Receive Kokiri Glade Blessing", Action: "equip_set"},
				{ID: "emote_channel", Label: "Meditate with the Glade", Action: "emote"},
				{ID: "lore_room", Label: "Ask about this chunk", Action: "lore"},
			},
		},
		{
			ID: "npc_forgemaster_orion", Name: "Master Orion", Role: "Chrono-Smith",
			RoomX: 1, RoomY: 0, PosX: 2.8, PosZ: -2.2, FacingYaw: -0.85, Pose: "idle",
			Outfit:    npcGearWithWeapon(WakfuGear(), "weapon_xelor_staff"),
			HairColor: [3]float32{0.22, 0.68, 0.96}, EyeColor: [3]float32{0.18, 0.88, 1.00},
			SkinRGB: [3]float32{0.92, 0.76, 0.66}, AccentRGB: [3]float32{0.14, 0.24, 0.44},
			HairStyle: 3, FacialHair: 2,
			Greeting: "Every gear change is forged at tick speed in the Astral Forge Hall.",
			Options: []sdk.NPCDialogOptionSpec{
				{ID: "equip_wakfu", Label: "Attune Wakfu Chronomancer Set", Action: "equip_set"},
				{ID: "emote_channel", Label: "Channel Chrono Resonance", Action: "emote"},
				{ID: "lore_room", Label: "Ask about the Forge", Action: "lore"},
			},
		},
		{
			ID: "npc_champion_bram", Name: "Champion Bram", Role: "Arena Brawler",
			RoomX: -1, RoomY: 0, PosX: -2.6, PosZ: 2.2, FacingYaw: 2.35, Pose: "slash",
			Outfit:    npcGearWithWeapon(GobballGear(), "weapon_goultard_cleaver"),
			HairColor: [3]float32{0.46, 0.24, 0.14}, EyeColor: [3]float32{0.96, 0.56, 0.18},
			SkinRGB: [3]float32{0.90, 0.68, 0.54}, AccentRGB: [3]float32{0.46, 0.18, 0.14},
			HairStyle: 5, FacialHair: 3,
			Greeting: "Step lively! Roaming fighters spar across the western arena between portal patrols.",
			Options: []sdk.NPCDialogOptionSpec{
				{ID: "equip_gobball", Label: "Claim Gobball Warchief Armor", Action: "equip_set"},
				{ID: "emote_slash", Label: "Show Arena Cleave Form", Action: "emote"},
				{ID: "lore_room", Label: "Ask about the Arena", Action: "lore"},
			},
		},
		{
			ID: "npc_keeper_nyx", Name: "Keeper Nyx", Role: "Twilight Seer",
			RoomX: 0, RoomY: -1, PosX: 2.4, PosZ: 2.2, FacingYaw: -2.35, Pose: "channel",
			Outfit:    npcGearWithWeapon(TwilightGear(), "weapon_xelor_staff"),
			HairColor: [3]float32{0.58, 0.34, 0.88}, EyeColor: [3]float32{0.80, 0.46, 1.00},
			SkinRGB: [3]float32{0.88, 0.78, 0.84}, AccentRGB: [3]float32{0.18, 0.12, 0.30},
			HairStyle: 2, FacialHair: 0,
			Greeting: "Shadows gather calmly beneath the southern crypt lanterns.",
			Options: []sdk.NPCDialogOptionSpec{
				{ID: "equip_twilight", Label: "Don Twilight Sanctuary Armor", Action: "equip_set"},
				{ID: "emote_heroic", Label: "Raise a Twilight Salute", Action: "emote"},
				{ID: "lore_room", Label: "Ask about the Crypt", Action: "lore"},
			},
		},
	}
}

// RegisterNPCs registers all 7 classic Sanctuary NPCs.
func RegisterNPCs() {
	for _, n := range NPCs() {
		sdk.RegisterNPC(n)
	}
}

// HandleNPCInteract executes classic Sanctuary NPC dialog options (set blessings,
// hero pose emotes, merchant shop opening, and dynamic room/scout lore) via SDK hostcalls.
func HandleNPCInteract(ev sdk.NPCInteractEvent) (sdk.NPCDialogResult, bool) {
	var npc *sdk.NPCSpec
	for _, cand := range NPCs() {
		if cand.ID == ev.NPCID {
			c := cand
			npc = &c
			break
		}
	}
	if npc == nil {
		return sdk.NPCDialogResult{}, false
	}
	if ev.OptionIndex < 0 || ev.OptionIndex >= len(npc.Options) {
		return sdk.NPCDialogResult{}, false
	}
	opt := npc.Options[ev.OptionIndex]
	switch opt.ID {
	case "equip_kokiri":
		banner := fmt.Sprintf("%s granted Kokiri Guardian Set.", npc.Name)
		sdk.SetHeroLoadout(sdk.SetHeroLoadoutCmd{TargetHeroID: ev.HeroID, Equipped: KokiriGear(), BannerText: banner})
		return sdk.NPCDialogResult{ReplyText: "Emerald threads weave neatly around your hero!", BannerText: banner}, true
	case "equip_gobball":
		banner := fmt.Sprintf("%s granted Gobball Warchief Set.", npc.Name)
		sdk.SetHeroLoadout(sdk.SetHeroLoadoutCmd{TargetHeroID: ev.HeroID, Equipped: GobballGear(), BannerText: banner})
		return sdk.NPCDialogResult{ReplyText: "Heavy warchief plates lock onto your shoulders with a battle-ready heft!", BannerText: banner}, true
	case "equip_wakfu":
		banner := fmt.Sprintf("%s granted Wakfu Chronomancer Set.", npc.Name)
		sdk.SetHeroLoadout(sdk.SetHeroLoadoutCmd{TargetHeroID: ev.HeroID, Equipped: WakfuGear(), BannerText: banner})
		return sdk.NPCDialogResult{ReplyText: "Chrono-infused gauntlets and robe shimmer across your silhouette!", BannerText: banner}, true
	case "equip_twilight":
		banner := fmt.Sprintf("%s granted Twilight Sanctuary Set.", npc.Name)
		sdk.SetHeroLoadout(sdk.SetHeroLoadoutCmd{TargetHeroID: ev.HeroID, Equipped: TwilightGear(), BannerText: banner})
		return sdk.NPCDialogResult{ReplyText: "Star-tempered steel and deep sanctuary mail settle into place!", BannerText: banner}, true

	case "emote_heroic":
		sdk.SetHeroLoadout(sdk.SetHeroLoadoutCmd{TargetHeroID: ev.HeroID, Pose: "heroic"})
		return sdk.NPCDialogResult{ReplyText: npc.Name + " snaps a crisp salute alongside you across the courtyard!", BannerText: "Hero pose switched to heroic."}, true
	case "emote_slash":
		sdk.SetHeroLoadout(sdk.SetHeroLoadoutCmd{TargetHeroID: ev.HeroID, Pose: "slash"})
		return sdk.NPCDialogResult{ReplyText: npc.Name + " nods approvingly as you demonstrate a sweeping cleave form!", BannerText: "Hero pose switched to slash."}, true
	case "emote_channel":
		sdk.SetHeroLoadout(sdk.SetHeroLoadoutCmd{TargetHeroID: ev.HeroID, Pose: "channel"})
		return sdk.NPCDialogResult{ReplyText: "Ambient energy resonates through your Armory stance alongside " + npc.Name + ".", BannerText: "Hero pose switched to channel."}, true

	case "open_merchant":
		return sdk.NPCDialogResult{
			ReplyText:    "Browse my full catalog! Sell any unused items for 50% of their Gold value.",
			OpenMerchant: true,
			ShopEnabled:  true,
		}, true
	case "lore_merchant":
		return sdk.NPCDialogResult{
			ReplyText: "New heroes start with 2,500g, and you can sell spare gear or materials back to me anytime. Portal Style items can be slotted in House Access while inside your 10x10 house!",
		}, true
	case "lore_sanctuary":
		portals := 4
		if absInt(ev.RoomX) == 2 && absInt(ev.RoomY) == 2 {
			portals = 2
		} else if absInt(ev.RoomX) == 2 || absInt(ev.RoomY) == 2 {
			portals = 3
		}
		return sdk.NPCDialogResult{
			ReplyText: fmt.Sprintf("This chunk connects to %d neighboring rooms. Step onto any glowing doorway pad to transition chunks.", portals),
		}, true
	case "lore_scouts":
		return sdk.NPCDialogResult{
			ReplyText: fmt.Sprintf("There are currently %d roaming scouts patrolling the Sanctuary chunks and crossing portals just like players.", ev.ActiveBots),
		}, true
	case "lore_room":
		roomName := ev.RoomName
		if roomName == "" {
			roomName = fmt.Sprintf("(%d,%d)", ev.RoomX, ev.RoomY)
		}
		return sdk.NPCDialogResult{
			ReplyText: fmt.Sprintf("You stand inside %s (%d,%d) with %d active hero(es) sharing this chunk.", roomName, ev.RoomX, ev.RoomY, ev.RoomOccupants),
		}, true
	default:
		return sdk.NPCDialogResult{}, false
	}
}
