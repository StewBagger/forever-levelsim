package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rank a rotation names; the hunter registers the one it has learned in its place.
var huntersMarkRank = spellData.HuntersMark.Highest()

func (hunter *Hunter) registerHuntersMark() {
	huntersMarkRank := spelldata.LearnedRank(&hunter.Unit, spellData.HuntersMark, huntersMarkRank)
	if huntersMarkRank == spelldata.Nil {
		return
	}

	auras := hunter.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataHuntersMarkAura(target, true, 0)
	})

	config := spelldata.SpellConfig(&hunter.Unit, huntersMarkRank, spelldata.Flags(core.SpellFlagAPL))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		auras.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = auras.ToMap()

	hunter.RegisterSpell(config)
}
