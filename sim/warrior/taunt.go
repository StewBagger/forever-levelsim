package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warrior *Warrior) registerTaunt() {
	tauntRank := warrior.learnedRank(spellData.Taunt, spellData.Taunt.Highest())
	if tauntRank == spelldata.Nil {
		return
	}
	config := spelldata.SpellConfig(&warrior.Unit, tauntRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	config.Cast.DefaultCast.NonEmpty = true

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	}

	warrior.Taunt = warrior.RegisterSpell(config)
}
