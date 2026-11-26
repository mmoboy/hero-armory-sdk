package tactical

// EndTurnActionID is the action ID bound to hotbar slot 10 (key 0).
const EndTurnActionID = "end_turn"

// SpellSpec defines one tactical player spell mounted on hotbar slots 1..8.
type SpellSpec struct {
	Slot        int
	ID          string
	Name        string
	IconName    string
	APCost      int
	Cooldown    int
	MinRange    float32
	Range       float32
	AoERadius   float32
	BaseDamage  int
	CritChance  float32
	CritMult    float32
	TargetKind  string // "enemy", "self", "ground"
	EffectKind  string // "damage", "shield", "heal", "dash_or_knockback"
	StatusID    string
	StatusLabel string
	StatusTurns int
	StatusPower int
	StatusColor [3]float32
	CritStun    bool
	Cleanse     bool
	Knockback   float32
	VFXKind     string
	VFXColor    [4]float32
	HeroPose    string
	Tooltip     string
}

// DefaultSpells returns the 8 default tactical player spells for slots 1..8.
func DefaultSpells() []SpellSpec {
	return []SpellSpec{
		{
			Slot:       1,
			ID:         "heavy_strike",
			Name:       "Heavy Strike",
			IconName:   "ha_spell_strike",
			APCost:     2,
			Cooldown:   0,
			Range:      3.4,
			BaseDamage: 28,
			CritChance: 0.15,
			CritMult:   1.6,
			TargetKind: "enemy",
			EffectKind: "damage",
			VFXKind:    "slash",
			VFXColor:   [4]float32{1.0, 0.82, 0.32, 1.0},
			HeroPose:   "slash",
			Tooltip:    "Heavy melee cleave (2 AP, 3.4m, 28 dmg).",
		},
		{
			Slot:       2,
			ID:         "arcane_bolt",
			Name:       "Arcane Bolt",
			IconName:   "ha_spell_bolt",
			APCost:     3,
			Cooldown:   0,
			Range:      12.0,
			BaseDamage: 30,
			CritChance: 0.14,
			CritMult:   1.55,
			TargetKind: "enemy",
			EffectKind: "damage",
			VFXKind:    "bolt",
			VFXColor:   [4]float32{0.34, 0.84, 1.0, 1.0},
			HeroPose:   "channel",
			Tooltip:    "Ranged arcane projectile (3 AP, 12m, 30 dmg).",
		},
		{
			Slot:       3,
			ID:         "starlight_nova",
			Name:       "Starlight Nova",
			IconName:   "ha_spell_nova",
			APCost:     4,
			Cooldown:   2,
			Range:      8.5,
			AoERadius:  4.0,
			BaseDamage: 28,
			CritChance: 0.12,
			CritMult:   1.5,
			TargetKind: "enemy",
			EffectKind: "damage",
			VFXKind:    "nova",
			VFXColor:   [4]float32{0.78, 0.56, 1.0, 1.0},
			HeroPose:   "heroic",
			Tooltip:    "Explosive starlight blast (4 AP, 8.5m, 4m AoE, CD 2).",
		},
		{
			Slot:        4,
			ID:          "ember_lance",
			Name:        "Ember Lance",
			IconName:    "ha_spell_bolt",
			APCost:      3,
			Cooldown:    2,
			Range:       11.0,
			BaseDamage:  24,
			CritChance:  0.14,
			CritMult:    1.55,
			TargetKind:  "enemy",
			EffectKind:  "damage",
			StatusID:    "burn",
			StatusLabel: "Burn",
			StatusTurns: 2,
			StatusPower: 10,
			StatusColor: [3]float32{1.0, 0.44, 0.16},
			VFXKind:     "beam",
			VFXColor:    [4]float32{1.0, 0.46, 0.18, 1.0},
			HeroPose:    "channel",
			Tooltip:     "Scorching beam (3 AP, 11m, 24 dmg + Burn 10/turn for 2t, CD 2).",
		},
		{
			Slot:        5,
			ID:          "frost_bind",
			Name:        "Frost Bind",
			IconName:    "ha_spell_nova",
			APCost:      3,
			Cooldown:    2,
			Range:       10.5,
			BaseDamage:  22,
			CritChance:  0.22,
			CritMult:    1.5,
			TargetKind:  "enemy",
			EffectKind:  "damage",
			StatusID:    "chill",
			StatusLabel: "Chill",
			StatusTurns: 2,
			StatusPower: 2,
			StatusColor: [3]float32{0.42, 0.86, 1.0},
			CritStun:    true,
			VFXKind:     "ring",
			VFXColor:    [4]float32{0.45, 0.88, 1.0, 1.0},
			HeroPose:    "channel",
			Tooltip:     "Freezing ring (3 AP, 10.5m, 22 dmg + Chill -2 MP; Stuns on crit, CD 2).",
		},
		{
			Slot:        6,
			ID:          "aegis_ward",
			Name:        "Aegis Ward",
			IconName:    "shield",
			APCost:      2,
			Cooldown:    3,
			Range:       0,
			BaseDamage:  36,
			TargetKind:  "self",
			EffectKind:  "shield",
			StatusID:    "aegis",
			StatusLabel: "Aegis",
			StatusTurns: 2,
			StatusColor: [3]float32{0.34, 0.88, 1.0},
			VFXKind:     "shield",
			VFXColor:    [4]float32{0.36, 0.90, 1.0, 1.0},
			HeroPose:    "heroic",
			Tooltip:     "Fortify self with +36 Shield and Aegis (-20% dmg taken for 2t, 2 AP, CD 3).",
		},
		{
			Slot:       7,
			ID:         "mending_light",
			Name:       "Mending Light",
			IconName:   "star",
			APCost:     3,
			Cooldown:   3,
			Range:      0,
			BaseDamage: 42,
			TargetKind: "self",
			EffectKind: "heal",
			Cleanse:    true,
			VFXKind:    "heal",
			VFXColor:   [4]float32{0.32, 0.96, 0.56, 1.0},
			HeroPose:   "channel",
			Tooltip:    "Restore +42 HP and cleanse active debuffs (3 AP, CD 3).",
		},
		{
			Slot:       8,
			ID:         "gale_rush",
			Name:       "Gale Rush",
			IconName:   "compass",
			APCost:     2,
			Cooldown:   2,
			Range:      7.5,
			BaseDamage: 18,
			CritChance: 0.10,
			CritMult:   1.5,
			TargetKind: "ground",
			EffectKind: "dash_or_knockback",
			Knockback:  3.5,
			VFXKind:    "buff",
			VFXColor:   [4]float32{0.46, 0.98, 0.78, 1.0},
			HeroPose:   "slash",
			Tooltip:    "Click ground to dash up to 7.5m, or click an enemy to knock back 3.5m + 18 dmg (2 AP, CD 2).",
		},
	}
}
