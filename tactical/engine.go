package tactical

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

// EncounterSpec configures one overworld tactical encounter actor (or pack).
type EncounterSpec struct {
	ID              string
	RoomX           int
	RoomY           int
	PosX            float32
	PosZ            float32
	TemplateID      string
	Level           int
	Tier            DifficultyTier // empty = inherit engine realm difficulty
	BalanceOverride *MobBalance
	PackSize        int     // 1..4 (default 1)
	RespawnSec      float32 // 0 = default 20s, <0 = one-shot
	ModelID         string
	ModelPath       string
	Name            string
}

// BattleResult is delivered to Config.OnBattleEnd when a tactical battle concludes.
type BattleResult struct {
	EncounterID string
	TemplateID  string
	HeroID      string
	RoomX       int
	RoomY       int
	Won         bool
	Turns       int
	GoldAwarded int64
	XPAwarded   int64
}

// Config configures a tactical.Engine instance.
type Config struct {
	DefaultDifficulty DifficultyTier
	HeroBaseHP        int
	HeroBaseAP        int
	HeroBaseMP        float32
	TurnStepMs        uint32
	Seed              uint64
	HUDChannel        uint16 // packet channel for BattleHUDSnapshot (default DefaultHUDChannel)
	ResolveLoot       func(res BattleResult) []LootDrop
	OnBattleEnd       func(res BattleResult)
}

type statusState struct {
	ID    string
	Label string
	Turns int
	Power int
	Color [3]float32
}

type encounterState struct {
	spec         EncounterSpec
	active       bool
	inBattle     bool
	respawnTimer float32
}

type mobCombatant struct {
	actorID   string
	name      string
	tmpl      MobTemplate
	stats     ScaledMobStats
	posX      float32
	posZ      float32
	yaw       float32
	hp        int
	maxHP     int
	shield    int
	ap        int
	mp        float32
	cooldowns map[string]int
	statuses  []statusState
}

type battleState struct {
	encounterID      string
	roomX            int
	roomY            int
	heroID           string
	heroName         string
	heroPosX         float32
	heroPosZ         float32
	heroHP           int
	heroMaxHP        int
	heroShield       int
	heroAP           int
	heroMaxAP        int
	heroMP           float32
	heroMaxMP        float32
	heroCooldowns    map[string]int
	heroStatuses     []statusState
	armedSpellID     string
	selectedMobID    string
	playerTurn       bool
	turnNumber       int
	aiStepIndex      int
	aiAccumMs        uint32
	mobs             []*mobCombatant
	balance          MobBalance
	tier             DifficultyTier
	baseTemplateID   string
	enemyDisplayName string
}

// Engine runs plugin-side turn-based tactical combat on top of API Level 5 primitives.
type Engine struct {
	cfg           Config
	difficulty    DifficultyTier
	monsters      map[string]MobTemplate
	spells        map[string]SpellSpec
	spellOrder    []string
	encounters    map[string]*encounterState
	battlesByHero map[string]*battleState
	actorToEnc    map[string]string
	actorToBattle map[string]string
	rngState      uint64
}

// NewEngine creates a tactical combat engine preloaded with DefaultMonsters and DefaultSpells.
func NewEngine(cfg Config) *Engine {
	if cfg.DefaultDifficulty == "" {
		cfg.DefaultDifficulty = TierNormal
	}
	if cfg.HeroBaseHP <= 0 {
		cfg.HeroBaseHP = 140
	}
	if cfg.HeroBaseAP <= 0 {
		cfg.HeroBaseAP = 6
	}
	if cfg.HeroBaseMP <= 0 {
		cfg.HeroBaseMP = 6.0
	}
	if cfg.TurnStepMs == 0 {
		cfg.TurnStepMs = 650
	}
	if cfg.HUDChannel == 0 {
		cfg.HUDChannel = DefaultHUDChannel
	}
	seed := cfg.Seed
	if seed == 0 {
		seed = 0x9e3779b97f4a7c15
	}
	e := &Engine{
		cfg:           cfg,
		difficulty:    cfg.DefaultDifficulty,
		monsters:      map[string]MobTemplate{},
		spells:        map[string]SpellSpec{},
		encounters:    map[string]*encounterState{},
		battlesByHero: map[string]*battleState{},
		actorToEnc:    map[string]string{},
		actorToBattle: map[string]string{},
		rngState:      seed,
	}
	for _, m := range DefaultMonsters() {
		e.RegisterMonster(m)
	}
	for _, sp := range DefaultSpells() {
		e.RegisterSpell(sp)
	}
	return e
}

func (e *Engine) nextRandFloat() float32 {
	x := e.rngState
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	e.rngState = x
	return float32(x&0xFFFFFF) / float32(0x1000000)
}

// RegisterMonster registers or replaces a MobTemplate.
func (e *Engine) RegisterMonster(tmpl MobTemplate) {
	e.monsters[tmpl.ID] = tmpl
}

// RegisterSpell registers or replaces a player SpellSpec.
func (e *Engine) RegisterSpell(sp SpellSpec) {
	if _, exists := e.spells[sp.ID]; !exists {
		e.spellOrder = append(e.spellOrder, sp.ID)
	}
	e.spells[sp.ID] = sp
	sort.Slice(e.spellOrder, func(i, j int) bool {
		return e.spells[e.spellOrder[i]].Slot < e.spells[e.spellOrder[j]].Slot
	})
}

// RealmDifficulty returns the current realm-wide DifficultyTier.
func (e *Engine) RealmDifficulty() DifficultyTier {
	return e.difficulty
}

// SetRealmDifficulty updates the realm-wide difficulty tier and refreshes all
// idle encounters that inherit the realm tier.
func (e *Engine) SetRealmDifficulty(tier DifficultyTier) {
	e.difficulty = tier
	for _, enc := range e.encounters {
		if !enc.active || enc.inBattle || enc.spec.Tier != "" || enc.spec.BalanceOverride != nil {
			continue
		}
		_ = e.spawnEncounterActor(enc)
	}
}

// HandleDifficultyCommand handles `/difficulty <story|normal|veteran|heroic|mythic>`
// chat commands. Returns true when consumed.
func (e *Engine) HandleDifficultyCommand(ev sdk.ChatEvent) bool {
	fields := strings.Fields(strings.TrimSpace(ev.Text))
	if len(fields) == 0 || strings.ToLower(fields[0]) != "/difficulty" {
		return false
	}
	if len(fields) < 2 {
		_ = sdk.SendChatToHero(ev.HeroID, fmt.Sprintf("Current difficulty: %s. Usage: /difficulty <story|normal|veteran|heroic|mythic>", e.difficulty))
		return true
	}
	tier, ok := ParseDifficulty(fields[1])
	if !ok {
		_ = sdk.SendChatToHero(ev.HeroID, "Unknown difficulty. Choose: story, normal, veteran, heroic, mythic.")
		return true
	}
	e.SetRealmDifficulty(tier)
	bal := PresetBalance(tier)
	_ = sdk.BroadcastChat(fmt.Sprintf("Realm difficulty set to %s (HP %.0f%%, DMG %.0f%%, AI %s, Rewards %.0f%%).",
		tier, bal.HPMult*100, bal.DamageMult*100, bal.AIProfile, bal.RewardMult*100))
	return true
}

func (e *Engine) resolveBalance(spec EncounterSpec) (DifficultyTier, MobBalance) {
	tier := spec.Tier
	if tier == "" {
		tier = e.difficulty
	}
	bal := PresetBalance(tier)
	if spec.BalanceOverride != nil {
		bal = *spec.BalanceOverride
	}
	return tier, bal
}

