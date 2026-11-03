//go:build with_db

package shaman_test

import (
	"fmt"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/shaman/elemental"
)

func init() {
	elemental.RegisterElementalShaman()
}

// The spells core/buffs registers for a totem the shaman stands, at the totem's top rank whatever rank
// the shaman knows.
var topRankBuffSpells = map[int32]bool{
	16387: true, // Flametongue Totem's strike, from buffs.NewFlametongueTotemProcTrigger.
}

// A shaman below 60 sims each spec's Testing build with no talents, and registers only the ranks taught
// at or below its level. A rotation naming a rank above that level casts the learned one, so each
// build still deals damage.
func TestBelowMaxLevelRegistersOnlyTaughtRanks(t *testing.T) {
	for _, spec := range []string{"elemental", "enhancement"} {
		build := core.GetTestBuildFromJSON(proto.Class_ClassShaman, "../../ui/specs/shaman/"+spec+"/builds", "testing",
			core.ItemFilter{}, nil, nil)
		for _, level := range []int32{10, 20, 30, 40} {
			t.Run(fmt.Sprintf("%s/%d", spec, level), func(t *testing.T) {
				player := core.WithSpec(&proto.Player{
					Class:     proto.Class_ClassShaman,
					Race:      build.Race,
					Level:     level,
					Equipment: build.GearSet.GearSet,
					Rotation:  build.Rotation.Rotation,
				}, build.SpecOptions.SpecOptions)
				request := &proto.RaidSimRequest{
					Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
					Encounter:  &proto.Encounter{Duration: 30, Targets: []*proto.Target{{}}},
					SimOptions: &proto.SimOptions{Iterations: 2, IsTest: true, RandomSeed: 1},
				}

				result := core.RunRaidSim(request)
				if result.Error != nil {
					t.Fatalf("sim failed: %s", result.Error.Message)
				}
				if dps := result.RaidMetrics.Dps.Avg; dps <= 0 {
					t.Errorf("the rotation dealt no damage, %.1f dps", dps)
				}

				sim := core.NewSim(request, simsignals.CreateSignals())
				sim.Reset()
				sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
				for _, spell := range sham.Spellbook {
					row := spelldata.Find(spell.SpellID)
					if spell.SpellID == 0 || spell.ActionID.ItemID != 0 || row == spelldata.Nil {
						continue
					}
					if taught := spelldata.TaughtAt(row); taught > level && !topRankBuffSpells[row.ID] {
						t.Errorf("%s (%d) is registered at level %d, taught at %d", row.RankedName(), row.ID, level, taught)
					}
				}
			})
		}
	}
}
