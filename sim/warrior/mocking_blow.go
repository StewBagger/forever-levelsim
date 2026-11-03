package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warrior *Warrior) registerMockingBlow() {
	mockingBlowRank := warrior.learnedRank(spellData.MockingBlow, spellData.MockingBlow.Highest())
	if mockingBlowRank == spelldata.Nil {
		return
	}
	mockingBlowBaseDamage := mockingBlowRank.DamageEffect().AverageAt(warrior.Level)

	// TODO: Test the threat multiplier in-game
	config := spelldata.SpellConfig(&warrior.Unit, mockingBlowRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, mockingBlowBaseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.MockingBlow = warrior.RegisterSpell(config)
}
