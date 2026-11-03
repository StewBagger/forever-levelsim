package core

import (
	"math"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// BaseStatsAtLevel is a race and class's base stats at a level. At CharacterLevel it is BaseStats, as
// upstream keeps it. Below, the class attributes, health and mana are the generated per-level tables
// (level_stats_auto_gen.go), the race offset is the same at every level, and attack power follows the
// Classic level formulas that give ClassBaseStats' level 60 values. Base crit stays the class's level
// 60 value: no game file states it by level.
func BaseStatsAtLevel(race proto.Race, class proto.Class, level int32) stats.Stats {
	if level >= CharacterLevel || level < 1 {
		return BaseStats[BaseStatsKey{Race: race, Class: class}]
	}
	base := classBaseStatsByLevel[class][level]
	// The attribute curves are Classic's. Where Forever's level 60 value differs (shaman health), the
	// curve is scaled to end on it.
	at60 := classBaseStatsByLevel[class][CharacterLevel]
	for _, s := range []stats.Stat{stats.Health, stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit} {
		if want := ClassBaseStats[class][s]; at60[s] != 0 && at60[s] != want {
			base[s] = math.Round(base[s] * want / at60[s])
		}
	}
	base[stats.AttackPower] = baseAttackPowerAtLevel(class, level)
	if class == proto.Class_ClassHunter {
		base[stats.RangedAttackPower] = float64(level)*2 - 10
	}
	extra := ExtraClassBaseStats[class]
	extra[stats.Mana] = 0
	return base.Add(RaceOffsets[race]).Add(extra)
}

// The base attack power of a class at a level, before any attribute converts into it.
func baseAttackPowerAtLevel(class proto.Class, level int32) float64 {
	l := float64(level)
	switch class {
	case proto.Class_ClassWarrior, proto.Class_ClassPaladin:
		return l*3 - 20
	case proto.Class_ClassHunter, proto.Class_ClassRogue, proto.Class_ClassShaman:
		return l*2 - 20
	case proto.Class_ClassDruid:
		return -20
	default:
		return -10
	}
}

func clampLevel(level int32) int32 {
	return min(max(level, 1), CharacterLevel)
}

// CritPerAgiAtLevel is melee crit percent per point of Agility for a class at a level.
func CritPerAgiAtLevel(class proto.Class, level int32) float64 {
	return critPerAgiByLevel[class][clampLevel(level)]
}

// CritPerIntAtLevel is spell crit percent per point of Intellect for a class at a level.
func CritPerIntAtLevel(class proto.Class, level int32) float64 {
	return critPerIntByLevel[class][clampLevel(level)]
}

// CreatureStatsAtLevel is ExpectedStat's creature health, armor and auto attack dps at a level, the base
// a pet's stats scale from. At CharacterLevel it is the CreatureHealthMaxLevel, CreatureArmorMaxLevel and
// CreatureAutoAttackDpsMaxLevel constants.
func CreatureStatsAtLevel(level int32) (health, armor, autoAttackDps float64) {
	row := creatureStatsByLevel[clampLevel(level)]
	return row.Health, row.Armor, row.AutoAttackDps
}
