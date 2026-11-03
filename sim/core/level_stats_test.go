package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

// The per-level tables at level 60 are the max-level values the sim had before it took a level.
func TestLevelTablesMatchMaxLevel(t *testing.T) {
	for class, byLevel := range classBaseStatsByLevel {
		at60 := byLevel[CharacterLevel]
		want := ClassBaseStats[class]
		// One level below the cap the curve is the generated table, scaled to land on ClassBaseStats at 60.
		at59 := BaseStatsAtLevel(0, class, CharacterLevel-1)
		for _, s := range []stats.Stat{stats.Health, stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit} {
			if at59[s] > want[s] {
				t.Errorf("%s level 59 %s = %v, above level 60's %v", class, s.StatName(), at59[s], want[s])
			}
		}
		if at60[stats.Mana] != ExtraClassBaseStats[class][stats.Mana] {
			t.Errorf("%s level 60 mana = %v, ExtraClassBaseStats has %v", class, at60[stats.Mana], ExtraClassBaseStats[class][stats.Mana])
		}
		if got := baseAttackPowerAtLevel(class, CharacterLevel); got != want[stats.AttackPower] {
			t.Errorf("%s level 60 attack power = %v, ClassBaseStats has %v", class, got, want[stats.AttackPower])
		}
		if got, want := CritPerAgiAtLevel(class, CharacterLevel), CritPerAgiMaxLevel[class]; math.Abs(got-want) > 1e-6 {
			t.Errorf("%s crit per Agility at 60 = %v, want %v", class, got, want)
		}
		if got, want := CritPerIntAtLevel(class, CharacterLevel), CritPerIntMaxLevel[class]; math.Abs(got-want) > 1e-6 {
			t.Errorf("%s crit per Intellect at 60 = %v, want %v", class, got, want)
		}
	}
	h, a, d := CreatureStatsAtLevel(CharacterLevel)
	if h != CreatureHealthMaxLevel || a != CreatureArmorMaxLevel || math.Abs(d-CreatureAutoAttackDpsMaxLevel) > 1e-9 {
		t.Errorf("creature stats at 60 = %v %v %v, want the max-level constants", h, a, d)
	}
}