// AddEncounter registers and spawns an overworld tactical encounter actor.
func (e *Engine) AddEncounter(spec EncounterSpec) error {
	if !sdk.IsValidContentID(spec.ID) {
		return sdk.ErrInvalidArg
	}
	if _, ok := e.monsters[spec.TemplateID]; !ok {
		return sdk.ErrNotFound
	}
	if spec.Level <= 0 {
		spec.Level = 1
	}
	if spec.PackSize <= 0 {
		spec.PackSize = 1
	}
	if spec.PackSize > 4 {
		spec.PackSize = 4
	}
	if spec.RespawnSec == 0 {
		spec.RespawnSec = 20.0
	}
	enc := &encounterState{spec: spec, active: true}
	if err := e.spawnEncounterActor(enc); err != nil {
		return err
	}
	e.encounters[spec.ID] = enc
	e.actorToEnc[spec.ID] = spec.ID
	return nil
}

// RemoveEncounter removes an overworld encounter and its actor.
func (e *Engine) RemoveEncounter(id string) error {
	enc, ok := e.encounters[id]
	if !ok {
		return sdk.ErrNotFound
	}
	delete(e.encounters, id)
	delete(e.actorToEnc, id)
	if enc.active && !enc.inBattle {
		_ = sdk.RemoveActor(enc.spec.ID)
	}
	return nil
}

// InBattle reports whether heroID is currently in a tactical battle.
func (e *Engine) InBattle(heroID string) bool {
	_, ok := e.battlesByHero[heroID]
	return ok
}

// ActiveBattles returns the number of active tactical battles.
func (e *Engine) ActiveBattles() int {
	return len(e.battlesByHero)
}

func (e *Engine) spawnEncounterActor(enc *encounterState) error {
	tmpl := e.monsters[enc.spec.TemplateID]
	tier, bal := e.resolveBalance(enc.spec)
	stats := ScaleMob(tmpl, enc.spec.Level, bal)
	name := enc.spec.Name
	if name == "" {
		name = fmt.Sprintf("%s [Lv.%d]", tmpl.Name, stats.Level)
	}
	modelID := enc.spec.ModelID
	if modelID == "" {
		modelID = tmpl.ModelID
	}
	modelPath := enc.spec.ModelPath
	if modelPath == "" {
		modelPath = tmpl.ModelPath
	}
	statuses := []sdk.ActorStatusSpec{
		{ID: "tier", Label: strings.ToUpper(string(tier)), Color: tierColor(tier)},
	}
	return sdk.SpawnActor(sdk.ActorSpec{
		ID:         enc.spec.ID,
		RoomX:      enc.spec.RoomX,
		RoomY:      enc.spec.RoomY,
		PosX:       enc.spec.PosX,
		PosZ:       enc.spec.PosZ,
		Name:       name,
		Kind:       tmpl.Role,
		Team:       "enemy",
		ModelID:    modelID,
		ModelPath:  modelPath,
		BodyColor:  tmpl.BodyColor,
		TrimColor:  tmpl.TrimColor,
		Scale:      tmpl.Scale,
		HP:         stats.MaxHP,
		MaxHP:      stats.MaxHP,
		Shield:     stats.Shield,
		Pose:       "idle",
		Statuses:   statuses,
		Targetable: true,
	})
}

func tierColor(tier DifficultyTier) [3]float32 {
	switch tier {
	case TierStory:
		return [3]float32{0.55, 0.85, 0.65}
	case TierVeteran:
		return [3]float32{0.36, 0.82, 1.0}
	case TierHeroic:
		return [3]float32{0.96, 0.64, 0.24}
	case TierMythic:
		return [3]float32{0.92, 0.30, 0.96}
	default:
		return [3]float32{0.82, 0.86, 0.92}
	}
}

// OnTick advances AI turns and overworld encounter respawn timers.
func (e *Engine) OnTick(dtMs uint32) {
	if dtMs == 0 {
		return
	}
	dtSec := float32(dtMs) / 1000.0
	for _, enc := range e.encounters {
		if enc.active || enc.inBattle || enc.spec.RespawnSec < 0 {
			continue
		}
		enc.respawnTimer -= dtSec
		if enc.respawnTimer <= 0 {
			enc.active = true
			enc.respawnTimer = 0
			_ = e.spawnEncounterActor(enc)
		}
	}

	heroIDs := make([]string, 0, len(e.battlesByHero))
	for id := range e.battlesByHero {
		heroIDs = append(heroIDs, id)
	}
	sort.Strings(heroIDs)
	for _, heroID := range heroIDs {
		b, ok := e.battlesByHero[heroID]
		if !ok || b.playerTurn {
			continue
		}
		b.aiAccumMs += dtMs
		if b.aiAccumMs < e.cfg.TurnStepMs {
			continue
		}
		b.aiAccumMs = 0
		e.stepEnemyTurn(b)
	}
}

// OnPlayerLeave aborts any active battle for heroID cleanly.
func (e *Engine) OnPlayerLeave(heroID string) {
	if b, ok := e.battlesByHero[heroID]; ok {
		e.finishBattle(b, false, true)
	}
}

// OnRoomEnter aborts an active battle if the hero left the battle's room.
func (e *Engine) OnRoomEnter(ev sdk.RoomEvent) {
	if b, ok := e.battlesByHero[ev.HeroID]; ok {
		if ev.RoomX != b.roomX || ev.RoomY != b.roomY {
			e.finishBattle(b, false, true)
		}
	}
}

// OnPacket handles client->server tactical HUD actions (such as clicking the
// End Turn button inside tactical_hud). Returns true when channel matched.
func (e *Engine) OnPacket(heroID string, channel uint16, data []byte) bool {
	if channel != e.cfg.HUDChannel {
		return false
	}
	var act BattleHUDAction
	if json.Unmarshal(data, &act) != nil {
		return true
	}
	if act.Action == EndTurnActionID {
		if b, ok := e.battlesByHero[heroID]; ok && b.playerTurn {
			e.endPlayerTurn(b)
		}
	}
	return true
}

// OnActorClick handles clicking overworld encounter actors or active combat targets.
func (e *Engine) OnActorClick(ev sdk.ActorClickEvent) {
	// Right-click inspects either an active combatant or an overworld encounter.
	if ev.Button == "right" {
		if bHero, ok := e.actorToBattle[ev.ActorID]; ok {
			if b, exists := e.battlesByHero[bHero]; exists {
				if m := findMob(b, ev.ActorID); m != nil {
					_ = sdk.SendChatToHero(ev.HeroID, fmt.Sprintf("%s (%s, %s): %d/%d HP, %d Shield, %d DMG, AI %s",
						m.name, m.tmpl.Role, b.tier, m.hp, m.maxHP, m.shield, m.stats.Damage, m.stats.AIProfile))
					return
				}
			}
		}
		if encID, ok := e.actorToEnc[ev.ActorID]; ok {
			if enc, exists := e.encounters[encID]; exists {
				tmpl := e.monsters[enc.spec.TemplateID]
				tier, bal := e.resolveBalance(enc.spec)
				st := ScaleMob(tmpl, enc.spec.Level, bal)
				_ = sdk.SendChatToHero(ev.HeroID, fmt.Sprintf("%s [Lv.%d %s x%d]: %d HP, %d Shield, %d DMG, AI %s - %s",
					tmpl.Name, st.Level, tier, enc.spec.PackSize, st.MaxHP, st.Shield, st.Damage, st.AIProfile, tmpl.Description))
			}
		}
		return
	}

	// Left-click during active battle.
	if b, inBattle := e.battlesByHero[ev.HeroID]; inBattle {
		if !b.playerTurn {
			_ = sdk.SendChatToHero(ev.HeroID, "Wait for enemy turn to finish.")
			return
		}
		m := findMob(b, ev.ActorID)
		if m == nil || m.hp <= 0 {
			return
		}
		b.selectedMobID = m.actorID
		spellID := b.armedSpellID
		if spellID == "" {
			spellID = "heavy_strike"
		}
		e.castPlayerSpellOnMob(b, spellID, m)
		return
	}

	// Left-click on an idle overworld encounter starts tactical combat.
	if encID, ok := e.actorToEnc[ev.ActorID]; ok {
		enc := e.encounters[encID]
		if enc != nil && enc.active && !enc.inBattle {
			e.startBattle(ev.HeroID, enc)
		}
	}
}

