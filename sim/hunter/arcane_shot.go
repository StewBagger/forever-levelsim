package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: In-game testing required. The row states no attack power share; the Forever notes say the
// shot scales with 10% of ranged attack power.
const arcaneShotAPShare = 0.1

func (hunter *Hunter) registerArcaneShot() {
	rank := spelldata.LearnedRank(&hunter.Unit, spellData.ArcaneShot, spellData.ArcaneShot.Highest())
	if rank == spelldata.Nil {
		return
	}
	baseDamage := rank.DamageEffect().AverageAt(hunter.Level)

	config := spelldata.SpellConfig(&hunter.Unit, rank, spelldata.Melee(core.ProcMaskRangedSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		damage := baseDamage + arcaneShotAPShare*spell.RangedAttackPower(target)
		result := spell.CalcDamage(sim, target, damage, spell.OutcomeRangedHitAndCrit)

		spell.DealDamageAfterTravel(sim, result)
	}

	hunter.RegisterSpell(config)
}
