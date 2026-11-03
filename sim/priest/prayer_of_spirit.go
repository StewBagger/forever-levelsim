package priest

import (
	"github.com/wowsims/forever/sim/core/buffs"
)

func (priest *Priest) registerPrayerOfSpirit() {
	priest.registerPrayer(priest.learnedTopRank(spellData.PrayerOfSpirit), buffs.NewSpellDataPrayerOfSpiritAura)
}
