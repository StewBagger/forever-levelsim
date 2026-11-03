package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: The Attack Power share per tick (0.02) is verified in game on Rank 2 only; Ranks 1 and 3 to 7
// carry the same share unmeasured. Verify it on the higher ranks once logs at their level exist.
func (warrior *Warrior) registerRend() {
	rendRank := warrior.learnedRank(spellData.Rend, spellData.Rend.Highest())
	if rendRank == spelldata.Nil {
		return
	}
	tick := rendRank.PeriodicEffect()

	config := spelldata.SpellConfig(&warrior.Unit, rendRank,
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Flags(core.SpellFlagNoOnCastComplete))

	config.Dot = spelldata.DotConfig(rendRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := tick.AverageAt(dot.Spell.Unit.Level) + tick.APCoeff()*dot.Spell.MeleeAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, rendRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		} else {
			spell.IssueRefund(sim)
		}
	}

	warrior.Rend = warrior.RegisterSpell(config)
}
