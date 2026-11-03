package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warrior *Warrior) registerSunderArmor() {
	sunderArmorRank := warrior.learnedRank(spellData.SunderArmor, spellData.SunderArmor.Highest())
	if sunderArmorRank == spelldata.Nil {
		return
	}
	warrior.SunderArmorAuras = warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataSunderArmorAura(target, true, 0)
	})

	config := spelldata.SpellConfig(&warrior.Unit, sunderArmorRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)

		if result.Landed() {
			aura := warrior.SunderArmorAuras.Get(target)
			aura.Activate(sim)
			aura.AddStack(sim)
		} else {
			spell.IssueRefund(sim)
		}

		spell.DealOutcome(sim, result)
	}

	config.RelatedAuraArrays = warrior.SunderArmorAuras.ToMap()

	warrior.RegisterSpell(config)
}
