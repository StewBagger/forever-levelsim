package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (rogue *Rogue) registerSinisterStrikeSpell() {
	sinisterStrikeRank := spelldata.LearnedRank(&rogue.Unit, spellData.SinisterStrike, spellData.SinisterStrike.Highest())
	if sinisterStrikeRank == spelldata.Nil {
		return
	}
	sinisterStrikeBaseDamage := sinisterStrikeRank.DamageEffect().AverageAt(rogue.Level)
	config := spelldata.SpellConfig(&rogue.Unit, sinisterStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder | SpellFlagColdBlooded
	config.ClassSpellMask = RogueSpellSinisterStrike

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := sinisterStrikeBaseDamage +
			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(sinisterStrikeRank.ProcEnergizeEffect().RollAt(sim, rogue.Level)), spell.ComboPointMetrics())

		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.SinisterStrike = rogue.RegisterSpell(config)
}
