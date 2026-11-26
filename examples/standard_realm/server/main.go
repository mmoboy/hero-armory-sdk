package main

import (
	sdk "github.com/mmoboy/hero-armory-sdk"
	"github.com/mmoboy/hero-armory-sdk/tactical"
	"github.com/mmoboy/hero-armory-sdk/vanilla"
)

type standardRealmPlugin struct {
	sdk.BaseServerPlugin
	combat *tactical.Engine
}

func (p *standardRealmPlugin) Init() {
	vanilla.RegisterRooms()
	vanilla.RegisterItems()
	vanilla.RegisterJobsAndGatherNodes()
	// Register Sanctuary skills & monster definitions for Codex Bestiary.
	for _, sk := range vanilla.Skills() {
		sdk.RegisterPlayerSkill(sk)
	}
	for _, m := range vanilla.Monsters() {
		sdk.RegisterMonster(m)
	}
	vanilla.RegisterNPCs()
	vanilla.RegisterCodex()

	p.combat = tactical.NewEngine(tactical.Config{
		DefaultDifficulty: tactical.TierNormal,
		ResolveLoot:       vanilla.AwardTacticalLoot,
	})
	vanilla.RegisterTacticalEncounters(p.combat)
}

func (p *standardRealmPlugin) OnTick(dtMs uint32) {
	if p.combat != nil {
		p.combat.OnTick(dtMs)
	}
}

func (p *standardRealmPlugin) OnPlayerLeave(heroID string) {
	if p.combat != nil {
		p.combat.OnPlayerLeave(heroID)
	}
}

func (p *standardRealmPlugin) OnRoomEnter(ev sdk.RoomEvent) {
	if p.combat != nil {
		p.combat.OnRoomEnter(ev)
	}
}

func (p *standardRealmPlugin) OnActorClick(ev sdk.ActorClickEvent) {
	if p.combat != nil {
		p.combat.OnActorClick(ev)
	}
}

func (p *standardRealmPlugin) OnGroundClick(ev sdk.GroundClickEvent) {
	if p.combat != nil {
		p.combat.OnGroundClick(ev)
	}
}

func (p *standardRealmPlugin) OnHotbarPress(ev sdk.HotbarPressEvent) {
	if p.combat != nil {
		p.combat.OnHotbarPress(ev)
	}
}

func (p *standardRealmPlugin) OnPacket(senderHeroID string, channel uint16, data []byte) {
	if p.combat != nil {
		p.combat.OnPacket(senderHeroID, channel, data)
	}
}

func (p *standardRealmPlugin) OnNPCInteract(ev sdk.NPCInteractEvent) (sdk.NPCDialogResult, bool) {
	return vanilla.HandleNPCInteract(ev)
}

func (p *standardRealmPlugin) OnChatCommand(ev sdk.ChatEvent) bool {
	if p.combat != nil && p.combat.HandleDifficultyCommand(ev) {
		return true
	}
	return vanilla.HandleChatCommand(ev)
}

func init() {
	sdk.RegisterServer(&standardRealmPlugin{})
}

func main() {}
