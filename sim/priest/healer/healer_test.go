package healer

import (
	"fmt"
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/levelsim"
)

func init() {
	RegisterHealerPriest()
	common.RegisterAllEffects()
}

// Stats-only suite: this spec is a gear planner, it has no healing rotation.
// Pins the final character stats for each gear preset so the passives stay covered. The empty APL
// rotation and the fake prepull (no SkipRotation) make it exercise a full environment reset, the
// path the UI's stats request takes.
func TestHealerPriest(t *testing.T) {
	var generators []core.TestGenerator
	for _, gearSet := range []string{"preraid"} {
		player := core.WithSpec(
			&proto.Player{
				Class:         proto.Class_ClassPriest,
				Race:          proto.Race_RaceDwarf,
				Equipment:     core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", gearSet).GearSet,
				Consumables:   FullConsumes,
				Buffs:         core.FullIndividualBuffs,
				TalentsString: StandardTalents,
				Profession1:   proto.Profession_Tailoring,
				Profession2:   proto.Profession_Enchanting,
				Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			},
			PlayerOptions,
		)
		generators = append(generators, &core.SingleCharacterStatsTestGenerator{
			Name: gearSet,
			Request: &proto.ComputeStatsRequest{
				Raid: core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
			},
		})
	}
	core.RunTestSuite(t, t.Name(), generators)
}

// The healer in a 5-man group at levels 30 and 60: the tank and three damage dealers as target dummies,
// the level's damage profile from levelsim, and the group heal priority in ui/specs/priest/healer/apls/group.apl.json.
// Records the HPS and the HPS EP on effective healing at the healer's healing capacity.
func TestHealerPriestGroup(t *testing.T) {
	var generators []core.TestGenerator
	for _, level := range groupLevels {
		name := fmt.Sprintf("Level%d", level)
		generators = append(generators,
			&core.SingleDpsTestGenerator{Name: name, Request: groupRequest(level, groupSimOptions)},
			&core.SingleStatWeightsTestGenerator{Name: name + "-StatWeights", Request: groupStatWeightsRequest(level)},
		)
	}
	core.RunTestSuite(t, t.Name(), generators)
}

// The group fight reports what a healer is judged by: effective healing below all healing, the
// overhealing it records per heal, the healer's mana running out, and the group members that die.
func TestHealerPriestGroupReportsHealingOutcomes(t *testing.T) {
	for _, level := range groupLevels {
		t.Run(fmt.Sprintf("level %d", level), func(t *testing.T) {
			result := core.RunRaidSim(groupRequest(level, groupSimOptions))
			if result.Error != nil {
				t.Fatalf("the sim failed: %s", result.Error.Message)
			}
			players := result.RaidMetrics.Parties[0].Players
			healer := players[0]
			if !(healer.Ehps.Avg > 0) {
				t.Fatalf("effective HPS is %.1f", healer.Ehps.Avg)
			}
			if !(healer.Ehps.Avg < healer.Hps.Avg) {
				t.Errorf("effective HPS %.1f is not below HPS %.1f", healer.Ehps.Avg, healer.Hps.Avg)
			}
			overhealing := 0.0
			for _, action := range healer.Actions {
				for _, target := range action.Targets {
					overhealing += target.Overhealing
				}
			}
			if !(overhealing > 0) {
				t.Errorf("no heal recorded overhealing")
			}
			if want := 1 - healer.Ehps.Avg/healer.Hps.Avg; healer.OverhealingPercent < want-1e-9 || healer.OverhealingPercent > want+1e-9 {
				t.Errorf("overhealing percent %.4f, want 1 - ehps/hps = %.4f", healer.OverhealingPercent, want)
			}
			if !(healer.Tto.Avg > 0) {
				t.Errorf("no time to OOM reported")
			}
			if len(players) != 1+levelsim.HealingGroupDummies {
				t.Fatalf("%d group members, want the healer and %d dummies", len(players), levelsim.HealingGroupDummies)
			}
			for _, member := range players[1:] {
				if !(member.Dtps.Avg > 0) {
					t.Errorf("%s took no damage", member.Name)
				}
			}
			t.Logf("HPS %.1f, effective HPS %.1f, overhealing %.1f%%, time to OOM %.1f s, OOM %.1f s a fight",
				healer.Hps.Avg, healer.Ehps.Avg, 100*healer.OverhealingPercent, healer.Tto.Avg, healer.SecondsOomAvg)
			for _, member := range players {
				t.Logf("  %s: DTPS %.1f, chance of death %.1f%%", member.Name, member.Dtps.Avg, 100*member.ChanceOfDeath)
			}
		})
	}
}

