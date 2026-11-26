// Package tactical provides a turnkey, plugin-side turn-based tactical combat
// system built purely on hero-armory-sdk API Level 5 primitives.
package tactical

import (
	"math"
	"strings"
)

// DifficultyTier selects a preset difficulty curve for tactical encounters.
type DifficultyTier string

const (
	TierStory   DifficultyTier = "story"
	TierNormal  DifficultyTier = "normal"
	TierVeteran DifficultyTier = "veteran"
	TierHeroic  DifficultyTier = "heroic"
	TierMythic  DifficultyTier = "mythic"
)

// ParseDifficulty parses a case-insensitive difficulty tier name.
func ParseDifficulty(name string) (DifficultyTier, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "story":
		return TierStory, true
	case "normal":
		return TierNormal, true
	case "veteran":
		return TierVeteran, true
	case "heroic":
		return TierHeroic, true
	case "mythic":
		return TierMythic, true
	default:
		return "", false
	}
}

// MobBalance exposes explicit tuning knobs for scaling one mob or a whole realm.
type MobBalance struct {
	HPMult     float32
	DamageMult float32
	ShieldPct  float32
	APBonus    int
	MPBonus    float32
	CritChance float32
	CritMult   float32
	ArmorPct   float32
	RewardMult float32
	AIProfile  string // "basic", "tactical", "ruthless"
}

// PresetBalance returns the canonical MobBalance preset for tier.
func PresetBalance(tier DifficultyTier) MobBalance {
	switch tier {
	case TierStory:
		return MobBalance{
			HPMult: 0.70, DamageMult: 0.65, ShieldPct: 0.0,
			APBonus: 0, MPBonus: 0.0,
			CritChance: 0.03, CritMult: 1.35, ArmorPct: 0.0,
			RewardMult: 0.85, AIProfile: "basic",
		}
	case TierVeteran:
		return MobBalance{
			HPMult: 1.35, DamageMult: 1.25, ShieldPct: 0.15,
			APBonus: 1, MPBonus: 1.0,
			CritChance: 0.14, CritMult: 1.60, ArmorPct: 0.12,
			RewardMult: 1.35, AIProfile: "tactical",
		}
	case TierHeroic:
		return MobBalance{
			HPMult: 1.80, DamageMult: 1.55, ShieldPct: 0.25,
			APBonus: 2, MPBonus: 1.5,
			CritChance: 0.20, CritMult: 1.75, ArmorPct: 0.20,
			RewardMult: 1.80, AIProfile: "tactical",
		}
	case TierMythic:
		return MobBalance{
			HPMult: 2.50, DamageMult: 1.95, ShieldPct: 0.35,
			APBonus: 3, MPBonus: 2.5,
			CritChance: 0.28, CritMult: 1.90, ArmorPct: 0.28,
			RewardMult: 2.60, AIProfile: "ruthless",
		}
	default: // TierNormal
		return MobBalance{
			HPMult: 1.00, DamageMult: 1.00, ShieldPct: 0.05,
			APBonus: 0, MPBonus: 0.0,
			CritChance: 0.08, CritMult: 1.50, ArmorPct: 0.05,
			RewardMult: 1.00, AIProfile: "basic",
		}
	}
}

// ScaledMobStats holds the final computed combat stats for one spawned mob.
type ScaledMobStats struct {
	Level      int
	MaxHP      int
	Shield     int
	Damage     int
	AP         int
	MP         float32
	CritChance float32
	CritMult   float32
	ArmorPct   float32
	GoldReward int64
	XPReward   int64
	AIProfile  string
}

// ScaleMob computes ScaledMobStats for tmpl at level using bal.
func ScaleMob(tmpl MobTemplate, level int, bal MobBalance) ScaledMobStats {
	if level < 1 {
		level = 1
	}
	hpMult := bal.HPMult
	if hpMult <= 0 {
		hpMult = 1.0
	}
	dmgMult := bal.DamageMult
	if dmgMult <= 0 {
		dmgMult = 1.0
	}
	rewMult := bal.RewardMult
	if rewMult <= 0 {
		rewMult = 1.0
	}
	critMult := bal.CritMult
	if critMult < 1.0 {
		critMult = 1.5
	}
	critChance := clampF32(bal.CritChance, 0, 0.75)
	armorPct := clampF32(bal.ArmorPct, 0, 0.65)
	shieldPct := clampF32(bal.ShieldPct, 0, 1.0)

	lvlF := float64(level - 1)
	hpScale := (1.0 + 0.14*lvlF) * float64(hpMult)
	dmgScale := (1.0 + 0.11*lvlF) * float64(dmgMult)
	rewScale := (1.0 + 0.12*lvlF) * float64(rewMult)

	baseHP := tmpl.BaseHP
	if baseHP < 10 {
		baseHP = 50
	}
	baseDmg := tmpl.BaseDamage
	if baseDmg < 1 {
		baseDmg = 10
	}
	baseAP := tmpl.BaseAP
	if baseAP < 2 {
		baseAP = 4
	}
	baseMP := tmpl.BaseMP
	if baseMP < 2 {
		baseMP = 5.0
	}

	maxHP := int(math.Round(float64(baseHP) * hpScale))
	if maxHP < 10 {
		maxHP = 10
	}
	shield := int(math.Round(float64(maxHP) * float64(shieldPct)))
	dmg := int(math.Round(float64(baseDmg) * dmgScale))
	if dmg < 1 {
		dmg = 1
	}
	ap := baseAP + bal.APBonus
	if ap < 2 {
		ap = 2
	}
	mp := baseMP + bal.MPBonus
	if mp < 2.0 {
		mp = 2.0
	}
	gold := int64(math.Round(float64(tmpl.GoldReward) * rewScale))
	xp := int64(math.Round(float64(tmpl.XPReward) * rewScale))
	ai := bal.AIProfile
	if ai != "basic" && ai != "tactical" && ai != "ruthless" {
		ai = "basic"
	}

	return ScaledMobStats{
		Level:      level,
		MaxHP:      maxHP,
		Shield:     shield,
		Damage:     dmg,
		AP:         ap,
		MP:         mp,
		CritChance: critChance,
		CritMult:   critMult,
		ArmorPct:   armorPct,
		GoldReward: gold,
		XPReward:   xp,
		AIProfile:  ai,
	}
}

func clampF32(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
