package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank the mage knows at its level is registered, not every rank
// spellData.FrostNova lists.
func (mage *Mage) registerFrostNovaSpell() {
	index, rank := mage.learnedRank(spellData.FrostNova, spellData.FrostNova.Highest())
	if rank == spelldata.Nil {
		return
	}
	mage.registerFrostNovaRank(index, rank)
}

// Frost Nova
// https://www.wowhead.com/forever/spell=10230
//
// Frost damage to every enemy around the caster, and a root. The root does nothing to a boss the
// sim fights, so only the damage is modelled.
func (mage *Mage) registerFrostNovaRank(_ int32, rank *spelldata.Spell) {
	// Binary: a spell with a non-damage effect (the root) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellFrostNova

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(rank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}
