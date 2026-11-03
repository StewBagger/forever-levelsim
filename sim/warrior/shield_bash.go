package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warrior *Warrior) registerShieldBash() {
	shieldBashRank := warrior.learnedRank(spellData.ShieldBash, spellData.ShieldBash.Highest())
	if shieldBashRank == spelldata.Nil {
		return
	}
	// TODO: Manual review needed -- the game files state no threat coefficient; 1 until measured in game.
	config := spelldata.SpellConfig(&warrior.Unit, shieldBashRank, spelldata.Melee(core.ProcMaskMeleeOHSpecial))

	// TODO: In-game test required
	config.FlatThreatBonus = 0

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := shieldBashRank.DamageEffect().RollAt(sim, spell.Unit.Level)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
