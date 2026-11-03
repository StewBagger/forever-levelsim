package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank the mage knows at its level is registered, not every rank
// spellData.ArcaneBlast lists.
func (mage *Mage) registerArcaneBlastSpell() {
	index, rank := mage.learnedRank(spellData.ArcaneBlast, spellData.ArcaneBlast.Highest())
	if rank == spelldata.Nil {
		return
	}
	mage.registerArcaneBlastRank(index, rank)
}

func (mage *Mage) registerArcaneBlastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.AverageAt(spell.Unit.Level), spell.OutcomeMagicHitAndCrit)

		if !mage.ArcaneChargesAura.IsActive() {
			mage.ArcaneChargesAura.Activate(sim)
		}
		mage.ArcaneChargesAura.AddStack(sim)
	}

	mage.RegisterSpell(config)
}
