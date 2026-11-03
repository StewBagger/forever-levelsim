package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank the mage knows at its level is registered, not every rank
// spellData.ArcaneWrath lists.
func (mage *Mage) registerArcaneWrathSpell() {
	index, rank := mage.learnedRank(spellData.ArcaneWrath, spellData.ArcaneWrath.Highest())
	if rank == spelldata.Nil {
		return
	}
	mage.registerArcaneWrathRank(index, rank)
}

// Arcane Wrath
// https://www.wowhead.com/forever/spell=1326082
//
// "Launches a projectile of pure Arcane energy at the enemy target, causing $s1 Arcane damage. This
// spell gains additional Spell Damage equal to $s2% of your maximum Mana." Effect 2 holds the share
// of maximum mana, which counts as spell damage of this spell alone, scaled by the damage effect's
// coefficient like the rest.
func (mage *Mage) registerArcaneWrathRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()
	maxManaShare := rank.EffectN(2).Percent()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneWrath

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		bonus := maxManaShare * mage.MaxMana()

		spell.BonusSpellDamage += bonus
		result := spell.CalcDamage(sim, target, damage.AverageAt(spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		spell.BonusSpellDamage -= bonus

		spell.DealDamageAfterTravel(sim, result)
	}

	mage.RegisterSpell(config)
}
