package warlock

import (
	"fmt"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	googleProto "google.golang.org/protobuf/proto"
)

// Below level 60 each build runs its rotation with no talents, and the warlock and its demons know only the
// ranks taught at or below their level. A demon the warlock cannot yet summon has no summon and is not out.
func TestBelowMaxLevelKnowsOnlyTaughtRanks(t *testing.T) {
	builds := []string{"demo_fire_brand", "demo_shadow_brand", "destro", "sm_ruin", "aff_wrack"}
	for _, buildFile := range builds {
		build := core.GetTestBuildFromJSON(proto.Class_ClassWarlock, "../../ui/specs/warlock/dps/builds", buildFile, core.ItemFilter{}, nil, nil)
		spec := googleProto.Clone(build.SpecOptions.SpecOptions.(*proto.Player_Warlock).Warlock).(*proto.Warlock)

		for _, level := range []int32{10, 20, 30, 40} {
			t.Run(fmt.Sprintf("%s/%d", buildFile, level), func(t *testing.T) {
				player := &proto.Player{
					Name:          "Warlock",
					Level:         level,
					Race:          proto.Race_RaceOrc,
					Class:         proto.Class_ClassWarlock,
					Equipment:     &proto.EquipmentSpec{},
					TalentsString: "",
					Spec:          &proto.Player_Warlock{Warlock: googleProto.Clone(spec).(*proto.Warlock)},
					Rotation:      build.Rotation.Rotation,
				}
				raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
				sim := newWarlockRaidSim(raid, core.MakeSingleTargetEncounter(0))
				warlock := sim.Raid.Parties[0].Players[0].(*Warlock)
				sim.PrePull()
				stepFor(sim, 60*time.Second)

				assertTaughtAtLevel(t, &warlock.Unit, level)
				if !castAny(&warlock.Unit) {
					t.Errorf("the rotation cast no spell")
				}
				for _, pet := range warlock.demons() {
					assertTaughtAtLevel(t, &pet.Unit, level)
				}

				for summon, ladder := range summonLadders {
					if spelldata.TaughtAt(ladder.Rank(1)) <= level {
						continue
					}
					if pet := warlock.demon(summon); pet != nil {
						t.Errorf("the %s is registered, though its summon is taught at %d", pet.Name, spelldata.TaughtAt(ladder.Rank(1)))
					}
					if warlock.Options.Summon == summon || warlock.Options.SacrificedDemon == summon {
						t.Errorf("%s is out or sacrificed, though its summon is taught at %d", summon, spelldata.TaughtAt(ladder.Rank(1)))
					}
				}
			})
		}
	}
}

// Whether the unit cast any spell that has a row in the store: the rotation names the level 60 ranks, and
// reaches the learned ones through them.
func castAny(unit *core.Unit) bool {
	for _, spell := range unit.Spellbook {
		if spelldata.Find(spell.SpellID) == spelldata.Nil {
			continue
		}
		for _, metrics := range spell.SpellMetrics {
			if metrics.Casts > 0 {
				return true
			}
		}
	}
	return false
}

// Every spell in the unit's spellbook with a row in the store is taught at or below the level.
func assertTaughtAtLevel(t *testing.T, unit *core.Unit, level int32) {
	t.Helper()
	for _, spell := range unit.Spellbook {
		row := spelldata.Find(spell.SpellID)
		if spell.SpellID == 0 || row == spelldata.Nil {
			continue
		}
		if taught := spelldata.TaughtAt(row); taught > level {
			t.Errorf("%s knows %s (%d), taught at %d", unit.Label, row.Name, row.ID, taught)
		}
	}
}
