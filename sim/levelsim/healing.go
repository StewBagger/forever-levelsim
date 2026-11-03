package levelsim

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// A healer's default fight at a level: a 5-man dungeon group, the healer and four target dummies, one
// of them the tank, with incoming damage derived from ExpectedStat at the level. The constants are a
// starting assumption, to be tuned against combat logs.
const (
	HealingGroupDummies = 4 // The tank and three damage dealers beside the healer.

	// The mobs whose melee is on the tank at once: a dungeon trash pull, or a boss and its adds.
	HealingGroupPullSize = 3
	// An elite's melee over an ExpectedStat creature's of its level (CreatureAutoAttackDps): dungeon
	// mobs are elite, and an elite of the level hits about twice as hard as a normal one.
	HealingGroupEliteDamage = 2.0
	// The share of the pull's melee that reaches the tank's health once its dodge, parry, block and
	// armor have taken theirs.
	HealingGroupTankDamageTaken = 0.35
	// The tank's health over ExpectedStat's player health at the level, for a tank's stamina.
	HealingGroupTankHealth = 1.5

	// Each group pulse takes this share of a group member's health: a cleave, a breath or an area spell.
	HealingGroupPulseHealthShare = 0.10
	HealingGroupPulseVariation   = 0.2 // ±20% on each pulse hit.
	HealingGroupPulseSeconds     = 8.0 // Seconds between pulses, about a dungeon boss's ability cadence.
	HealingGroupPulseJitter      = 2.0 // ± seconds on each interval.
	HealingGroupPulseTargets     = 2   // Group members each pulse strikes.

	// How far a dungeon boss sits above the group.
	HealingGroupBossLevels = 2
)

// The tank's damage per second at a level, before the dummy's own armor (it has none):
// pull size × elite multiplier × ExpectedStat creature auto attack dps × the share that lands.
func HealingGroupTankDtps(level int32) float64 {
	_, _, autoAttackDps := core.CreatureStatsAtLevel(level)
	return HealingGroupPullSize * HealingGroupEliteDamage * autoAttackDps * HealingGroupTankDamageTaken
}

// The damage each group pulse deals each unit it strikes at a level.
func HealingGroupPulseDamage(level int32) float64 {
	return HealingGroupPulseHealthShare * core.PlayerHealthAtLevel(level)
}

// The incoming damage of the level's default healing fight.
func HealingGroupDamageProfile(level int32) *proto.DamageProfile {
	return &proto.DamageProfile{
		TankDtps:              HealingGroupTankDtps(level),
		TankHitSeconds:        core.DamageProfileTankHitSeconds,
		GroupDamage:           HealingGroupPulseDamage(level),
		GroupDamageVariation:  HealingGroupPulseVariation,
		GroupCadenceSeconds:   HealingGroupPulseSeconds,
		GroupCadenceVariation: HealingGroupPulseJitter,
		GroupTargets:          HealingGroupPulseTargets,
		GroupSchool:           proto.SpellSchool_SpellSchoolFire,
	}
}

// The level's group members: healable dummies of a player's health at the level, the tank with
// HealingGroupTankHealth of it, all of them able to die.
func HealingGroupDummySettings(level int32) *proto.TargetDummySettings {
	return &proto.TargetDummySettings{
		Healable:      true,
		TankMaxHealth: HealingGroupTankHealth * core.PlayerHealthAtLevel(level),
		TrackDeaths:   true,
	}
}

// The tank slot of a single-player raid with the group's dummies: the first dummy takes raid index 1.
func HealingGroupTanks() []*proto.UnitReference {
	return []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 1}}
}

// A boss for the healing fight: HealingGroupBossLevels above the level, with no melee of its own, so
// the damage profile is all the damage the group takes.
func HealingGroupTarget(level int32) *proto.Target {
	return &proto.Target{
		Level:     level + HealingGroupBossLevels,
		MobType:   proto.MobType_MobTypeHumanoid,
		TankIndex: 0,
	}
}

// Gives a stat weights request for a single healer the level's healing group where it has none: the
// dummies and the tank slot, the damage profile on its encounter, and effective healing as the HPS it
// weighs.
func WithHealingGroup(swr *proto.StatWeightsRequest, level int32) {
	if swr.TargetDummies == 0 {
		swr.TargetDummies = HealingGroupDummies
		swr.TargetDummySettings = HealingGroupDummySettings(level)
		swr.Tanks = HealingGroupTanks()
	}
	if swr.Encounter == nil {
		swr.Encounter = &proto.Encounter{Duration: 180, Targets: []*proto.Target{HealingGroupTarget(level)}}
	}
	if swr.Encounter.DamageProfile == nil {
		swr.Encounter.DamageProfile = HealingGroupDamageProfile(level)
	}
	swr.EffectiveHealing = true
}
