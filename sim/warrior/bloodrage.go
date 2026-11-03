package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (warrior *Warrior) registerBloodrage() {
	bloodrageRank := warrior.learnedRank(spellData.Bloodrage, spellData.Bloodrage.Highest())
	if bloodrageRank == spelldata.Nil {
		return
	}
	bloodrageOverTime := bloodrageRank.EffectN(2).Trigger()
	bloodrageOverTimeTick := bloodrageOverTime.PeriodicEffect()

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: bloodrageRank.ID})
	healthCost := warrior.GetBaseStats()[stats.Health] * float64(bloodrageRank.Power(dbcenums.POWER_HEALTH).CostPct) / 100
	improvedBloodrage := spellData.ImprovedBloodrage.MultiplierAt(warrior.Talents.ImprovedBloodrage)
	instantRage := bloodrageRank.EnergizeEffect().Tenths() * improvedBloodrage
	ragePerTick := bloodrageOverTimeTick.Tenths() * improvedBloodrage

	config := spelldata.SpellConfig(&warrior.Unit, bloodrageRank)

	config.Hot = spelldata.DotConfig(bloodrageOverTime, bloodrageOverTimeTick)
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
		warrior.AddRage(sim, ragePerTick, rageMetrics)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		warrior.AddRage(sim, instantRage, rageMetrics)
		warrior.RemoveHealth(sim, healthCost)
		spell.SelfHot().Apply(sim)
	}

	spell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
			return warrior.CurrentRage() < 70
		},
	})
}
