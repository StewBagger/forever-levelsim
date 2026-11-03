package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// One spell carries the hit and the dot, so a SPELLMOD_DAMAGE (direct hits only) and a SPELLMOD_DOT
// (ticks only) that both name Moonfire, as Improved Moonfire does, each reach their own half.
func (druid *Druid) registerMoonfireSpell() {
	rank := druid.learned(spellData.Moonfire)
	if rank == spelldata.Nil {
		return
	}
	damage := rank.DamageEffect()
	tick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.Dot = spelldata.DotConfig(rank, tick)
	// Each tick is priced at the druid's level.
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.RollAt(sim, druid.level()), rank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.AverageAt(druid.level()), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealDamage(sim, result)
	}

	druid.Moonfire = druid.RegisterSpell(config)
}
