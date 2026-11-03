package retribution

import (
	"fmt"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/paladin"
	"github.com/wowsims/forever/sim/paladin/holy"
	"github.com/wowsims/forever/sim/paladin/protection"
)

func init() {
	holy.RegisterHolyPaladin()
	protection.RegisterProtectionPaladin()
}

// A levelling Retribution build for each level: the points the level has, up to Seal of Command at 30
// and the full tree from 40.
var levellingTalents = map[int32]string{
	10: "",
	20: "--0523",
	30: "--05233001",
	40: "--05233011201301351",
	50: "--05233011201301351",
}

// Below 60 each spec sims its rotation, and the paladin knows only the ranks taught at its level.
func TestPaladinBelowMaxLevel(t *testing.T) {
	ret := func(talents, rotation string) func() *proto.Player {
		return func() *proto.Player {
			return &proto.Player{
				Race:          proto.Race_RaceHuman,
				Class:         proto.Class_ClassPaladin,
				Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
				TalentsString: talents,
				Spec:          DefaultSpecOptions,
				Rotation:      core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", rotation).Rotation,
			}
		}
	}

	for _, level := range []int32{10, 20, 30, 40, 50} {
		specs := map[string]func() *proto.Player{
			"Retribution":            ret("", "default"),
			"Retribution levelling":  ret(levellingTalents[level], "default"),
			"Retribution Holy Shock": ret("", "holy_shock"),
			"Holy": func() *proto.Player {
				return &proto.Player{
					Race:      proto.Race_RaceHuman,
					Class:     proto.Class_ClassPaladin,
					Equipment: core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", "default").GearSet,
					Spec: &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{
						Options: &proto.HolyPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
					}},
					Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				}
			},
			"Protection": func() *proto.Player {
				return &proto.Player{
					Race:      proto.Race_RaceHuman,
					Class:     proto.Class_ClassPaladin,
					Equipment: core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet,
					Spec: &proto.Player_ProtectionPaladin{ProtectionPaladin: &proto.ProtectionPaladin{
						Options: &proto.ProtectionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
					}},
					Rotation: core.GetAplRotation("../../../ui/specs/paladin/protection/apls", "default").Rotation,
				}
			},
		}

		for name, makePlayer := range specs {
			t.Run(fmt.Sprintf("%s/%d", name, level), func(t *testing.T) {
				player := makePlayer()
				player.Level = level
				request := &proto.RaidSimRequest{
					Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
					Encounter:  core.MakeSingleTargetEncounter(0),
					SimOptions: &proto.SimOptions{Iterations: 5, RandomSeed: 1, IsTest: true},
				}
				request.Encounter.Duration = 60
				request.Encounter.Targets[0].Level = 0

				result := core.RunRaidSim(request)
				if result.Error != nil {
					t.Fatalf("sim failed: %s", result.Error.Message)
				}
				if name != "Holy" && result.RaidMetrics.Dps.Avg <= 0 {
					t.Errorf("dealt no damage")
				}

				sim := core.NewSim(request, simsignals.CreateSignals())
				pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
				if pal.Level != level {
					t.Fatalf("the paladin is level %d, want %d", pal.Level, level)
				}
				for _, spell := range pal.Spellbook {
					row := spelldata.Find(spell.SpellID)
					if spell.SpellID == 0 || row == spelldata.Nil {
						continue
					}
					if taught := spelldata.TaughtAt(row); taught > level {
						t.Errorf("%s (%d) is registered at level %d and taught at %d", spell.ActionID, spell.SpellID, level, taught)
					}
				}
			})
		}
	}
}
