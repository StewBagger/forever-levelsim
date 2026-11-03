package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 9 is the last rank the trainer teaches; rank 10 comes from the Grimoire of Shadow Bolt X book. Below
// the book's level both name the same learned rank, which registers once.
func (warlock *Warlock) registerShadowBolt() {
	trained := warlock.learnedRank(spellData.ShadowBolt, 11661)
	if trained == spelldata.Nil {
		return
	}
	warlock.ShadowBolt = warlock.registerShadowBoltRank(trained)
	if book := warlock.learnedRank(spellData.ShadowBolt, 25307); book != trained {
		warlock.registerShadowBoltRank(book)
	}
}

// Nightfall's Shadow Trance makes it instant through the cast time modifier parsed onto that aura.
func (warlock *Warlock) registerShadowBoltRank(rankedSpell *spelldata.Spell) *core.Spell {
	shadowBoltHit := rankedSpell.DamageEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellShadowBolt

	// Roll the damage and to hit, then land it after the bolt's travel time
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, shadowBoltHit.AverageAt(spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, shadowBoltHit.AverageAt(spell.Unit.Level), spell.OutcomeExpectedMagicCrit)
	}

	return warlock.RegisterSpell(config)
}
