package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warrior *Warrior) registerPummel() {
	pummelRank := warrior.learnedRank(spellData.Pummel, spellData.Pummel.ByID(6554))
	if pummelRank == spelldata.Nil {
		return
	}
	pummelBaseDamage := pummelRank.DamageEffect().AverageAt(warrior.Level)

	config := spelldata.SpellConfig(&warrior.Unit, pummelRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, pummelBaseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
