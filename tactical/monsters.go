package tactical

// MobSpellSpec defines one special ability usable by a tactical monster.
type MobSpellSpec struct {
	ID          string
	Name        string
	APCost      int
	Cooldown    int
	Range       float32
	AoERadius   float32
	Damage      int
	Heal        int
	Shield      int
	EffectKind  string // "damage", "heal", "shield"
	StatusID    string
	StatusLabel string
	StatusTurns int
	StatusPower int
	StatusColor [3]float32
	VFXKind     string
	VFXColor    [4]float32
}

// MobTemplate defines a reusable tactical monster archetype.
type MobTemplate struct {
	ID          string
	Name        string
	Role        string // "bruiser", "caster", "stalker", "sentinel", "golem", "assassin", "acolyte", "boss"
	BaseHP      int
	BaseDamage  int
	BaseAP      int
	BaseMP      float32
	Range       float32
	GoldReward  int64
	XPReward    int64
	BodyColor   [3]float32
	TrimColor   [3]float32
	Scale       float32
	ModelID     string
	ModelPath   string
	DefaultTier DifficultyTier
	Spells      []MobSpellSpec
	Description string
}

// DefaultMonsters returns the 8 canonical tactical monster templates across combat roles.
func DefaultMonsters() []MobTemplate {
	return []MobTemplate{
		{
			ID: "ironhide_bruiser", Name: "Ironhide Bruiser", Role: "bruiser",
			BaseHP: 95, BaseDamage: 16, BaseAP: 4, BaseMP: 6.0, Range: 3.0,
			GoldReward: 25, XPReward: 40,
			BodyColor: [3]float32{0.62, 0.42, 0.30}, TrimColor: [3]float32{0.88, 0.72, 0.34}, Scale: 1.08,
			DefaultTier: TierNormal,
			Description: "Frontline heavy hitter that closes distance and delivers crushing blows.",
			Spells: []MobSpellSpec{
				{
					ID: "crushing_blow", Name: "Crushing Blow", APCost: 3, Cooldown: 2, Range: 3.2,
					Damage: 24, EffectKind: "damage", VFXKind: "slash", VFXColor: [4]float32{0.96, 0.56, 0.22, 1},
				},
			},
		},
		{
			ID: "starweave_caster", Name: "Starweave Caster", Role: "caster",
			BaseHP: 72, BaseDamage: 18, BaseAP: 5, BaseMP: 5.5, Range: 10.5,
			GoldReward: 30, XPReward: 48,
			BodyColor: [3]float32{0.24, 0.36, 0.78}, TrimColor: [3]float32{0.42, 0.90, 1.00}, Scale: 1.0,
			DefaultTier: TierNormal,
			Description: "Ranged arcanist firing starlight bolts from afar.",
			Spells: []MobSpellSpec{
				{
					ID: "astral_bolt", Name: "Astral Bolt", APCost: 3, Cooldown: 1, Range: 11.0,
					Damage: 22, EffectKind: "damage", VFXKind: "bolt", VFXColor: [4]float32{0.38, 0.86, 1.0, 1},
				},
			},
		},
		{
			ID: "venom_stalker", Name: "Venom Stalker", Role: "stalker",
			BaseHP: 80, BaseDamage: 15, BaseAP: 5, BaseMP: 7.5, Range: 3.4,
			GoldReward: 34, XPReward: 55,
			BodyColor: [3]float32{0.20, 0.56, 0.28}, TrimColor: [3]float32{0.58, 0.98, 0.32}, Scale: 1.02,
			DefaultTier: TierVeteran,
			Description: "Agile predator whose venomous fangs inflict lingering poison.",
			Spells: []MobSpellSpec{
				{
					ID: "toxic_fang", Name: "Toxic Fang", APCost: 3, Cooldown: 2, Range: 3.6,
					Damage: 16, EffectKind: "damage",
					StatusID: "poison", StatusLabel: "Poison", StatusTurns: 3, StatusPower: 8, StatusColor: [3]float32{0.38, 0.96, 0.36},
					VFXKind: "debuff", VFXColor: [4]float32{0.42, 0.96, 0.34, 1},
				},
			},
		},
		{
			ID: "frost_sentinel", Name: "Frost Sentinel", Role: "sentinel",
			BaseHP: 110, BaseDamage: 17, BaseAP: 5, BaseMP: 5.5, Range: 8.5,
			GoldReward: 42, XPReward: 65,
			BodyColor: [3]float32{0.34, 0.68, 0.92}, TrimColor: [3]float32{0.84, 0.96, 1.00}, Scale: 1.15,
			DefaultTier: TierVeteran,
			Description: "Crystalline warden that chills foes and raises frost barriers.",
			Spells: []MobSpellSpec{
				{
					ID: "glacial_shard", Name: "Glacial Shard", APCost: 3, Cooldown: 2, Range: 9.0,
					Damage: 20, EffectKind: "damage",
					StatusID: "chill", StatusLabel: "Chill", StatusTurns: 2, StatusPower: 2, StatusColor: [3]float32{0.46, 0.88, 1.0},
					VFXKind: "ring", VFXColor: [4]float32{0.46, 0.88, 1.0, 1},
				},
				{
					ID: "ice_bulwark", Name: "Ice Bulwark", APCost: 2, Cooldown: 3, Range: 0,
					Shield: 28, EffectKind: "shield",
					StatusID: "aegis", StatusLabel: "Aegis", StatusTurns: 2, StatusColor: [3]float32{0.55, 0.92, 1.0},
					VFXKind: "shield", VFXColor: [4]float32{0.55, 0.92, 1.0, 1},
				},
			},
		},
		{
			ID: "ember_golem", Name: "Ember Golem", Role: "golem",
			BaseHP: 135, BaseDamage: 21, BaseAP: 5, BaseMP: 4.8, Range: 3.6,
			GoldReward: 50, XPReward: 78,
			BodyColor: [3]float32{0.46, 0.18, 0.14}, TrimColor: [3]float32{1.00, 0.48, 0.14}, Scale: 1.28,
			DefaultTier: TierHeroic,
			Description: "Volcanic construct that ignites targets with searing magma slams.",
			Spells: []MobSpellSpec{
				{
					ID: "magma_slam", Name: "Magma Slam", APCost: 3, Cooldown: 2, Range: 4.0,
					Damage: 24, EffectKind: "damage",
					StatusID: "burn", StatusLabel: "Burn", StatusTurns: 2, StatusPower: 10, StatusColor: [3]float32{1.0, 0.44, 0.14},
					VFXKind: "nova", VFXColor: [4]float32{1.0, 0.44, 0.14, 1},
				},
				{
					ID: "molten_shell", Name: "Molten Shell", APCost: 2, Cooldown: 3, Range: 0,
					Shield: 32, EffectKind: "shield",
					VFXKind: "shield", VFXColor: [4]float32{1.0, 0.56, 0.18, 1},
				},
			},
		},
		{
			ID: "shadow_assassin", Name: "Shadow Assassin", Role: "assassin",
			BaseHP: 82, BaseDamage: 24, BaseAP: 6, BaseMP: 8.5, Range: 3.4,
			GoldReward: 48, XPReward: 74,
			BodyColor: [3]float32{0.20, 0.16, 0.30}, TrimColor: [3]float32{0.82, 0.36, 0.98}, Scale: 1.02,
			DefaultTier: TierHeroic,
			Description: "Swift flanker capable of high-burst ambushes across long strides.",
			Spells: []MobSpellSpec{
				{
					ID: "phantom_strike", Name: "Phantom Strike", APCost: 3, Cooldown: 2, Range: 4.5,
					Damage: 30, EffectKind: "damage", VFXKind: "slash", VFXColor: [4]float32{0.84, 0.34, 1.0, 1},
				},
			},
		},
		{
			ID: "dawn_acolyte", Name: "Dawn Acolyte", Role: "acolyte",
			BaseHP: 88, BaseDamage: 15, BaseAP: 5, BaseMP: 6.0, Range: 9.5,
			GoldReward: 44, XPReward: 68,
			BodyColor: [3]float32{0.88, 0.80, 0.56}, TrimColor: [3]float32{1.00, 0.90, 0.42}, Scale: 1.0,
			DefaultTier: TierVeteran,
			Description: "Radiant support caster that mends injured allies and casts smite rays.",
			Spells: []MobSpellSpec{
				{
					ID: "radiant_mend", Name: "Radiant Mend", APCost: 3, Cooldown: 2, Range: 9.0,
					Heal: 32, EffectKind: "heal", VFXKind: "heal", VFXColor: [4]float32{0.42, 0.98, 0.60, 1},
				},
				{
					ID: "smite_ray", Name: "Smite Ray", APCost: 3, Cooldown: 1, Range: 9.5,
					Damage: 19, EffectKind: "damage", VFXKind: "beam", VFXColor: [4]float32{1.0, 0.90, 0.45, 1},
				},
			},
		},
		{
			ID: "astral_overlord", Name: "Astral Overlord", Role: "boss",
			BaseHP: 190, BaseDamage: 26, BaseAP: 6, BaseMP: 6.5, Range: 10.0,
			GoldReward: 120, XPReward: 180,
			BodyColor: [3]float32{0.24, 0.14, 0.44}, TrimColor: [3]float32{0.34, 0.92, 1.00}, Scale: 1.42,
			DefaultTier: TierMythic,
			Description: "Apex celestial warden commanding eclipse beams, void novas, and aegis shields.",
			Spells: []MobSpellSpec{
				{
					ID: "eclipse_beam", Name: "Eclipse Beam", APCost: 3, Cooldown: 2, Range: 11.0,
					Damage: 28, EffectKind: "damage",
					StatusID: "burn", StatusLabel: "Burn", StatusTurns: 2, StatusPower: 12, StatusColor: [3]float32{0.96, 0.44, 0.24},
					VFXKind: "beam", VFXColor: [4]float32{0.86, 0.34, 1.0, 1},
				},
				{
					ID: "void_nova", Name: "Void Nova", APCost: 4, Cooldown: 3, Range: 8.5, AoERadius: 4.5,
					Damage: 32, EffectKind: "damage", VFXKind: "nova", VFXColor: [4]float32{0.68, 0.28, 1.0, 1},
				},
				{
					ID: "celestial_aegis", Name: "Celestial Aegis", APCost: 2, Cooldown: 3, Range: 0,
					Shield: 45, EffectKind: "shield",
					StatusID: "aegis", StatusLabel: "Aegis", StatusTurns: 2, StatusColor: [3]float32{0.34, 0.92, 1.0},
					VFXKind: "shield", VFXColor: [4]float32{0.34, 0.92, 1.0, 1},
				},
			},
		},
	}
}

// MonsterByID looks up one of the default tactical monster templates by ID.
func MonsterByID(id string) (MobTemplate, bool) {
	for _, m := range DefaultMonsters() {
		if m.ID == id {
			return m, true
		}
	}
	return MobTemplate{}, false
}
