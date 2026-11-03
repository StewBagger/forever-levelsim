package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The top rank, 7, and the cheaper rank 6, which an APL can cast in its place. Below rank 7's level
// both name the highest rank the druid knows, which is registered once.
func (druid *Druid) registerStarfireSpells() {
	registered := map[int32]bool{}
	for _, named := range []int32{6, 7} {
		rank := druid.learnedRank(spellData.Starfire, spellData.Starfire.Rank(named))
		if rank == spelldata.Nil || registered[rank.ID] {
			continue
		}
		registered[rank.ID] = true
		druid.registerStarfireSpell(rank)
	}
}

func (druid *Druid) registerStarfireSpell(rankConfig *spelldata.Spell) {
	damage := rankConfig.DamageEffect()

	config := spelldata.SpellConfig(&druid.Unit, rankConfig, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.AverageAt(druid.level()), spell.OutcomeMagicHitAndCrit)
	}

	druid.Starfire = append(druid.Starfire, druid.RegisterSpell(config))
}
