package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 1 is taught at level 1, so every rogue registers the aura the other abilities read.
func (rogue *Rogue) registerStealthAura() {
	stealthRank := spelldata.LearnedRank(&rogue.Unit, spellData.Stealth, spellData.Stealth.Highest())
	config := spelldata.SpellConfig(&rogue.Unit, stealthRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagPrepullOnly))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.StealthAura.Activate(sim)
	}

	rogue.StealthAura = rogue.RegisterAura(spelldata.AuraConfig(stealthRank)).
		ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
			rogue.ShapeshiftForm = stealthRank.ShapeshiftForm()
		}).
		ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
			rogue.ShapeshiftForm = 0
		})

	spelldata.ParseEffects(&rogue.Character, rogue.StealthAura, stealthRank)

	rogue.RegisterSpell(config)
}
