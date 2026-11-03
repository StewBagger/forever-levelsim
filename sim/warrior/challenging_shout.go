package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warrior *Warrior) registerChallengingShout() {
	challengingShoutRank := warrior.learnedRank(spellData.ChallengingShout, spellData.ChallengingShout.Highest())
	if challengingShoutRank == spelldata.Nil {
		return
	}
	config := spelldata.SpellConfig(&warrior.Unit, challengingShoutRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
			spell.CalcAndDealOutcome(sim, aoeTarget, spell.OutcomeMagicHit)
		}
	}

	warrior.ChallengingShout = warrior.RegisterSpell(config)
}