// OnGroundClick handles tactical movement and ground-targeted spells (like Gale Rush).
func (e *Engine) OnGroundClick(ev sdk.GroundClickEvent) {
	b, ok := e.battlesByHero[ev.HeroID]
	if !ok || !b.playerTurn || ev.RoomX != b.roomX || ev.RoomY != b.roomY {
		return
	}
	if ev.PosX < -26 || ev.PosX > 26 || ev.PosZ < -26 || ev.PosZ > 26 {
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
			Text: "Out of bounds", Color: [4]float32{1.0, 0.55, 0.3, 1}, Scale: 0.9,
		})
		return
	}
	if b.armedSpellID != "" {
		sp, exists := e.spells[b.armedSpellID]
		if exists && sp.EffectKind == "dash_or_knockback" {
			e.castPlayerDash(b, sp, ev.PosX, ev.PosZ)
			return
		}
	}

	dist := dist2D(b.heroPosX, b.heroPosZ, ev.PosX, ev.PosZ)
	if dist < 0.25 {
		return
	}
	if dist > b.heroMP+0.05 {
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
			Text: fmt.Sprintf("Need %.1f MP", dist), Color: [4]float32{1.0, 0.72, 0.32, 1}, Scale: 0.9,
		})
		return
	}
	if err := sdk.TeleportHero(b.heroID, b.roomX, b.roomY, [3]float32{ev.PosX, 0, ev.PosZ}, true); err != nil {
		return
	}
	b.heroMP -= dist
	if b.heroMP < 0 {
		b.heroMP = 0
	}
	b.heroPosX = ev.PosX
	b.heroPosZ = ev.PosZ
	e.refreshBattleHUDAndMarkers(b)
}

// OnHotbarPress handles slots 1..8 (player spells) and slot 10 (End Turn).
func (e *Engine) OnHotbarPress(ev sdk.HotbarPressEvent) {
	b, ok := e.battlesByHero[ev.HeroID]
	if !ok || !b.playerTurn {
		return
	}
	if ev.Slot == 10 || ev.ActionID == EndTurnActionID {
		e.endPlayerTurn(b)
		return
	}
	sp, ok := e.spells[ev.ActionID]
	if !ok {
		for _, cand := range e.spells {
			if cand.Slot == ev.Slot {
				sp = cand
				ok = true
				break
			}
		}
	}
	if !ok {
		return
	}
	if b.heroCooldowns[sp.ID] > 0 {
		_ = sdk.SendChatToHero(b.heroID, fmt.Sprintf("%s is on cooldown (%d turns).", sp.Name, b.heroCooldowns[sp.ID]))
		return
	}
	if b.heroAP < sp.APCost {
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
			Text: "Not enough AP", Color: [4]float32{1.0, 0.45, 0.35, 1}, Scale: 0.9,
		})
		return
	}
	if sp.TargetKind == "self" {
		e.castPlayerSelfSpell(b, sp)
		return
	}
	// If pressing an already-armed enemy spell and a living target is selected, fire immediately.
	if b.armedSpellID == sp.ID && sp.TargetKind == "enemy" {
		if target := e.pickSelectedOrSoleMob(b); target != nil {
			e.castPlayerSpellOnMob(b, sp.ID, target)
			return
		}
	}
	b.armedSpellID = sp.ID
	e.refreshBattleHUDAndMarkers(b)
}

func (e *Engine) startBattle(heroID string, enc *encounterState) {
	hero, found, err := sdk.GetHero(heroID)
	if err != nil || !found {
		return
	}
	enc.inBattle = true
	tmpl := e.monsters[enc.spec.TemplateID]
	tier, bal := e.resolveBalance(enc.spec)
	stats := ScaleMob(tmpl, enc.spec.Level, bal)

	heroMaxHP := e.cfg.HeroBaseHP + (hero.Level-1)*14
	if heroMaxHP < e.cfg.HeroBaseHP {
		heroMaxHP = e.cfg.HeroBaseHP
	}
	b := &battleState{
		encounterID:      enc.spec.ID,
		roomX:            enc.spec.RoomX,
		roomY:            enc.spec.RoomY,
		heroID:           heroID,
		heroName:         hero.Name,
		heroPosX:         hero.PosX,
		heroPosZ:         hero.PosZ,
		heroHP:           heroMaxHP,
		heroMaxHP:        heroMaxHP,
		heroAP:           e.cfg.HeroBaseAP,
		heroMaxAP:        e.cfg.HeroBaseAP,
		heroMP:           e.cfg.HeroBaseMP,
		heroMaxMP:        e.cfg.HeroBaseMP,
		heroCooldowns:    map[string]int{},
		armedSpellID:     "heavy_strike",
		playerTurn:       true,
		turnNumber:       1,
		balance:          bal,
		tier:             tier,
		baseTemplateID:   tmpl.ID,
		enemyDisplayName: tmpl.Name,
	}

	offsets := [][2]float32{{0, 0}, {-2.6, -1.6}, {2.6, -1.6}, {0, -3.2}}
	for i := 0; i < enc.spec.PackSize; i++ {
		actorID := enc.spec.ID
		if i > 0 {
			actorID = fmt.Sprintf("%s_p%d", enc.spec.ID, i)
		}
		mx := enc.spec.PosX + offsets[i][0]
		mz := enc.spec.PosZ + offsets[i][1]
		name := tmpl.Name
		if enc.spec.PackSize > 1 {
			name = fmt.Sprintf("%s #%d", tmpl.Name, i+1)
		}
		m := &mobCombatant{
			actorID:   actorID,
			name:      name,
			tmpl:      tmpl,
			stats:     stats,
			posX:      mx,
			posZ:      mz,
			yaw:       yawToward(mx, mz, b.heroPosX, b.heroPosZ),
			hp:        stats.MaxHP,
			maxHP:     stats.MaxHP,
			shield:    stats.Shield,
			ap:        stats.AP,
			mp:        stats.MP,
			cooldowns: map[string]int{},
		}
		b.mobs = append(b.mobs, m)
		e.actorToBattle[actorID] = heroID
		modelID := enc.spec.ModelID
		if modelID == "" {
			modelID = tmpl.ModelID
		}
		modelPath := enc.spec.ModelPath
		if modelPath == "" {
			modelPath = tmpl.ModelPath
		}
		_ = sdk.SpawnActor(sdk.ActorSpec{
			ID:         actorID,
			RoomX:      b.roomX,
			RoomY:      b.roomY,
			PosX:       mx,
			PosZ:       mz,
			Yaw:        m.yaw,
			Name:       name,
			Kind:       tmpl.Role,
			Team:       "enemy",
			ModelID:    modelID,
			ModelPath:  modelPath,
			BodyColor:  tmpl.BodyColor,
			TrimColor:  tmpl.TrimColor,
			Scale:      tmpl.Scale,
			HP:         m.hp,
			MaxHP:      m.maxHP,
			Shield:     m.shield,
			Pose:       "idle",
			Targetable: true,
		})
	}
	if len(b.mobs) > 0 {
		b.selectedMobID = b.mobs[0].actorID
	}

	e.battlesByHero[heroID] = b
	_ = sdk.SetHeroControlLock(heroID, true)
	_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
		RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.5, PosZ: b.heroPosZ,
		Text: "BATTLE START!", Color: [4]float32{0.98, 0.84, 0.34, 1}, Scale: 1.2, Critical: true,
	})
	e.refreshBattleHUDAndMarkers(b)
}

