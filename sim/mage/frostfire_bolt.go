package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank the mage knows at its level is registered, not every rank
// spellData.FrostfireBolt lists.
func (mage *Mage) registerFrostfireBoltSpell() {
	index, rank := mage.learnedRank(spellData.FrostfireBolt, spellData.FrostfireBolt.Highest())
	if rank == spelldata.Nil {
		return
	}
	mage.registerFrostfireBoltRank(index, rank)
}

// The DoT (effect 3) is its own spell, like Fireball's.
// The slow (effect 1) does nothing to a boss.
func (mage *Mage) registerFrostfireBoltRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	dotConfig := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFrostfireBoltDot
	tick := rank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(rank, tick)
	dotConfig.Dot.OnTick = periodicDamageTick(tick, rank.TickOutcome)
	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}
	dotSpell := mage.RegisterSpell(dotConfig)

	// Binary: a spell with a non-damage effect (the slow) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellFrostfireBolt

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.AverageAt(spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		spell.DealCastResult(sim, result)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				dotSpell.Cast(sim, target)
			}
		})
	}

	mage.RegisterSpell(config).RelatedDotSpell = dotSpell
}
