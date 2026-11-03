package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank the mage knows at its level is registered, not every rank
// spellData.Pyroblast lists.
func (mage *Mage) registerPyroblastSpell() {
	index, rank := mage.learnedRank(spellData.Pyroblast, spellData.Pyroblast.Highest())
	if rank == spelldata.Nil {
		return
	}
	mage.registerPyroblastRank(index, rank)
}

// Pyroblast
// https://www.wowhead.com/forever/spell=18809
//
// The periodic damage is a separate spell (MageSpellPyroblastDot) so that class-mask mods can
// tell it from the hit.
func (mage *Mage) registerPyroblastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	dotConfig := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellPyroblastDot

	tick := rank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(rank, tick)
	dotConfig.Dot.OnTick = periodicDamageTick(tick, rank.TickOutcome)

	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	pyroblastDot := mage.RegisterSpell(dotConfig)

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellPyroblast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.AverageAt(spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		spell.DealCastResult(sim, result)
		spell.WaitTravelTime(sim, func(s *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				pyroblastDot.Cast(sim, target)
			}
		})
	}

	mage.RegisterSpell(config).RelatedDotSpell = pyroblastDot
}