func (e *Engine) pickSelectedOrSoleMob(b *battleState) *mobCombatant {
	if b.selectedMobID != "" {
		if m := findMob(b, b.selectedMobID); m != nil && m.hp > 0 {
			return m
		}
	}
	var alive *mobCombatant
	count := 0
	for _, m := range b.mobs {
		if m.hp > 0 {
			alive = m
			count++
		}
	}
	if count == 1 {
		return alive
	}
	return nil
}

func (e *Engine) castPlayerSelfSpell(b *battleState, sp SpellSpec) {
	b.heroAP -= sp.APCost
	if sp.Cooldown > 0 {
		b.heroCooldowns[sp.ID] = sp.Cooldown
	}
	_ = sdk.SetHeroPose(b.heroID, sp.HeroPose, 0)
	_ = sdk.PlayVFX(sdk.VFXSpec{
		RoomX: b.roomX, RoomY: b.roomY, Kind: sp.VFXKind,
		FromX: b.heroPosX, FromY: 1.0, FromZ: b.heroPosZ,
		ToX: b.heroPosX, ToY: 1.0, ToZ: b.heroPosZ,
		Radius: 2.0, Color: sp.VFXColor, DurationSec: 0.65,
	})
	switch sp.EffectKind {
	case "shield":
		b.heroShield += sp.BaseDamage
		if sp.StatusID != "" {
			b.heroStatuses = upsertStatus(b.heroStatuses, statusState{
				ID: sp.StatusID, Label: sp.StatusLabel, Turns: sp.StatusTurns, Power: sp.StatusPower, Color: sp.StatusColor,
			})
		}
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
			Text: fmt.Sprintf("+%d SHIELD", sp.BaseDamage), Color: sp.VFXColor, Scale: 1.1,
		})
	case "heal":
		healed := sp.BaseDamage
		if b.heroHP+healed > b.heroMaxHP {
			healed = b.heroMaxHP - b.heroHP
		}
		b.heroHP += healed
		if sp.Cleanse {
			b.heroStatuses = cleanseDebuffs(b.heroStatuses)
		}
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
			Text: fmt.Sprintf("+%d HP", healed), Color: sp.VFXColor, Scale: 1.1,
		})
	}
	e.refreshBattleHUDAndMarkers(b)
}

func (e *Engine) castPlayerDash(b *battleState, sp SpellSpec, tx, tz float32) {
	if b.heroAP < sp.APCost {
		return
	}
	if dist2D(b.heroPosX, b.heroPosZ, tx, tz) > sp.Range+0.1 {
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
			Text: "Out of range", Color: [4]float32{1.0, 0.55, 0.3, 1}, Scale: 0.9,
		})
		return
	}
	b.heroAP -= sp.APCost
	if sp.Cooldown > 0 {
		b.heroCooldowns[sp.ID] = sp.Cooldown
	}
	fromX, fromZ := b.heroPosX, b.heroPosZ
	b.heroPosX, b.heroPosZ = tx, tz
	_ = sdk.SetHeroPose(b.heroID, sp.HeroPose, yawToward(fromX, fromZ, tx, tz))
	_ = sdk.TeleportHero(b.heroID, b.roomX, b.roomY, [3]float32{tx, 0, tz}, true)
	_ = sdk.PlayVFX(sdk.VFXSpec{
		RoomX: b.roomX, RoomY: b.roomY, Kind: sp.VFXKind,
		FromX: fromX, FromY: 1.0, FromZ: fromZ, ToX: tx, ToY: 1.0, ToZ: tz,
		Radius: 1.8, Color: sp.VFXColor, DurationSec: 0.45,
	})
	_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
		RoomX: b.roomX, RoomY: b.roomY, PosX: tx, PosY: 2.2, PosZ: tz,
		Text: "DASH!", Color: sp.VFXColor, Scale: 1.0,
	})
	b.armedSpellID = "heavy_strike"
	e.refreshBattleHUDAndMarkers(b)
}

func (e *Engine) castPlayerSpellOnMob(b *battleState, spellID string, target *mobCombatant) {
	sp, ok := e.spells[spellID]
	if !ok {
		return
	}
	if b.heroCooldowns[sp.ID] > 0 {
		_ = sdk.SendChatToHero(b.heroID, fmt.Sprintf("%s is on cooldown (%d turns).", sp.Name, b.heroCooldowns[sp.ID]))
		return
	}
	if b.heroAP < sp.APCost {
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
			Text: "Not enough AP", Color: [4]float32{1.0, 0.45, 0.35, 1}, Scale: 0.9,
		})
		return
	}
	if sp.TargetKind == "self" {
		e.castPlayerSelfSpell(b, sp)
		return
	}
	dist := dist2D(b.heroPosX, b.heroPosZ, target.posX, target.posZ)
	if dist > sp.Range+0.2 || (sp.MinRange > 0 && dist < sp.MinRange) {
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
			Text: fmt.Sprintf("Range %.1fm", sp.Range), Color: [4]float32{1.0, 0.65, 0.3, 1}, Scale: 0.9,
		})
		return
	}

	b.heroAP -= sp.APCost
	if sp.Cooldown > 0 {
		b.heroCooldowns[sp.ID] = sp.Cooldown
	}
	yaw := yawToward(b.heroPosX, b.heroPosZ, target.posX, target.posZ)
	_ = sdk.SetHeroPose(b.heroID, sp.HeroPose, yaw)
	_ = sdk.PlayVFX(sdk.VFXSpec{
		RoomX: b.roomX, RoomY: b.roomY, Kind: sp.VFXKind,
		FromX: b.heroPosX, FromY: 1.2, FromZ: b.heroPosZ,
		ToX: target.posX, ToY: 1.2, ToZ: target.posZ,
		Radius: maxF32(sp.AoERadius, 1.4), Color: sp.VFXColor, DurationSec: 0.55,
	})

	targets := []*mobCombatant{target}
	if sp.AoERadius > 0 {
		for _, m := range b.mobs {
			if m == target || m.hp <= 0 {
				continue
			}
			if dist2D(target.posX, target.posZ, m.posX, m.posZ) <= sp.AoERadius {
				targets = append(targets, m)
			}
		}
	}

	for _, m := range targets {
		dmg := float32(sp.BaseDamage)
		crit := sp.CritChance > 0 && e.nextRandFloat() < sp.CritChance
		if crit {
			mult := sp.CritMult
			if mult < 1.25 {
				mult = 1.5
			}
			dmg *= mult
		}
		// Apply mob armor & aegis reduction.
		reduction := m.stats.ArmorPct
		if hasStatus(m.statuses, "aegis") {
			reduction += 0.20
		}
		if reduction > 0.75 {
			reduction = 0.75
		}
		finalDmg := int(math.Round(float64(dmg * (1.0 - reduction))))
		if finalDmg < 1 {
			finalDmg = 1
		}
		applyDamageToMob(m, finalDmg, false)

		if sp.StatusID != "" && m.hp > 0 {
			m.statuses = upsertStatus(m.statuses, statusState{
				ID: sp.StatusID, Label: sp.StatusLabel, Turns: sp.StatusTurns, Power: sp.StatusPower, Color: sp.StatusColor,
			})
		}
		if sp.CritStun && crit && m.hp > 0 {
			m.statuses = upsertStatus(m.statuses, statusState{
				ID: "stun", Label: "Stun", Turns: 1, Color: [3]float32{0.96, 0.90, 0.28},
			})
		}
		if sp.Knockback > 0 && m.hp > 0 {
			dx, dz := m.posX-b.heroPosX, m.posZ-b.heroPosZ
			l := float32(math.Hypot(float64(dx), float64(dz)))
			if l > 0.01 {
				m.posX = clampF32(m.posX+(dx/l)*sp.Knockback, -24, 24)
				m.posZ = clampF32(m.posZ+(dz/l)*sp.Knockback, -24, 24)
			}
		}

		label := fmt.Sprintf("-%d", finalDmg)
		if crit {
			label = fmt.Sprintf("-%d!", finalDmg)
		}
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: m.posX, PosY: 2.3, PosZ: m.posZ,
			Text: label, Color: sp.VFXColor, Scale: 1.15, Critical: crit,
		})
		e.syncMobActor(m, sp.Knockback > 0)
	}

	if allMobsDefeated(b) {
		e.finishBattle(b, true, false)
		return
	}
	e.ensureSelectedMobAlive(b)
	e.refreshBattleHUDAndMarkers(b)
}

