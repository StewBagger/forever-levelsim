package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 8 is the last rank the trainer teaches; rank 9 comes from the Handbook of Backstab IX book.
var backstabTrainerRank = spellData.Backstab.ByID(11281)
var backstabBookRank = spellData.Backstab.ByID(25300)

// Below the book's level both resolve to the same learned rank, which registers once.
func (rogue *Rogue) registerBackstabSpell() {
	trainer := spelldata.LearnedRank(&rogue.Unit, spellData.Backstab, backstabTrainerRank)
	book := spelldata.LearnedRank(&rogue.Unit, spellData.Backstab, backstabBookRank)
	if trainer != spelldata.Nil {
		rogue.registerBackstabRank(trainer)
	}
	if book != spelldata.Nil && book.ID != trainer.ID {
		rogue.registerBackstabRank(book)
	}
}

func (rogue *Rogue) registerBackstabRank(rank *spelldata.Spell) {
	backstabBaseDamage := rank.DamageEffect().AverageAt(rogue.Level)
	comboPointGen := rank.EnergizeEffect().AverageAt(rogue.Level)
	config := spelldata.SpellConfig(&rogue.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder | SpellFlagColdBlooded
	config.ClassSpellMask = RogueSpellBackstab
	config.DamageMultiplier = rank.EffectN(2).Percent()
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !rogue.PseudoStats.InFrontOfTarget && rogue.HasDagger(core.MainHand)
	}

	// The row lacks No Active Defense, so the target can dodge it; from behind it is never parried or blocked.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := backstabBaseDamage +
			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(comboPointGen), spell.ComboPointMetrics())
			chanceForBonus := spellData.PuncturingWounds.EffectAt(2).FractionAt(rogue.Talents.PuncturingWounds)
			if sim.Proc(chanceForBonus, "Puncturing Wounds") {
				rogue.AddComboPoints(sim, 1, rogue.puncturingWoundsComboPointMetrics)
			}
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.RegisterSpell(config)
}
