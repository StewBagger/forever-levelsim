package priest

import (
	"fmt"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Below 60 the priest registers only the ranks taught at its level, and the default rotation, which
// names the top ranks, runs on the ones it knows. Every race is built for its racial spells.
func TestShadowPriestBelowMaxLevelKnowsOnlyTaughtRanks(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceTroll, proto.Race_RaceUndead, proto.Race_RaceDwarf, proto.Race_RaceNightElf} {
		for _, level := range []int32{10, 20, 30, 40} {
			t.Run(fmt.Sprintf("%s level %d", race, level), func(t *testing.T) {
				player := &proto.Player{
					Name:      "Leveller",
					Race:      race,
					Class:     proto.Class_ClassPriest,
					Level:     level,
					Equipment: shadowTestingBuild.GearSet.GearSet,
					Spec: &proto.Player_DpsPriest{
						DpsPriest: &proto.DpsPriest{
							Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{
								Armor: proto.PriestOptions_InnerFire,
							}},
						},
					},
					Rotation: shadowTestingBuild.Rotation.Rotation,
				}
				request := &proto.RaidSimRequest{
					Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
					Encounter:  core.MakeSingleTargetEncounter(0),
					SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
				}

				result := core.RunRaidSim(request)
				if result.Error != nil {
					t.Fatalf("the sim failed: %s", result.Error.Message)
				}

				sim := core.NewSim(request, simsignals.CreateSignals())
				sim.Reset()
				priest := sim.Raid.Parties[0].Players[0].(*Priest)
				for _, spell := range priest.Spellbook {
					row := spelldata.Find(spell.SpellID)
					if row == spelldata.Nil {
						continue
					}
					if taught := spelldata.TaughtAt(row); taught > level {
						t.Errorf("%s %d is taught at %d and registered at level %d", row.Name, row.ID, taught, level)
					}
				}
			})
		}
	}
}