func (e *Engine) ensureSelectedMobAlive(b *battleState) {
	if b.selectedMobID != "" {
		if m := findMob(b, b.selectedMobID); m != nil && m.hp > 0 {
			return
		}
	}
	b.selectedMobID = ""
	for _, m := range b.mobs {
		if m.hp > 0 {
			b.selectedMobID = m.actorID
			return
		}
	}
}

func (e *Engine) endPlayerTurn(b *battleState) {
	b.playerTurn = false
	b.aiStepIndex = 0
	b.aiAccumMs = 0
	_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
		RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.3, PosZ: b.heroPosZ,
		Text: "END TURN", Color: [4]float32{0.96, 0.84, 0.36, 1}, Scale: 0.95,
	})
	_ = sdk.ClearGroundMarkers(b.roomX, b.roomY, b.heroID)
	e.refreshBattleHUDAndMarkers(b)
}

func (e *Engine) stepEnemyTurn(b *battleState) {
	for b.aiStepIndex < len(b.mobs) {
		m := b.mobs[b.aiStepIndex]
		b.aiStepIndex++
		if m.hp <= 0 {
			continue
		}
		e.executeMobTurn(b, m)
		if b.heroHP <= 0 {
			e.finishBattle(b, false, false)
			return
		}
		if allMobsDefeated(b) {
			e.finishBattle(b, true, false)
			return
		}
		e.refreshBattleHUDAndMarkers(b)
		return
	}
	// All mobs acted -> begin next player turn.
	e.beginPlayerTurn(b)
}

func (e *Engine) executeMobTurn(b *battleState, m *mobCombatant) {
	// Decay mob cooldowns and tick statuses.
	for k, v := range m.cooldowns {
		if v <= 1 {
			delete(m.cooldowns, k)
		} else {
			m.cooldowns[k] = v - 1
		}
	}
	stunned := e.tickMobStatuses(b, m)
	if m.hp <= 0 {
		e.syncMobActor(m, false)
		return
	}
	if stunned {
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: m.posX, PosY: 2.4, PosZ: m.posZ,
			Text: "STUNNED", Color: [4]float32{0.96, 0.90, 0.28, 1}, Scale: 1.1,
		})
		e.syncMobActor(m, false)
		return
	}

	m.ap = m.stats.AP
	m.mp = m.stats.MP
	if hasStatus(m.statuses, "chill") {
		m.mp = maxF32(1.5, m.mp-2.0)
	}

	// Move toward hero if out of preferred range.
	desiredRange := m.tmpl.Range
	if desiredRange < 2.5 {
		desiredRange = 2.8
	}
	dist := dist2D(m.posX, m.posZ, b.heroPosX, b.heroPosZ)
	moved := false
	if dist > desiredRange && m.mp > 0.2 {
		step := minF32(m.mp, dist-desiredRange+0.4)
		if step > 0.2 {
			dx, dz := b.heroPosX-m.posX, b.heroPosZ-m.posZ
			m.posX = clampF32(m.posX+(dx/dist)*step, -24, 24)
			m.posZ = clampF32(m.posZ+(dz/dist)*step, -24, 24)
			m.yaw = yawToward(m.posX, m.posZ, b.heroPosX, b.heroPosZ)
			moved = true
		}
	}

	maxActions := 1
	if m.stats.AIProfile == "ruthless" {
		maxActions = 2
	}
	for act := 0; act < maxActions && m.ap >= 2 && b.heroHP > 0; act++ {
		if !e.performOneMobAction(b, m) {
			break
		}
	}
	e.syncMobActor(m, moved)
}

func (e *Engine) performOneMobAction(b *battleState, m *mobCombatant) bool {
	dist := dist2D(m.posX, m.posZ, b.heroPosX, b.heroPosZ)

	// Tactical/ruthless AI checks defensive/healing spells when hurt.
	if m.stats.AIProfile == "tactical" || m.stats.AIProfile == "ruthless" {
		for _, sp := range m.tmpl.Spells {
			if m.ap < sp.APCost || m.cooldowns[sp.ID] > 0 {
				continue
			}
			if sp.EffectKind == "heal" && m.hp < (m.maxHP*65)/100 {
				m.ap -= sp.APCost
				m.cooldowns[sp.ID] = sp.Cooldown
				m.hp = minInt(m.maxHP, m.hp+sp.Heal)
				_ = sdk.PlayVFX(sdk.VFXSpec{
					RoomX: b.roomX, RoomY: b.roomY, Kind: sp.VFXKind,
					FromX: m.posX, FromY: 1.0, FromZ: m.posZ, ToX: m.posX, ToY: 1.0, ToZ: m.posZ,
					Radius: 2.0, Color: sp.VFXColor, DurationSec: 0.6,
				})
				_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
					RoomX: b.roomX, RoomY: b.roomY, PosX: m.posX, PosY: 2.3, PosZ: m.posZ,
					Text: fmt.Sprintf("+%d HP", sp.Heal), Color: sp.VFXColor, Scale: 1.0,
				})
				return true
			}
			if sp.EffectKind == "shield" && (m.shield == 0 && (m.stats.AIProfile == "ruthless" || m.hp < (m.maxHP*75)/100)) {
				m.ap -= sp.APCost
				m.cooldowns[sp.ID] = sp.Cooldown
				m.shield += sp.Shield
				if sp.StatusID != "" {
					m.statuses = upsertStatus(m.statuses, statusState{
						ID: sp.StatusID, Label: sp.StatusLabel, Turns: sp.StatusTurns, Power: sp.StatusPower, Color: sp.StatusColor,
					})
				}
				_ = sdk.PlayVFX(sdk.VFXSpec{
					RoomX: b.roomX, RoomY: b.roomY, Kind: sp.VFXKind,
					FromX: m.posX, FromY: 1.0, FromZ: m.posZ, ToX: m.posX, ToY: 1.0, ToZ: m.posZ,
					Radius: 2.0, Color: sp.VFXColor, DurationSec: 0.6,
				})
				return true
			}
		}
	}

	// Pick best offensive spell in range.
	var chosen *MobSpellSpec
	for i := range m.tmpl.Spells {
		sp := &m.tmpl.Spells[i]
		if sp.EffectKind != "damage" || m.ap < sp.APCost || m.cooldowns[sp.ID] > 0 {
			continue
		}
		if dist <= sp.Range+0.3 {
			if chosen == nil || sp.Damage > chosen.Damage {
				chosen = sp
			}
		}
	}
	if chosen != nil {
		m.ap -= chosen.APCost
		m.cooldowns[chosen.ID] = chosen.Cooldown
		dmg := float32(chosen.Damage) * (float32(m.stats.Damage) / maxF32(float32(m.tmpl.BaseDamage), 1))
		e.applyMobAttackToHero(b, m, dmg, chosen.VFXKind, chosen.VFXColor, chosen)
		return true
	}

	// Fallback basic attack if in range.
	if dist <= m.tmpl.Range+0.4 && m.ap >= 2 {
		m.ap -= 2
		vfx := "slash"
		if m.tmpl.Range > 5.0 {
			vfx = "bolt"
		}
		e.applyMobAttackToHero(b, m, float32(m.stats.Damage), vfx, [4]float32{1.0, 0.36, 0.28, 1}, nil)
		return true
	}
	return false
}

