package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank the mage knows at its level is registered, not every rank
// spellData.IceLance lists.
func (mage *Mage) registerIceLanceSpell() {
	index, rank := mage.learnedRank(spellData.IceLance, spellData.IceLance.Highest())
	if rank == spelldata.Nil {
		return
	}
	mage.registerIceLanceRank(index, rank)
}

// Ice Lance
// https://www.wowhead.com/forever/spell=1240047
//
// Frost damage to one enemy, raised by effect 2's percentage ("Deals $s2% increased damage to Frozen
// targets", 300) while the target is frozen. Its 8 sec cooldown runs on category 19, which
// Fire Blast shares, so casting either puts both on cooldown.
//
// The damage effect states no spell power coefficient; the store carries one as an override
// (tools/database/overrides). Effects 3 (20) and 4 (5) are server-side scripts the tooltip does not
// name, and are not modelled.
func (mage *Mage) registerIceLanceRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()
	frozenMultiplier := 1 + rank.EffectN(2).Percent()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellIceLance

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		multiplier := 1.0
		if mage.isFrozen(target) {
			multiplier = frozenMultiplier
		}

		spell.DamageMultiplier *= multiplier
		result := spell.CalcDamage(sim, target, damage.AverageAt(spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= multiplier

		spell.DealDamageAfterTravel(sim, result)
	}

	mage.RegisterSpell(config)
}
