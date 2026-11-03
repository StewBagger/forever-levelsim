package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	levelsimPlan string
	levelsimOut  string
)

var levelsimCmd = &cobra.Command{
	Use:   "levelsim",
	Short: "sim a plan of setups below level 60 for DPS and stat weights",
	Long: `Sims each run of a plan file: a setup (an x.siml, or a RaidSimRequest in protojson) at a level, with a talent
string, against targets of the same level, and writes the DPS and stat weights of every run as JSON with the fork
commit and the run date. A run is solo by default: no raid, party or individual buffs, no debuffs from others and no
consumables, since a levelling character has none of the level 60 ones a setup carries.`,
	Run: levelsimMain,
}

func init() {
	levelsimCmd.Flags().StringVar(&levelsimPlan, "plan", "", "the plan file (JSON, see levelsimPlanFile)")
	levelsimCmd.Flags().StringVar(&levelsimOut, "out", "", "where to write the results (JSON); stdout when unset")
	levelsimCmd.MarkFlagRequired("plan")
}

type levelsimPlanFile struct {
	// Iterations per stat weight pass; 0 is the setup's own or the site's default.
	Iterations int32         `json:"iterations"`
	Runs       []levelsimRun `json:"runs"`
}

type levelsimRun struct {
	Name    string `json:"name"`
	Setup   string `json:"setup"`
	Level   int32  `json:"level"`
	Talents string `json:"talents"`
	// The stats to weigh, by their proto names (StatStrength, StatHitRating, ...), and pseudo stats
	// (PseudoStatMainHandDps, ...). The reference stat's weight is 1.
	Stats       []string `json:"stats"`
	PseudoStats []string `json:"pseudoStats"`
	Reference   string   `json:"reference"`
	// Keep the setup's buffs, debuffs and consumables instead of simming the character alone.
	KeepBuffs bool `json:"keepBuffs"`
	// The target's level relative to the character's; 0 is a same-level target.
	TargetLevelOffset int32 `json:"targetLevelOffset"`
	// Keep the setup's target stats (a raid boss's armor) instead of a creature of the target's level.
	KeepTarget bool `json:"keepTarget"`
	// "setup" (or unset) wears the setup's gear; "bis" the best in slot for the level with no raid gear.
	Gear string `json:"gear"`
}

type levelsimResult struct {
	Name        string             `json:"name"`
	Level       int32              `json:"level"`
	TargetLevel int32              `json:"targetLevel"`
	TargetArmor float64            `json:"targetArmor"`
	Talents     string             `json:"talents"`
	Setup       string             `json:"setup"`
	Solo        bool               `json:"solo"`
	Gear        []string           `json:"gear,omitempty"`
	GearRounds  int                `json:"gearRounds,omitempty"`
	Dps         float64            `json:"dps"`
	DpsStdev    float64            `json:"dpsStdev"`
	Reference   string             `json:"reference"`
	Weights     map[string]float64 `json:"weights"`
	EP          map[string]float64 `json:"ep"`
	EPStdev     map[string]float64 `json:"epStdev"`
	Warnings    []string           `json:"warnings,omitempty"`
}

type levelsimOutput struct {
	Commit  string           `json:"commit"`
	RunDate string           `json:"runDate"`
	Results []levelsimResult `json:"results"`
}

