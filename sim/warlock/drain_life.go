package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerDrainLife() {
	rankedSpell := warlock.learnedRank(spellData.DrainLife, 11700)
	if rankedSpell == spelldata.Nil {
		return
	}
	drainLifeTick := rankedSpell.PeriodicEffect()

	healthMetrics := warlock.NewHealthMetrics(core.ActionID{SpellID: rankedSpell.ID})

	// Heal for the damage each tick dealt
	healOnTick := func(sim *core.Simulation, result *core.SpellResult) {
		heal := result.Damage * warlock.PseudoStats.SelfHealingMultiplier * warlock.PseudoStats.HealingTakenMultiplier
		warlock.GainHealth(sim, heal, healthMetrics)
	}

	config := warlock.soulSiphonChannelConfig(rankedSpell, drainLifeTick, healOnTick)
	config.ClassSpellMask = WarlockSpellDrainLife

	warlock.DrainLife = warlock.RegisterSpell(config)
}
