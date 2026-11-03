package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Tags every warlock's Immolate dot, so Incinerate can see any Immolate on the target.
const immolateTag = "Immolate"

// Rank 7 is the last rank the trainer teaches; rank 8 comes from the Grimoire of Immolate VIII book. Below
// the book's level both name the same learned rank, which registers once.
func (warlock *Warlock) registerImmolate() {
	trained := warlock.learnedRank(spellData.Immolate, 11668)
	if trained == spelldata.Nil {
		return
	}
	warlock.Immolate = []*core.Spell{warlock.registerImmolateRank(trained)}
	if book := warlock.learnedRank(spellData.Immolate, 25309); book != trained {
		warlock.Immolate = append(warlock.Immolate, warlock.registerImmolateRank(book))
	}
}

// The ranks share one rank group, so at most one of them burns on a target.
func (warlock *Warlock) activeImmolate(target *core.Unit) *core.Dot {
	for _, spell := range warlock.Immolate {
		if dot := spell.Dot(target); dot.IsActive() {
			return dot
		}
	}
	return nil
}

// A hit, then a dot. Effect 3 is a server script (E_SCRIPT_EFFECT) the game files do not describe.
func (warlock *Warlock) registerImmolateRank(rankedSpell *spelldata.Spell) *core.Spell {
	immolateHit := rankedSpell.DamageEffect()
	immolateTick := rankedSpell.PeriodicEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellImmolate

	// Tick the row's average, rolling only the crit: the hit was rolled when the dot landed
	config.Dot = spelldata.DotConfig(rankedSpell, immolateTick)
	config.Dot.Aura.Tag = immolateTag
	config.Dot.OnTick = periodicDamageTick(immolateTick, rankedSpell.TickOutcome)

	// Roll the hit, and apply the dot if it landed
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, immolateHit.AverageAt(spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	// Calculate expected damage from the row's average hit and tick
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, immolateHit.AverageAt(spell.Unit.Level), spell.OutcomeExpectedMagicCrit)
	}
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, immolateTick.AverageAt(spell.Unit.Level), spell.OutcomeExpectedMagicAlwaysHit)
	}

	return warlock.RegisterSpell(config)
}