func levelsimMain(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(levelsimPlan)
	if err != nil {
		log.Fatal(err)
	}
	var plan levelsimPlanFile
	if err := json.Unmarshal(data, &plan); err != nil {
		log.Fatalf("%s: %v", levelsimPlan, err)
	}

	out := levelsimOutput{Commit: forkCommit(), RunDate: time.Now().UTC().Format(time.RFC3339)}
	for _, run := range plan.Runs {
		result, err := levelsimOne(run, plan.Iterations)
		if err != nil {
			log.Fatalf("run %q: %v", run.Name, err)
		}
		fmt.Fprintf(os.Stderr, "%s: level %d, %.1f dps\n", run.Name, run.Level, result.Dps)
		out.Results = append(out.Results, result)
	}

	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if levelsimOut == "" {
		fmt.Println(string(encoded))
		return
	}
	if err := os.WriteFile(levelsimOut, append(encoded, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}

func levelsimOne(run levelsimRun, iterations int32) (levelsimResult, error) {
	request, warnings, err := readSimRequest(run.Setup)
	if err != nil {
		return levelsimResult{}, err
	}
	if run.Level < 1 || run.Level > core.CharacterLevel {
		return levelsimResult{}, fmt.Errorf("level %d is outside 1-%d", run.Level, core.CharacterLevel)
	}
	player := request.Raid.Parties[0].Players[0]
	player.Level = run.Level
	player.TalentsString = run.Talents

	partyBuffs, raidBuffs, debuffs := request.Raid.Parties[0].Buffs, request.Raid.Buffs, request.Raid.Debuffs
	if !run.KeepBuffs {
		player.Buffs = &proto.IndividualBuffs{}
		player.Consumables = &proto.ConsumesSpec{}
		partyBuffs, raidBuffs, debuffs = &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}
	}

	encounter := googleProto.Clone(request.Encounter).(*proto.Encounter)
	// A levelling target is an ordinary creature of its level: ExpectedStat's armor for that level in
	// place of the raid boss's, unless the run keeps the setup's target.
	targetLevel := run.Level + run.TargetLevelOffset
	_, creatureArmor, _ := core.CreatureStatsAtLevel(targetLevel)
	for _, target := range encounter.Targets {
		target.Level = targetLevel
		if !run.KeepTarget && int(stats.Armor) < len(target.Stats) {
			target.Stats[stats.Armor] = creatureArmor
		}
	}

	var statsToWeigh []proto.Stat
	var pseudoStatsToWeigh []proto.PseudoStat
	for _, name := range run.Stats {
		v, ok := proto.Stat_value[name]
		if !ok {
			return levelsimResult{}, fmt.Errorf("unknown stat %q", name)
		}
		statsToWeigh = append(statsToWeigh, proto.Stat(v))
	}
	for _, name := range run.PseudoStats {
		v, ok := proto.PseudoStat_value[name]
		if !ok {
			return levelsimResult{}, fmt.Errorf("unknown pseudo stat %q", name)
		}
		pseudoStatsToWeigh = append(pseudoStatsToWeigh, proto.PseudoStat(v))
	}
	ref, ok := proto.Stat_value[run.Reference]
	if !ok {
		return levelsimResult{}, fmt.Errorf("unknown reference stat %q", run.Reference)
	}
	simOptions := googleProto.Clone(request.SimOptions).(*proto.SimOptions)
	if iterations > 0 {
		simOptions.Iterations = iterations
	}

	// The stat weights of the player as it stands. Each call gets its own options, which the stat
	// weights request rewrites.
	weigh := func() (*proto.StatWeightValues, error) {
		sw := core.StatWeights(&proto.StatWeightsRequest{
			Player:             player,
			RaidBuffs:          raidBuffs,
			PartyBuffs:         partyBuffs,
			Debuffs:            debuffs,
			Encounter:          encounter,
			SimOptions:         googleProto.Clone(simOptions).(*proto.SimOptions),
			Tanks:              request.Raid.Tanks,
			StatsToWeigh:       statsToWeigh,
			PseudoStatsToWeigh: pseudoStatsToWeigh,
			EpReferenceStat:    proto.Stat(ref),
			HpsReferenceStat:   proto.Stat(ref),
		})
		if os.Getenv("LEVELSIM_DEBUG") != "" {
			fmt.Fprintln(os.Stderr, protojson.Format(sw))
		}
		if sw.Error != nil {
			return nil, fmt.Errorf("%s", sw.Error.Message)
		}
		return sw.Dps, nil
	}
	epOf := func(dps *proto.StatWeightValues) map[string]float64 {
		ep := map[string]float64{}
		for _, s := range statsToWeigh {
			ep[s.String()] = dps.EpValues.Stats[s]
		}
		for _, s := range pseudoStatsToWeigh {
			ep[s.String()] = dps.EpValues.PseudoStats[s]
		}
		return ep
	}

	var gear []string
	var gearRounds int
	var dps *proto.StatWeightValues
	switch run.Gear {
	case "", "setup":
	case "bis":
		// Start from every weighed stat at 1 EP and weapon damage at 10, then alternate the set and
		// the EP it is worn with until the set stops changing.
		picker := newGearPicker(player, run.Level)
		ep := map[string]float64{}
		for _, s := range statsToWeigh {
			ep[s.String()] = 1
		}
		for _, s := range pseudoStatsToWeigh {
			ep[s.String()] = 10
		}
		var last *proto.EquipmentSpec
		for gearRounds = 1; gearRounds <= 6; gearRounds++ {
			set := picker.pick(ep)
			if sameGear(set, last) {
				break
			}
			player.Equipment, last = set, set
			if dps, err = weigh(); err != nil {
				return levelsimResult{}, err
			}
			ep = epOf(dps)
		}
		gear = gearNames(player.Equipment)
	default:
		return levelsimResult{}, fmt.Errorf("gear %q is neither setup nor bis", run.Gear)
	}
	if dps == nil {
		if dps, err = weigh(); err != nil {
			return levelsimResult{}, err
		}
	}

	// The DPS at the baseline, from a plain sim of the same request.
	raid := core.SinglePlayerRaidProto(player, partyBuffs, raidBuffs, debuffs)
	raid.Tanks = request.Raid.Tanks
	baseline := core.RunRaidSimConcurrent(&proto.RaidSimRequest{Raid: raid, Encounter: encounter, SimOptions: googleProto.Clone(simOptions).(*proto.SimOptions), Type: proto.SimType_SimTypeIndividual})
	if baseline.Error != nil {
		return levelsimResult{}, fmt.Errorf("%s", baseline.Error.Message)
	}

	result := levelsimResult{
		Name:        run.Name,
		Level:       run.Level,
		TargetLevel: targetLevel,
		TargetArmor: encounter.Targets[0].Stats[stats.Armor],
		Talents:     run.Talents,
		Setup:       run.Setup,
		Solo:        !run.KeepBuffs,
		Gear:        gear,
		GearRounds:  gearRounds,
		Dps:         baseline.RaidMetrics.Dps.Avg,
		DpsStdev:    baseline.RaidMetrics.Dps.Stdev,
		Reference:   run.Reference,
		Weights:     map[string]float64{},
		EP:          epOf(dps),
		EPStdev:     map[string]float64{},
		Warnings:    warnings,
	}
	for _, s := range statsToWeigh {
		result.Weights[s.String()] = dps.Weights.Stats[s]
		result.EPStdev[s.String()] = dps.EpValuesStdev.Stats[s]
	}
	for _, s := range pseudoStatsToWeigh {
		result.Weights[s.String()] = dps.Weights.PseudoStats[s]
		result.EPStdev[s.String()] = dps.EpValuesStdev.PseudoStats[s]
	}
	return result, nil
}

// The commit the binary runs from, marked dirty when the tree has changes; "unknown" outside a checkout.
func forkCommit() string {
	head, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	commit := strings.TrimSpace(string(head))
	if status, err := exec.Command("git", "status", "--porcelain", "--untracked-files=no").Output(); err == nil && len(status) > 0 {
		commit += "-dirty"
	}
	return commit
}
