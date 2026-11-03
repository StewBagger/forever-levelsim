package core

import (
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// The damage profile's defaults where a field is 0.
const (
	DamageProfileTankHitSeconds     = 2.0 // A boss's usual swing timer.
	DamageProfileGroupCadenceSecond = 8.0 // Seconds between group pulses.
	damageProfileMinGroupInterval   = 0.5 // The shortest interval a cadence roll can give.
)

// Registers the encounter's damage profile on the first target, which deals it to the raid as its
// own spells, through the struck unit's armor, resistances, absorbs and damage taken modifiers.
func (env *Environment) applyDamageProfile(profile *proto.DamageProfile) {
	if profile == nil || len(env.Encounter.AllTargetUnits) == 0 {
		return
	}
	boss := env.Encounter.AllTargetUnits[0]
	if profile.TankDtps > 0 && env.Raid.tank != nil {
		env.registerTankDamage(boss, env.Raid.tank, profile)
	}
	if profile.GroupDamage > 0 {
		env.registerGroupDamage(boss, profile)
	}
}

// A spell of the boss that deals *amount, read at each hit. It skips the cast, so the boss's own casts
// and GCD never hold it back.
func damageProfileSpell(boss *Unit, otherID proto.OtherAction, school SpellSchool, amount *float64) *Spell {
	return boss.RegisterSpell(SpellConfig{
		ActionID:         ActionID{OtherID: otherID},
		SpellSchool:      school,
		ProcMask:         ProcMaskEmpty,
		Flags:            SpellFlagNoOnCastComplete,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, *amount, spell.OutcomeAlwaysHit)
		},
	})
}

// Physical hits on the tank every tank_hit_seconds, each tank_dtps times that, from a random first
// swing within one interval of the pull.
func (env *Environment) registerTankDamage(boss *Unit, tank *Unit, profile *proto.DamageProfile) {
	interval := profile.TankHitSeconds
	if interval <= 0 {
		interval = DamageProfileTankHitSeconds
	}
	amount := profile.TankDtps * interval
	spell := damageProfileSpell(boss, proto.OtherAction_OtherActionTankDamage, SpellSchoolPhysical, &amount)

	boss.RegisterResetEffect(func(sim *Simulation) {
		pa := &PendingAction{
			NextActionAt: max(sim.CurrentTime, 0) + DurationFromSeconds(interval*sim.RandomFloat("Tank Damage Timing")),
		}
		pa.OnAction = func(sim *Simulation) {
			if tank.isHealableAlly() {
				spell.SkipCastAndApplyEffects(sim, tank)
			}
			pa.NextActionAt = sim.CurrentTime + DurationFromSeconds(interval)
			sim.AddPendingAction(pa)
		}
		sim.AddPendingAction(pa)
	})
}

// A pulse every group_cadence_seconds (± group_cadence_variation) on group_targets living raid
// members picked at random, each struck for group_damage (± group_damage_variation of it).
func (env *Environment) registerGroupDamage(boss *Unit, profile *proto.DamageProfile) {
	cadence := profile.GroupCadenceSeconds
	if cadence <= 0 {
		cadence = DamageProfileGroupCadenceSecond
	}
	var amount float64
	spell := damageProfileSpell(boss, proto.OtherAction_OtherActionGroupDamage, SpellSchoolFromProto(profile.GroupSchool), &amount)
	tank := env.Raid.tank

	nextInterval := func(sim *Simulation) time.Duration {
		roll := cadence + profile.GroupCadenceVariation*(2*sim.RandomFloat("Group Damage Cadence")-1)
		return DurationFromSeconds(max(roll, damageProfileMinGroupInterval))
	}

	var candidates []*Unit
	boss.RegisterResetEffect(func(sim *Simulation) {
		pa := &PendingAction{
			NextActionAt: max(sim.CurrentTime, 0) + DurationFromSeconds(cadence*sim.RandomFloat("Group Damage Timing")),
		}
		pa.OnAction = func(sim *Simulation) {
			candidates = candidates[:0]
			for _, unit := range env.Raid.AllUnits {
				if unit.isHealableAlly() && (profile.GroupHitsTank || unit != tank) {
					candidates = append(candidates, unit)
				}
			}
			count := len(candidates)
			if profile.GroupTargets > 0 {
				count = min(count, int(profile.GroupTargets))
			}
			// A partial shuffle draws the struck units without repeats.
			for i := 0; i < count; i++ {
				j := i + int(sim.RandomFloat("Group Damage Target")*float64(len(candidates)-i))
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
			for _, target := range candidates[:count] {
				amount = profile.GroupDamage * (1 + profile.GroupDamageVariation*(2*sim.RandomFloat("Group Damage Roll")-1))
				spell.SkipCastAndApplyEffects(sim, target)
			}
			pa.NextActionAt = sim.CurrentTime + nextInterval(sim)
			sim.AddPendingAction(pa)
		}
		sim.AddPendingAction(pa)
	})
}
