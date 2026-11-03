package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var victoriousRank = spellData.VictoryRushTriggered.Highest()

// This spell works but the Sim never kills a target
func (warrior *Warrior) registerVictoryRush() {
	victoryRushRank := warrior.learnedRank(spellData.VictoryRush, spellData.VictoryRush.Highest())
	if victoryRushRank == spelldata.Nil {
		return
	}
	victoryRushAPCoef := victoryRushRank.EffectN(3).Percent()
	victoryRushHealPercent := victoryRushRank.EffectN(2).Percent()

	healthMetrics := warrior.NewHealthMetrics(core.ActionID{SpellID: victoryRushRank.ID})

	victoriousAura := warrior.RegisterAura(spelldata.AuraConfig(victoriousRank))

	config := spelldata.SpellConfig(&warrior.Unit, victoryRushRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return victoriousAura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := victoryRushRank.DamageEffect().AverageAt(spell.Unit.Level) + victoryRushAPCoef*spell.MeleeAttackPower(target)
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		warrior.GainHealth(sim, warrior.MaxHealth()*victoryRushHealPercent, healthMetrics)
		victoriousAura.Deactivate(sim)
	}

	config.RelatedSelfBuff = victoriousAura

	warrior.VictoryRush = warrior.RegisterSpell(config)
}