func (e *Engine) applyMobAttackToHero(b *battleState, m *mobCombatant, rawDmg float32, vfxKind string, vfxCol [4]float32, sp *MobSpellSpec) {
	crit := m.stats.CritChance > 0 && e.nextRandFloat() < m.stats.CritChance
	if crit {
		rawDmg *= m.stats.CritMult
	}
	if hasStatus(b.heroStatuses, "aegis") {
		rawDmg *= 0.80
	}
	dmg := int(math.Round(float64(rawDmg)))
	if dmg < 1 {
		dmg = 1
	}
	applyDamageToHero(b, dmg, false)
	if sp != nil && sp.StatusID != "" {
		b.heroStatuses = upsertStatus(b.heroStatuses, statusState{
			ID: sp.StatusID, Label: sp.StatusLabel, Turns: sp.StatusTurns, Power: sp.StatusPower, Color: sp.StatusColor,
		})
	}
	_ = sdk.UpdateActor(sdk.ActorPatch{ID: m.actorID, Pose: sdk.PtrString("slash"), Yaw: sdk.PtrFloat32(yawToward(m.posX, m.posZ, b.heroPosX, b.heroPosZ))})
	_ = sdk.SetHeroPose(b.heroID, "hit", yawToward(b.heroPosX, b.heroPosZ, m.posX, m.posZ))
	_ = sdk.PlayVFX(sdk.VFXSpec{
		RoomX: b.roomX, RoomY: b.roomY, Kind: vfxKind,
		FromX: m.posX, FromY: 1.2, FromZ: m.posZ, ToX: b.heroPosX, ToY: 1.2, ToZ: b.heroPosZ,
		Radius: 1.6, Color: vfxCol, DurationSec: 0.5,
	})
	lbl := fmt.Sprintf("-%d", dmg)
	if crit {
		lbl = fmt.Sprintf("-%d!", dmg)
	}
	_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
		RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.3, PosZ: b.heroPosZ,
		Text: lbl, Color: [4]float32{1.0, 0.28, 0.24, 1}, Scale: 1.15, Critical: crit,
	})
}

func (e *Engine) beginPlayerTurn(b *battleState) {
	b.turnNumber++
	b.playerTurn = true
	b.heroAP = b.heroMaxAP
	b.heroMP = b.heroMaxMP
	for k, v := range b.heroCooldowns {
		if v <= 1 {
			delete(b.heroCooldowns, k)
		} else {
			b.heroCooldowns[k] = v - 1
		}
	}
	stunned := e.tickHeroStatuses(b)
	if b.heroHP <= 0 {
		e.finishBattle(b, false, false)
		return
	}
	if hasStatus(b.heroStatuses, "chill") {
		b.heroMP = maxF32(1.5, b.heroMP-2.0)
	}
	if stunned {
		_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
			RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.4, PosZ: b.heroPosZ,
			Text: "STUNNED", Color: [4]float32{0.96, 0.90, 0.28, 1}, Scale: 1.15,
		})
		e.endPlayerTurn(b)
		return
	}
	e.ensureSelectedMobAlive(b)
	if sp, ok := e.spells[b.armedSpellID]; !ok || b.heroCooldowns[sp.ID] > 0 || b.heroAP < sp.APCost {
		b.armedSpellID = "heavy_strike"
	}
	e.refreshBattleHUDAndMarkers(b)
}

func (e *Engine) tickMobStatuses(b *battleState, m *mobCombatant) (stunned bool) {
	var next []statusState
	for _, st := range m.statuses {
		switch st.ID {
		case "burn":
			applyDamageToMob(m, st.Power, false)
			_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
				RoomX: b.roomX, RoomY: b.roomY, PosX: m.posX, PosY: 2.2, PosZ: m.posZ,
				Text: fmt.Sprintf("-%d BURN", st.Power), Color: [4]float32{1.0, 0.48, 0.16, 1}, Scale: 0.95,
			})
		case "poison":
			applyDamageToMob(m, st.Power, true)
			_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
				RoomX: b.roomX, RoomY: b.roomY, PosX: m.posX, PosY: 2.2, PosZ: m.posZ,
				Text: fmt.Sprintf("-%d PSN", st.Power), Color: [4]float32{0.40, 0.96, 0.36, 1}, Scale: 0.95,
			})
		case "stun":
			stunned = true
		}
		st.Turns--
		if st.Turns > 0 {
			next = append(next, st)
		}
	}
	m.statuses = next
	return stunned
}

func (e *Engine) tickHeroStatuses(b *battleState) (stunned bool) {
	var next []statusState
	for _, st := range b.heroStatuses {
		switch st.ID {
		case "burn":
			applyDamageToHero(b, st.Power, false)
			_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
				RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
				Text: fmt.Sprintf("-%d BURN", st.Power), Color: [4]float32{1.0, 0.48, 0.16, 1}, Scale: 0.95,
			})
		case "poison":
			applyDamageToHero(b, st.Power, true)
			_ = sdk.ShowFloatingText(sdk.FloatingTextSpec{
				RoomX: b.roomX, RoomY: b.roomY, PosX: b.heroPosX, PosY: 2.2, PosZ: b.heroPosZ,
				Text: fmt.Sprintf("-%d PSN", st.Power), Color: [4]float32{0.40, 0.96, 0.36, 1}, Scale: 0.95,
			})
		case "stun":
			stunned = true
		}
		st.Turns--
		if st.Turns > 0 {
			next = append(next, st)
		}
	}
	b.heroStatuses = next
	return stunned
}

