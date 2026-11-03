package warlock

import "github.com/wowsims/forever/sim/core/spelldata"

func (warlock *Warlock) registerDrainSoul() {
	rankedSpell := warlock.learnedRank(spellData.DrainSoul, 11675)
	if rankedSpell == spelldata.Nil {
		return
	}
	drainSoulTick := rankedSpell.PeriodicEffect()

	config := warlock.soulSiphonChannelConfig(rankedSpell, drainSoulTick, nil)
	config.ClassSpellMask = WarlockSpellDrainSoul

	warlock.DrainSoul = warlock.RegisterSpell(config)
}
