package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerInsectSwarmSpell() {
	insectSwarmRank := druid.learned(spellData.InsectSwarm)
	if insectSwarmRank == spelldata.Nil {
		return
	}
	auras := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataInsectSwarmAura(target, true, 0)
	})

	// Binary: the row's second effect lowers the target's chance to hit, and a spell with a
	// non-damage effect is resisted whole or not at all. Its ticks still take partial resists, against
	// the full resistance.
	config := spelldata.SpellConfig(&druid.Unit, insectSwarmRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary|core.SpellFlagFullResistanceTicks))
	tick := insectSwarmRank.PeriodicEffect()
	config.Dot = spelldata.DotConfig(insectSwarmRank, tick)
	// Each tick is priced at the druid's level.
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.RollAt(sim, druid.level()), insectSwarmRank.TickOutcome(dot))
	}
	// A duration mod such as Nature's Splendor's lengthens the hit debuff through it, as it does the dot.
	config.RelatedAuraArrays = auras.ToMap()

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			auras.Get(target).Activate(sim)
		}
		spell.DealOutcome(sim, result)
	}

	druid.InsectSwarm = druid.RegisterSpell(config)
}
