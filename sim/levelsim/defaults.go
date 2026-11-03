package levelsim

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// The most rounds of picking gear and re-weighing before the set is taken as it stands.
const maxRounds = 6

// Defaults gives a level's starting talents, BIS gear and EP for a spec's setup. The talents are the
// setup's spent up to the level's points (TrimTalents); the target is a creature of the level, with
// its ExpectedStat armor. The gear starts from every weighed stat at 1 EP and weapon damage at 10,
// then the set and the EP it is worn with alternate until the set stops changing.
func Defaults(req *proto.LevelDefaultsRequest) (result *proto.LevelDefaultsResult) {
	defer func() {
		if r := recover(); r != nil {
			result = &proto.LevelDefaultsResult{Error: &proto.ErrorOutcome{Message: fmt.Sprint(r)}}
		}
	}()
	level := req.Level
	if level < 1 || level > core.CharacterLevel {
		return &proto.LevelDefaultsResult{Error: &proto.ErrorOutcome{Message: fmt.Sprintf("level %d is outside 1-%d", level, core.CharacterLevel)}}
	}
	swr := googleProto.Clone(req.StatWeights).(*proto.StatWeightsRequest)
	player := swr.Player
	player.Level = level
	player.TalentsString = TrimTalents(player.TalentsString, core.TalentPointsAtLevel(level))

	_, armor, _ := core.CreatureStatsAtLevel(level)
	for _, target := range swr.Encounter.GetTargets() {
		target.Level = level
		if int(stats.Armor) < len(target.Stats) {
			target.Stats[stats.Armor] = armor
		}
	}

	ep := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
	for _, s := range swr.StatsToWeigh {
		ep.Stats[s] = 1
	}
	for _, s := range swr.PseudoStatsToWeigh {
		ep.PseudoStats[s] = 10
	}

	picker := NewGearPicker(player, level)
	var last *proto.EquipmentSpec
	var dps *proto.StatWeightValues
	rounds := int32(0)
	for rounds < maxRounds {
		set := picker.Pick(ep)
		if SameGear(set, last) {
			break
		}
		rounds++
		player.Equipment, last = set, set
		round := googleProto.Clone(swr).(*proto.StatWeightsRequest)
		sw := core.StatWeights(round)
		if sw.Error != nil {
			return &proto.LevelDefaultsResult{Error: sw.Error}
		}
		dps = sw.Dps
		ep = dps.EpValues
	}
	return &proto.LevelDefaultsResult{
		TalentsString: player.TalentsString,
		Equipment:     player.Equipment,
		Dps:           dps,
		TargetArmor:   armor,
		Rounds:        rounds,
	}
}

// BisGear is the best gear in slot for a level without raid gear by the EP given; it sims nothing.
func BisGear(req *proto.BisGearRequest) (result *proto.BisGearResult) {
	defer func() {
		if r := recover(); r != nil {
			result = &proto.BisGearResult{Error: &proto.ErrorOutcome{Message: fmt.Sprint(r)}}
		}
	}()
	return &proto.BisGearResult{Equipment: NewGearPicker(req.Player, core.PlayerLevel(&proto.Player{Level: req.Level})).Pick(req.Ep)}
}
