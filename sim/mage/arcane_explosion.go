package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank the mage knows at its level is registered, not every rank
// spellData.ArcaneExplosion lists.
func (mage *Mage) registerArcaneExplosionSpell() {
	index, rank := mage.learnedRank(spellData.ArcaneExplosion, spellData.ArcaneExplosion.Highest())
	if rank == spelldata.Nil {
		return
	}
	mage.registerArcaneExplosionRank(index, rank)
}

func (mage *Mage) registerArcaneExplosionRank(_ int32, rank *spelldata.Spell) {
	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneExplosion

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(rank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}
