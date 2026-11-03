package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerWrathSpell() {
	rank := druid.learned(spellData.Wrath)
	if rank == spelldata.Nil {
		return
	}
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.AverageAt(druid.level()), spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}

	druid.Wrath = druid.RegisterSpell(config)
}
