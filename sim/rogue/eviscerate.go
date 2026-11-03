package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 8 is the last rank the trainer teaches; rank 9 comes from the Manual of Eviscerate IX book.
var eviscerateTrainerRank = spellData.Eviscerate.ByID(11300)
var eviscerateBookRank = spellData.Eviscerate.ByID(31016)

// Below the book's level both resolve to the same learned rank, which registers once.
func (rogue *Rogue) registerEviscerate() {
	trainer := spelldata.LearnedRank(&rogue.Unit, spellData.Eviscerate, eviscerateTrainerRank)
	book := spelldata.LearnedRank(&rogue.Unit, spellData.Eviscerate, eviscerateBookRank)
	if trainer != spelldata.Nil {
		rogue.registerEviscerateRank(trainer)
	}
	if book != spelldata.Nil && book.ID != trainer.ID {
		rogue.registerEviscerateRank(book)
	}
}

func (rogue *Rogue) registerEviscerateRank(rank *spelldata.Spell) {
	eviscerateBaseDamage := rank.DamageEffect()
	eviscerateDamagePerPoint := float64(rank.DamageEffect().PointsPerResource)
	eviscerateApScalingPerPoint := 0.04 // Estimated from in-game tests.
	config := spelldata.SpellConfig(&rogue.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.Flags |= SpellFlagFinisher | SpellFlagColdBlooded

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		comboPoints := float64(rogue.ComboPoints())
		// The roll spreads the base points alone; the per-point damage and attack power add on top.
		flatBaseDamage := eviscerateBaseDamage.RollAt(sim, rogue.Level) + eviscerateDamagePerPoint*comboPoints

		baseDamage := flatBaseDamage + eviscerateApScalingPerPoint*comboPoints*spell.MeleeAttackPower(target)

		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if result.Landed() {
			rogue.ApplyFinisher(sim, spell)
		} else {
			spell.IssueRefund(sim)
		}

		spell.DealDamage(sim, result)
	}

	config.MetricSplits = 6
	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		spell.SetMetricsSplit(rogue.ComboPoints())
	}

	rogue.RegisterSpell(config)
}