var groupLevels = []int32{30, 60}

var groupSimOptions = &proto.SimOptions{Iterations: 200, IsTest: true, RandomSeed: 101}

// The EP the group's gear is picked by, the Holy preset's from ui/specs/priest/healer/presets.ts.
var groupGearEP = map[proto.Stat]float64{
	proto.Stat_StatHealingPower: 1,
	proto.Stat_StatCritRating:   0.7,
	proto.Stat_StatHasteRating:  0.15,
	proto.Stat_StatIntellect:    0.2,
	proto.Stat_StatSpirit:       0.15,
	proto.Stat_StatMP5:          0.4,
}

// The healer at a level: the Holy talents spent up to its points and the level's best gear by
// groupGearEP, with no consumables or buffs.
func groupHealer(level int32) *proto.Player {
	player := core.WithSpec(&proto.Player{
		Name:          "Healer",
		Class:         proto.Class_ClassPriest,
		Race:          proto.Race_RaceDwarf,
		Level:         level,
		TalentsString: levelsim.TrimTalents(StandardTalents, core.TalentPointsAtLevel(level)),
		Rotation:      core.GetAplRotation("../../../ui/specs/priest/healer/apls", "group").Rotation,
		Profession1:   proto.Profession_Tailoring,
		Profession2:   proto.Profession_Enchanting,
	}, PlayerOptions)
	ep := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
	for stat, value := range groupGearEP {
		ep.Stats[stat] = value
	}
	player.Equipment = levelsim.BisGear(&proto.BisGearRequest{Player: player, Level: level, Ep: ep}).Equipment
	return player
}

func groupEncounter(level int32) *proto.Encounter {
	return &proto.Encounter{
		Duration:          120,
		DurationVariation: 5,
		Targets:           []*proto.Target{levelsim.HealingGroupTarget(level)},
		DamageProfile:     levelsim.HealingGroupDamageProfile(level),
	}
}

func groupRequest(level int32, options *proto.SimOptions) *proto.RaidSimRequest {
	raid := core.SinglePlayerRaidProto(groupHealer(level), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = levelsim.HealingGroupTanks()
	core.AddTargetDummies(raid, levelsim.HealingGroupDummies, levelsim.HealingGroupDummySettings(level))
	return &proto.RaidSimRequest{Raid: raid, Encounter: groupEncounter(level), SimOptions: options}
}

func groupStatWeightsRequest(level int32) *proto.StatWeightsRequest {
	swr := &proto.StatWeightsRequest{
		Player:           groupHealer(level),
		PartyBuffs:       &proto.PartyBuffs{},
		RaidBuffs:        &proto.RaidBuffs{},
		Debuffs:          &proto.Debuffs{},
		Encounter:        groupEncounter(level),
		SimOptions:       core.StatWeightsDefaultSimTestOptions,
		StatsToWeigh:     []proto.Stat{proto.Stat_StatHealingPower, proto.Stat_StatIntellect, proto.Stat_StatSpirit, proto.Stat_StatMP5, proto.Stat_StatCritRating},
		EpReferenceStat:  proto.Stat_StatHealingPower,
		HpsReferenceStat: proto.Stat_StatHealingPower,
	}
	levelsim.WithHealingGroup(swr, level)
	return swr
}

// The Holy preset from ui/specs/priest/healer/presets.ts.
var StandardTalents = "00500500123-33505003030121531"

var FullConsumes = &proto.ConsumesSpec{
	FoodId:        250066, // Soaring Pamplona
	ResourcePotId: 13444,  // Major Mana Potion
}

var PlayerOptions = &proto.Player_HealerPriest{
	HealerPriest: &proto.HealerPriest{
		Options: &proto.HealerPriest_Options{
			ClassOptions: &proto.PriestOptions{},
		},
	},
}
