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

// The iterations each round's stat weights run at most. The weights only choose gear here, and the
// paired, fixed-seed stat weights hold steady at this count; the Stat Weights action refines them.
const maxRoundIterations = 4000

// Defaults gives a level's starting talents, BIS gear and EP for a spec's setup. The talents are the
// setup's spent up to the level's points (TrimTalents); the target is a creature of the level, with
// its ExpectedStat armor. The gear starts from every weighed stat at 1 EP and weapon damage at 10,
// then the set and the EP it is worn with alternate until the set stops changing. Where the set could
// dual wield or wear a two-hander, a sim of each settles it.
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
	if swr.SimOptions == nil {
		swr.SimOptions = &proto.SimOptions{}
	}
	if swr.SimOptions.Iterations == 0 || swr.SimOptions.Iterations > maxRoundIterations {
		swr.SimOptions.Iterations = maxRoundIterations
	}
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
	weigh := func(set *proto.EquipmentSpec) *proto.ErrorOutcome {
		rounds++
		player.Equipment, last = set, set
		sw := core.StatWeights(googleProto.Clone(swr).(*proto.StatWeightsRequest))
		if sw.Error != nil {
			return sw.Error
		}
		dps = sw.Dps
		ep = dps.EpValues
		return nil
	}
	for rounds < maxRounds {
		set := picker.Pick(ep)
		if SameGear(set, last) {
			break
		}
		if err := weigh(set); err != nil {
			return &proto.LevelDefaultsResult{Error: err}
		}
	}

	// Dual wielding adds to every white swing's miss chance, which no EP sees, so a two-hander and a
	// main and off hand pair are each simmed with the rest of the set, and the run keeps the higher
	// DPS, with its own EP.
	other := handsTwo
	if handsWorn(last) == handsTwo {
		other = handsOne
	}
	if alt := picker.pick(ep, other); !SameGear(alt, last) && handsWorn(alt) != handsWorn(last) && (dualWields(alt) || dualWields(last)) {
		lastDps, err := simDps(swr, last)
		if err != nil {
			return &proto.LevelDefaultsResult{Error: err}
		}
		altDps, err := simDps(swr, alt)
		if err != nil {
			return &proto.LevelDefaultsResult{Error: err}
		}
		if altDps > lastDps {
			if err := weigh(alt); err != nil {
				return &proto.LevelDefaultsResult{Error: err}
			}
		}
	}
	return &proto.LevelDefaultsResult{
		TalentsString: player.TalentsString,
		Equipment:     player.Equipment,
		Dps:           dps,
		TargetArmor:   armor,
		Rounds:        rounds,
	}
}

// The DPS of the stat weights request's player in a set, from a plain sim of the same request at its
// iterations. Both sides of a comparison run on one seed, so they meet the same rolls.
func simDps(swr *proto.StatWeightsRequest, set *proto.EquipmentSpec) (float64, *proto.ErrorOutcome) {
	player := googleProto.Clone(swr.Player).(*proto.Player)
	player.Equipment = set
	raid := core.SinglePlayerRaidProto(player, swr.PartyBuffs, swr.RaidBuffs, swr.Debuffs)
	raid.Tanks = swr.Tanks
	options := googleProto.Clone(swr.SimOptions).(*proto.SimOptions)
	if options.RandomSeed == 0 {
		options.RandomSeed = 1
	}
	options.UseLabeledRands = true
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  googleProto.Clone(swr.Encounter).(*proto.Encounter),
		SimOptions: options,
		Type:       proto.SimType_SimTypeIndividual,
	})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RaidMetrics.Dps.Avg, nil
}

// BisGear is the best gear in slot for a level without raid gear by the EP given; it sims nothing, so
// it keeps the weapons the player wears, a two-hander or a main and off hand, which Defaults chose by
// sim, and weighs the two by EP only for a player with no weapon.
func BisGear(req *proto.BisGearRequest) (result *proto.BisGearResult) {
	defer func() {
		if r := recover(); r != nil {
			result = &proto.BisGearResult{Error: &proto.ErrorOutcome{Message: fmt.Sprint(r)}}
		}
	}()
	picker := NewGearPicker(req.Player, core.PlayerLevel(&proto.Player{Level: req.Level}))
	return &proto.BisGearResult{Equipment: picker.pick(req.Ep, handsWorn(req.Player.Equipment))}
}