func (e *Engine) syncMobActor(m *mobCombatant, moved bool) {
	pose := "hit"
	targetable := true
	if m.hp <= 0 {
		pose = "defeat"
		targetable = false
	}
	patch := sdk.ActorPatch{
		ID:              m.actorID,
		HP:              sdk.PtrInt(m.hp),
		Shield:          sdk.PtrInt(m.shield),
		Pose:            sdk.PtrString(pose),
		Statuses:        toActorStatuses(m.statuses),
		ReplaceStatuses: true,
		Targetable:      sdk.PtrBool(targetable),
	}
	if moved {
		patch.PosX = sdk.PtrFloat32(m.posX)
		patch.PosZ = sdk.PtrFloat32(m.posZ)
		patch.Yaw = sdk.PtrFloat32(m.yaw)
		patch.MoveTo = true
	}
	_ = sdk.UpdateActor(patch)
}

func (e *Engine) refreshBattleHUDAndMarkers(b *battleState) {
	slots := make([]sdk.HotbarSlotSpec, 0, len(e.spellOrder)+1)
	for _, id := range e.spellOrder {
		sp := e.spells[id]
		cd := b.heroCooldowns[sp.ID]
		disabled := !b.playerTurn || cd > 0 || b.heroAP < sp.APCost
		group := "spell"
		if sp.TargetKind == "self" || sp.EffectKind == "dash_or_knockback" {
			group = "utility"
		}
		accent := [3]float32{sp.VFXColor[0], sp.VFXColor[1], sp.VFXColor[2]}
		if accent == ([3]float32{}) {
			accent = [3]float32{0.34, 0.84, 1.0}
		}
		selected := b.playerTurn && b.armedSpellID == sp.ID && !disabled
		slots = append(slots, sdk.HotbarSlotSpec{
			Slot:        sp.Slot,
			ID:          sp.ID,
			Label:       sp.Name,
			IconName:    sp.IconName,
			CostText:    fmt.Sprintf("%d AP", sp.APCost),
			Cooldown:    cd,
			Disabled:    disabled,
			Active:      selected,
			Selected:    selected,
			Group:       group,
			KeyHint:     fmt.Sprintf("%d", sp.Slot),
			Badge:       fmt.Sprintf("%d AP", sp.APCost),
			AccentColor: accent,
			Tooltip:     fmt.Sprintf("%s | HP %d/%d (+%d) AP %d/%d MP %.1f", sp.Tooltip, b.heroHP, b.heroMaxHP, b.heroShield, b.heroAP, b.heroMaxAP, b.heroMP),
		})
	}
	slots = append(slots, sdk.HotbarSlotSpec{
		Slot:        10,
		ID:          EndTurnActionID,
		Label:       "END",
		IconName:    "",
		CostText:    "SPC",
		Disabled:    !b.playerTurn,
		Group:       "end_turn",
		KeyHint:     "SPC",
		Badge:       "END",
		AccentColor: [3]float32{0.96, 0.78, 0.28},
		Tooltip:     "End your turn and let enemies act (Space / key 0).",
	})
	_ = sdk.SetHeroHotbar(b.heroID, slots)

	// Push top-center tactical battle HUD snapshot to client plugin.
	statuses := make([]string, 0, len(b.heroStatuses))
	for _, st := range b.heroStatuses {
		statuses = append(statuses, fmt.Sprintf("%s(%d)", st.Label, st.Turns))
	}
	turnLabel := "YOUR TURN"
	prompt := "Click ground to move or select a target"
	if !b.playerTurn {
		turnLabel = "ENEMY TURN"
		prompt = "Enemies are acting..."
	} else if sp, ok := e.spells[b.armedSpellID]; ok {
		if sp.EffectKind == "dash_or_knockback" {
			prompt = fmt.Sprintf("%s armed: click ground to dash or enemy to knock back", sp.Name)
		} else {
			prompt = fmt.Sprintf("%s armed (%.1fm): select a target or click ground to move", sp.Name, sp.Range)
		}
	}
	// Build ordered Battle Timeline (Hero + enemies with active turn marker).
	heroName := b.heroName
	if heroName == "" {
		heroName = "Hero"
	}
	timeline := make([]TimelineEntry, 0, 1+len(b.mobs))
	timeline = append(timeline, TimelineEntry{
		ID:       b.heroID,
		Name:     heroName,
		Team:     "ally",
		HP:       b.heroHP,
		MaxHP:    b.heroMaxHP,
		Shield:   b.heroShield,
		Active:   b.playerTurn,
		Defeated: b.heroHP <= 0,
		Statuses: statuses,
	})
	activeMobIdx := -1
	if !b.playerTurn {
		for idx := b.aiStepIndex; idx < len(b.mobs); idx++ {
			if b.mobs[idx].hp > 0 {
				activeMobIdx = idx
				break
			}
		}
		if activeMobIdx < 0 && b.aiStepIndex > 0 && b.aiStepIndex-1 < len(b.mobs) {
			activeMobIdx = b.aiStepIndex - 1
		}
	}
	for i, m := range b.mobs {
		mStats := make([]string, 0, len(m.statuses))
		for _, st := range m.statuses {
			mStats = append(mStats, fmt.Sprintf("%s(%d)", st.Label, st.Turns))
		}
		timeline = append(timeline, TimelineEntry{
			ID:       m.actorID,
			Name:     m.name,
			Team:     "enemy",
			HP:       m.hp,
			MaxHP:    m.maxHP,
			Shield:   m.shield,
			Active:   !b.playerTurn && i == activeMobIdx,
			Defeated: m.hp <= 0,
			Statuses: mStats,
		})
	}

	_ = sdk.SendJSONToHero(b.heroID, e.cfg.HUDChannel, BattleHUDSnapshot{
		Active:     true,
		Round:      b.turnNumber,
		PlayerTurn: b.playerTurn,
		TurnLabel:  turnLabel,
		HeroHP:     b.heroHP,
		HeroMaxHP:  b.heroMaxHP,
		HeroShield: b.heroShield,
		HeroAP:     b.heroAP,
		HeroMaxAP:  b.heroMaxAP,
		HeroMP:     b.heroMP,
		HeroMaxMP:  b.heroMaxMP,
		Statuses:   statuses,
		Prompt:     prompt,
		EnemyName:  b.baseTemplateID,
		Tier:       string(b.tier),
		Timeline:   timeline,
	})

	_ = sdk.ClearGroundMarkers(b.roomX, b.roomY, b.heroID)
	if !b.playerTurn {
		return
	}
	heroKey := safeIDPart(b.heroID)
	if b.heroMP >= 0.5 {
		_ = sdk.SetGroundMarker(sdk.GroundMarkerSpec{
			ID:          "tac_mp_" + heroKey,
			RoomX:       b.roomX,
			RoomY:       b.roomY,
			Shape:       "circle",
			PosX:        b.heroPosX,
			PosZ:        b.heroPosZ,
			Radius:      b.heroMP,
			Color:       [4]float32{0.20, 0.78, 1.00, 0.34},
			OwnerHeroID: b.heroID,
		})
	}
	if sp, ok := e.spells[b.armedSpellID]; ok && sp.Range > 0 {
		zoneColor := [4]float32{0.98, 0.28, 0.24, 0.28}
		switch {
		case sp.EffectKind == "heal" || sp.EffectKind == "shield":
			zoneColor = [4]float32{0.25, 0.92, 0.56, 0.30}
		case sp.StatusID == "chill":
			zoneColor = [4]float32{0.34, 0.84, 1.00, 0.28}
		case sp.EffectKind == "dash_or_knockback":
			zoneColor = [4]float32{0.76, 0.46, 1.00, 0.28}
		}
		_ = sdk.SetGroundMarker(sdk.GroundMarkerSpec{
			ID:          "tac_rng_" + heroKey,
			RoomX:       b.roomX,
			RoomY:       b.roomY,
			Shape:       "circle",
			PosX:        b.heroPosX,
			PosZ:        b.heroPosZ,
			Radius:      sp.Range,
			Color:       zoneColor,
			Pulse:       true,
			OwnerHeroID: b.heroID,
		})
	}
	if sel := e.pickSelectedOrSoleMob(b); sel != nil && sel.hp > 0 {
		_ = sdk.SetGroundMarker(sdk.GroundMarkerSpec{
			ID:          "tac_tgt_" + heroKey,
			RoomX:       b.roomX,
			RoomY:       b.roomY,
			Shape:       "ring",
			PosX:        sel.posX,
			PosZ:        sel.posZ,
			Radius:      1.15,
			Color:       [4]float32{1.0, 0.28, 0.24, 0.95},
			Pulse:       true,
			OwnerHeroID: b.heroID,
		})
	}
}

