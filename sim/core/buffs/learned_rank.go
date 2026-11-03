package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A character below max level applies the rank it has learned, not the top rank the raid config
// reads. These build the aura from that rank's row; at the top rank they are the generated ones.

func NewSpellDataExposeArmorAuraAtRank(unit *core.Unit, isPlayer bool, talentPoints int32, rank *spelldata.Spell) *core.Aura {
	return newDebuff(unit, exposeArmorMeta.atState(rank, exposeArmorMeta.Label), isPlayer, talentPoints)
}
