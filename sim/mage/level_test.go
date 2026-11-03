package mage

import (
	"fmt"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A mage below 60 sims each spec's rotation, and registers only the ranks taught at or below its level.
func TestBelowMaxLevelRegistersOnlyTaughtRanks(t *testing.T) {
	for _, suite := range []core.CharacterSuiteConfig{arcaneSuite(), ffbSuite(), fireSuite(), frostSuite()} {
		for _, level := range []int32{10, 20, 30, 40} {
			t.Run(fmt.Sprintf("%s/%d", suite.Rotation.Label, level), func(t *testing.T) {
				player := core.WithSpec(&proto.Player{
					Class:     suite.Class,
					Race:      suite.Race,
					Level:     level,
					Equipment: &proto.EquipmentSpec{},
					Rotation:  suite.Rotation.Rotation,
				}, suite.SpecOptions.SpecOptions)
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
				mage := sim.Raid.Parties[0].Players[0].(*Mage)
				for _, spell := range mage.Spellbook {
					row := spelldata.Find(spell.SpellID)
					if spell.SpellID == 0 || row == spelldata.Nil {
						continue
					}
					if taught := spelldata.TaughtAt(row); taught > level {
						t.Errorf("%s (%d) is registered at level %d, taught at %d", row.RankedName(), row.ID, level, taught)
					}
				}
			})
		}
	}
}
