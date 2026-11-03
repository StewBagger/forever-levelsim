package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 4 is the last rank the trainer teaches; rank 5 comes from the Handbook of Feint V book.
var feintTrainerRank = spellData.Feint.ByID(11303)
var feintBookRank = spellData.Feint.ByID(25302)

// This exists purely to model the DPS loss of having to occasionally feint
// Below the book's level both resolve to the same learned rank, which registers once.
func (rogue *Rogue) registerFeintSpell() {
	trainer := spelldata.LearnedRank(&rogue.Unit, spellData.Feint, feintTrainerRank)
	book := spelldata.LearnedRank(&rogue.Unit, spellData.Feint, feintBookRank)
	if trainer != spelldata.Nil {
		rogue.registerFeintRank(trainer)
	}
	if book != spelldata.Nil && book.ID != trainer.ID {
		rogue.registerFeintRank(book)
	}
}

func (rogue *Rogue) registerFeintRank(rank *spelldata.Spell) {
	config := spelldata.SpellConfig(&rogue.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	// The row lacks No Active Defense, so the target can dodge it and, from the front, parry it.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcAndDealDamage(sim, target, 0, spell.OutcomeMeleeSpecialHit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	rogue.RegisterSpell(config)
}
