package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (rogue *Rogue) registerGarrote() {
	garroteRank := spelldata.LearnedRank(&rogue.Unit, spellData.Garrote, spellData.Garrote.Highest())
	if garroteRank == spelldata.Nil {
		return
	}
	tick := garroteRank.PeriodicEffect()
	apScaling := 0.03 // Estimated from in-game tests.
	comboPointGen := garroteRank.EnergizeEffect().AverageAt(rogue.Level)

	config := spelldata.SpellConfig(&rogue.Unit, garroteRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder
	config.ClassSpellMask = RogueSpellGarrote
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return (!rogue.PseudoStats.InFrontOfTarget || rogue.Talents.DirtyDeeds > 0) && rogue.IsStealthed()
	}

	config.Dot = spelldata.DotConfig(garroteRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.AverageAt(rogue.Level)+dot.Spell.MeleeAttackPower(target)*apScaling, garroteRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			rogue.AddComboPoints(sim, int32(comboPointGen), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Garrote = rogue.RegisterSpell(config)
}
