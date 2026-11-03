package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// This exists purely to model the DPS loss of having to occasionally kick
func (rogue *Rogue) registerKickSpell() {
	kickRank := spelldata.LearnedRank(&rogue.Unit, spellData.Kick, spellData.Kick.Highest())
	if kickRank == spelldata.Nil {
		return
	}
	kickBaseDamage := kickRank.DamageEffect().AverageAt(rogue.Level)
	config := spelldata.SpellConfig(&rogue.Unit, kickRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := kickBaseDamage
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	rogue.RegisterSpell(config)
}
