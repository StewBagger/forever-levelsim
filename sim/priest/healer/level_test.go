package healer

import (
	"fmt"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Below 60 the healer registers only the heals and shields taught at its level.
func TestHealerPriestBelowMaxLevelKnowsOnlyTaughtRanks(t *testing.T) {
	for _, level := range []int32{10, 20, 30, 40} {
		t.Run(fmt.Sprintf("level %d", level), func(t *testing.T) {
			player := core.WithSpec(
				&proto.Player{
					Name:      "Leveller",
					Class:     proto.Class_ClassPriest,
					Race:      proto.Race_RaceDwarf,
					Level:     level,
					Equipment: core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", "preraid").GearSet,
					Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				},
				PlayerOptions,
			)
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
			healer := sim.Raid.Parties[0].Players[0].(*HealerPriest)
			for _, spell := range healer.Spellbook {
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