func (e *Engine) finishBattle(b *battleState, won bool, aborted bool) {
	delete(e.battlesByHero, b.heroID)
	for _, m := range b.mobs {
		delete(e.actorToBattle, m.actorID)
		_ = sdk.RemoveActor(m.actorID)
	}
	_ = sdk.ClearGroundMarkers(b.roomX, b.roomY, b.heroID)
	_ = sdk.SetHeroHotbar(b.heroID, nil)
	_ = sdk.SetHeroControlLock(b.heroID, false)

	var goldTotal, xpTotal int64
	if won && !aborted {
		for _, m := range b.mobs {
			goldTotal += m.stats.GoldReward
			xpTotal += m.stats.XPReward
		}
		if goldTotal > 0 {
			_ = sdk.GiveGold(b.heroID, goldTotal)
		}
		if xpTotal > 0 {
			_ = sdk.GiveXP(b.heroID, xpTotal)
		}
		_ = sdk.PlayVFX(sdk.VFXSpec{
			RoomX: b.roomX, RoomY: b.roomY, Kind: "nova",
			FromX: b.heroPosX, FromY: 1.0, FromZ: b.heroPosZ,
			ToX: b.heroPosX, ToY: 1.0, ToZ: b.heroPosZ,
			Radius: 3.0, Color: [4]float32{1.0, 0.86, 0.34, 1}, DurationSec: 0.8,
		})
		_ = sdk.SendChatToHero(b.heroID, fmt.Sprintf("Victory against %s (%s) in %d turns! +%dg, +%d XP.", b.enemyDisplayName, b.tier, b.turnNumber, goldTotal, xpTotal))
	} else if !aborted {
		_ = sdk.SendChatToHero(b.heroID, fmt.Sprintf("Defeated by %s (%s) after %d turns.", b.enemyDisplayName, b.tier, b.turnNumber))
	}

	res := BattleResult{
		EncounterID: b.encounterID,
		TemplateID:  b.baseTemplateID,
		HeroID:      b.heroID,
		RoomX:       b.roomX,
		RoomY:       b.roomY,
		Won:         won && !aborted,
		Turns:       b.turnNumber,
		GoldAwarded: goldTotal,
		XPAwarded:   xpTotal,
	}

	var loot []LootDrop
	if res.Won && e.cfg.ResolveLoot != nil {
		loot = e.cfg.ResolveLoot(res)
	}

	if aborted {
		_ = sdk.SendJSONToHero(b.heroID, e.cfg.HUDChannel, BattleHUDSnapshot{Active: false})
	} else {
		enemyName := b.enemyDisplayName
		if enemyName == "" {
			enemyName = b.baseTemplateID
		}
		_ = sdk.SendJSONToHero(b.heroID, e.cfg.HUDChannel, BattleHUDSnapshot{
			Kind:   "summary",
			Active: false,
			Summary: &BattleSummarySnapshot{
				Won:        won,
				EnemyName:  enemyName,
				Difficulty: string(b.tier),
				Turns:      b.turnNumber,
				Gold:       goldTotal,
				XP:         xpTotal,
				Loot:       loot,
			},
		})
	}

	if enc, ok := e.encounters[b.encounterID]; ok {
		enc.inBattle = false
		if won && !aborted {
			enc.active = false
			enc.respawnTimer = enc.spec.RespawnSec
		} else {
			enc.active = true
			_ = e.spawnEncounterActor(enc)
		}
	}

	if e.cfg.OnBattleEnd != nil {
		e.cfg.OnBattleEnd(res)
	}
}

func findMob(b *battleState, actorID string) *mobCombatant {
	for _, m := range b.mobs {
		if m.actorID == actorID {
			return m
		}
	}
	return nil
}

func allMobsDefeated(b *battleState) bool {
	for _, m := range b.mobs {
		if m.hp > 0 {
			return false
		}
	}
	return true
}

func applyDamageToMob(m *mobCombatant, dmg int, bypassShield bool) {
	if !bypassShield && m.shield > 0 {
		if m.shield >= dmg {
			m.shield -= dmg
			return
		}
		dmg -= m.shield
		m.shield = 0
	}
	m.hp -= dmg
	if m.hp < 0 {
		m.hp = 0
	}
}

func applyDamageToHero(b *battleState, dmg int, bypassShield bool) {
	if !bypassShield && b.heroShield > 0 {
		if b.heroShield >= dmg {
			b.heroShield -= dmg
			return
		}
		dmg -= b.heroShield
		b.heroShield = 0
	}
	b.heroHP -= dmg
	if b.heroHP < 0 {
		b.heroHP = 0
	}
}

func upsertStatus(list []statusState, st statusState) []statusState {
	for i := range list {
		if list[i].ID == st.ID {
			if st.Turns > list[i].Turns {
				list[i].Turns = st.Turns
			}
			if st.Power > list[i].Power {
				list[i].Power = st.Power
			}
			return list
		}
	}
	return append(list, st)
}

func hasStatus(list []statusState, id string) bool {
	for _, st := range list {
		if st.ID == id && st.Turns > 0 {
			return true
		}
	}
	return false
}

func cleanseDebuffs(list []statusState) []statusState {
	var out []statusState
	for _, st := range list {
		if st.ID == "burn" || st.ID == "poison" || st.ID == "chill" || st.ID == "stun" {
			continue
		}
		out = append(out, st)
	}
	return out
}

func toActorStatuses(list []statusState) []sdk.ActorStatusSpec {
	if len(list) == 0 {
		return nil
	}
	out := make([]sdk.ActorStatusSpec, 0, len(list))
	for _, st := range list {
		out = append(out, sdk.ActorStatusSpec{
			ID:     st.ID,
			Label:  st.Label,
			Color:  st.Color,
			Stacks: st.Turns,
		})
	}
	return out
}

func safeIDPart(raw string) string {
	var b strings.Builder
	for i := 0; i < len(raw) && b.Len() < 24; i++ {
		c := raw[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			b.WriteByte(c)
		} else if c >= 'A' && c <= 'Z' {
			b.WriteByte(c + ('a' - 'A'))
		}
	}
	if b.Len() == 0 {
		return "h"
	}
	return b.String()
}

func dist2D(ax, az, bx, bz float32) float32 {
	return float32(math.Hypot(float64(bx-ax), float64(bz-az)))
}

func yawToward(ax, az, bx, bz float32) float32 {
	return float32(math.Atan2(float64(bx-ax), float64(bz-az)))
}

func maxF32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func minF32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
