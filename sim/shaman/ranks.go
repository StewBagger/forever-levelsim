package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// In game a rank deals less damage once the caster is far enough past the level it is taught at; a
// rank taught within this many levels of the caster's own is clear of that, so at level 60 the ranks
// taught at 44 or later.
const rankFalloffLevels = 16 // TODO: Ingame test needed: the cutoff is believed to be 42 or 43 at level 60.

// The ranks of a damage spell the shaman registers: those it has learned at its level, taught within
// rankFalloffLevels of it. The highest learned rank is always one of them, so a shaman whose newest
// rank is older than the cutoff still has the spell. A rank above the shaman's level is aliased to the
// highest learned one, so a rotation naming it casts that.
func castableRanks(unit *core.Unit, ladder spelldata.Ladder) spelldata.Ladder {
	minRankLevel := unit.Level - rankFalloffLevels
	highest := spelldata.LearnedRank(unit, ladder, ladder.Highest())

	var ids []int32
	ladder.Each(func(_ int32, rank *spelldata.Spell) {
		if spelldata.TaughtAt(rank) > unit.Level {
			spelldata.LearnedRank(unit, ladder, rank)
			return
		}
		if spelldata.TaughtAt(rank) >= minRankLevel || rank == highest {
			ids = append(ids, rank.ID)
		}
	})
	return spelldata.Ranked(ids...)
}
