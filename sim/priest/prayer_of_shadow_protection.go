package priest

import (
	"github.com/wowsims/forever/sim/core/buffs"
)

func (priest *Priest) registerPrayerOfShadowProtection() {
	priest.registerPrayer(priest.learnedTopRank(spellData.PrayerOfShadowProtection), buffs.NewSpellDataPrayerOfShadowProtectionAura)
}
