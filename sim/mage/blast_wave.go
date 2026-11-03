package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank the mage knows at its level is registered, not every rank
// spellData.BlastWave lists.
func (mage *Mage) registerBlastWaveSpell() {
	index, rank := mage.learnedRank(spellData.BlastWave, spellData.BlastWave.Highest())
	if rank == spelldata.Nil {
		return
	}
	mage.registerBlastWaveRank(index, rank)
}

// The daze (effect 2) does nothing to a boss.
func (mage *Mage) registerBlastWaveRank(_ int32, rank *spelldata.Spell) {
	// Binary: a spell with a non-damage effect (the daze) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellBlastWave

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(rank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}
