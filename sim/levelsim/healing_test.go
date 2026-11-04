package levelsim

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// The Healer Priest's stat weights request at a level as Defaults weighs it: the site's setup from
// tools/levelsim/requests, the level's healing group, the talents spent up to the level, and the level's
// best gear by the Holy preset's EP.
func healerPriestCapacityRequest(t *testing.T, level int32) *proto.StatWeightsRequest {
	data, err := os.ReadFile("../../tools/levelsim/requests/priest_healer.json")
	if err != nil {
		t.Fatal(err)
	}
	req := &proto.LevelDefaultsRequest{}
	if err := protojson.Unmarshal(data, req); err != nil {
		t.Fatal(err)
	}
	swr := req.StatWeights
	WithHealingGroup(swr, level)
	player := swr.Player
	player.Level = level
	player.TalentsString = TrimTalents(player.TalentsString, core.TalentPointsAtLevel(level))
	for _, target := range swr.Encounter.Targets {
		target.Level = level
	}
	ep := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
	for stat, value := range map[proto.Stat]float64{
		proto.Stat_StatHealingPower: 1,
		proto.Stat_StatCritRating:   0.7,
		proto.Stat_StatHasteRating:  0.15,
		proto.Stat_StatIntellect:    0.2,
		proto.Stat_StatSpirit:       0.15,
		proto.Stat_StatMP5:          0.4,
	} {
		ep.Stats[stat] = value
	}
	player.Equipment = NewGearPicker(player, level).Pick(ep)
	swr.SimOptions = &proto.SimOptions{Iterations: 3000, RandomSeed: 101}
	return swr
}

// At healing capacity the weights count throughput and mana together. The site's group fight leaves the
// healer mana to spare at the level's own damage, so there healing power only overheals and its weight
// sits near 0; the search raises the damage until the healer stops keeping up, so the multiple is above
// 1, and healing power's weight is above 0 by more than its 90% interval.
//
// The bands come from what a point of each stat buys at most. A point of healing power adds most of a
// point to each of the few dozen heals of the 180-second fight, so it is worth 20 healing or more, and
// a point of mana is worth at most the 6 healing of the cheapest heal per point of mana it costs.
// Intellect is 15 mana a point and spirit regenerates less, so either is under 15 × 6 / 20, or 5; a
// point of MP5 is 36 mana over the fight, under 12. A point of crit rating is a small part of a percent
// of crit, worth a share of each heal smaller than healing power's, so it is under 5 as well.
func TestHealingCapacityWeighsThroughputAndMana(t *testing.T) {
	bands := map[proto.Stat]float64{
		proto.Stat_StatIntellect:  5,
		proto.Stat_StatSpirit:     5,
		proto.Stat_StatCritRating: 5,
		proto.Stat_StatMP5:        12,
	}
	for _, level := range []int32{30, 60} {
		t.Run(fmt.Sprintf("level %d", level), func(t *testing.T) {
			result := core.StatWeights(healerPriestCapacityRequest(t, level))
			if result.Error != nil {
				t.Fatalf("the stat weights failed: %s", result.Error.Message)
			}
			if !(result.HealingCapacityMultiplier > 1) {
				t.Errorf("weighed at %.3fx the level's damage, which the healer already keeps up with", result.HealingCapacityMultiplier)
			}
			hp := proto.Stat_StatHealingPower
			weight, samples := result.Hps.Weights.Stats[hp], result.SampleCounts.Stats[hp]
			if ci := core.StatWeightZ90 * result.Hps.WeightsStdev.Stats[hp] / math.Sqrt(samples); !(weight > ci) {
				t.Errorf("healing power weighs %.4f ± %.4f effective HPS a point", weight, ci)
			}
			for stat, band := range bands {
				ep := result.Hps.EpValues.Stats[stat]
				if !(result.Hps.Weights.Stats[stat] > 0) || !(ep < band) {
					t.Errorf("%s: EP %.3f (weight %.4f), want above 0 and below %.0f", stat, ep, result.Hps.Weights.Stats[stat], band)
				}
			}
			t.Logf("at %.3fx the level's damage, healing power weighs %.4f a point; EP intellect %.2f, spirit %.2f, crit %.2f, MP5 %.2f",
				result.HealingCapacityMultiplier, weight, result.Hps.EpValues.Stats[proto.Stat_StatIntellect], result.Hps.EpValues.Stats[proto.Stat_StatSpirit],
				result.Hps.EpValues.Stats[proto.Stat_StatCritRating], result.Hps.EpValues.Stats[proto.Stat_StatMP5])
		})
	}
}

// The search reads the same multiples on the same seed, so a request weighs at one capacity every time.
func TestHealingCapacityIsRepeatable(t *testing.T) {
	first := core.StatWeights(healerPriestCapacityRequest(t, 30))
	second := core.StatWeights(healerPriestCapacityRequest(t, 30))
	if first.HealingCapacityMultiplier != second.HealingCapacityMultiplier {
		t.Errorf("capacity %.4f, then %.4f", first.HealingCapacityMultiplier, second.HealingCapacityMultiplier)
	}
	hp := proto.Stat_StatHealingPower
	if first.Hps.Weights.Stats[hp] != second.Hps.Weights.Stats[hp] {
		t.Errorf("healing power weighs %.6f, then %.6f", first.Hps.Weights.Stats[hp], second.Hps.Weights.Stats[hp])
	}
}
