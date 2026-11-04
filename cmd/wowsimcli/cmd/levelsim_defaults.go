package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/levelsim"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

var (
	levelDefaultsRequests   string
	levelDefaultsLevels     string
	levelDefaultsOut        string
	levelDefaultsSpecs      string
	levelDefaultsJobs       int
	levelDefaultsIterations int32
	levelDefaultsRedo       bool
)

var levelDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "precompute each spec's level defaults (talents, gear, EP, target armor) for the site",
	Long: `Runs levelsim.Defaults, the sim a level edit on the site runs, for each spec's LevelDefaultsRequest (a
<class>_<spec>.json in protojson, as the page sends it with the spec's default talents) at each level, and writes
one <class>_<spec>.json per spec into the out folder, with a creature's armor at every level from 1 to 60. The site
applies a level the file holds as it stands; any other takes the EP of the level below it, and the site trims the
talents and picks the gear for it without a sim. An out file that exists is added to: its levels are kept unless
--redo is given, so a stopped run picks up where it left off. A request with no random seed runs on seed 1, so a
run is repeatable.`,
	Run: levelDefaultsMain,
}

func init() {
	levelDefaultsCmd.Flags().StringVar(&levelDefaultsRequests, "requests", "tools/levelsim/requests", "the folder of <class>_<spec>.json LevelDefaultsRequests")
	levelDefaultsCmd.Flags().StringVar(&levelDefaultsLevels, "levels", "10,20,30,40,50", "the levels to work out: a range (10-59), a list (10,20,60) or both")
	levelDefaultsCmd.Flags().StringVar(&levelDefaultsOut, "out", "assets/levelsim/defaults", "the folder to write each spec's defaults to")
	levelDefaultsCmd.Flags().StringVar(&levelDefaultsSpecs, "specs", "", "only these request names, comma separated (warrior_dps,...); every request when unset")
	levelDefaultsCmd.Flags().IntVar(&levelDefaultsJobs, "jobs", runtime.NumCPU(), "how many levels run at once; each one's sims split across the CPUs as well")
	levelDefaultsCmd.Flags().Int32Var(&levelDefaultsIterations, "iterations", 0, "iterations per stat weights round; 0 is the request's own, which Defaults caps")
	levelDefaultsCmd.Flags().BoolVar(&levelDefaultsRedo, "redo", false, "work out levels the out file already holds again")
	levelsimCmd.AddCommand(levelDefaultsCmd)
}

// One level's defaults as the site applies them: the trimmed talents, the gear by item, the EP and the target's armor.
type levelDefaultsEntry struct {
	Talents     string            `json:"talents"`
	Equipment   []json.RawMessage `json:"equipment"`
	EP          levelDefaultsEP   `json:"ep"`
	TargetArmor float64           `json:"targetArmor"`
}

// The EP as a UnitStats in protojson, indexed by Stat and PseudoStat.
type levelDefaultsEP struct {
	Stats       []float64 `json:"stats"`
	PseudoStats []float64 `json:"pseudoStats"`
}

type levelDefaultsFile struct {
	Spec          string `json:"spec"`
	GeneratedFrom string `json:"generatedFrom"`
	Date          string `json:"date"`
	// The armor of a creature of each level, 1 to 60 from index 0, which the site gives a level the file holds no
	// entry for.
	CreatureArmor []float64                     `json:"creatureArmor"`
	Levels        map[string]levelDefaultsEntry `json:"levels"`
}

// core.CreatureStatsAtLevel's armor at each level from 1 to 60.
func creatureArmorByLevel() []float64 {
	armor := make([]float64, core.CharacterLevel)
	for level := range armor {
		_, armor[level], _ = core.CreatureStatsAtLevel(int32(level + 1))
	}
	return armor
}

// The levels a --levels value names, in order: "1-59", "10,20,60" or "1-9,60".
func parseLevels(value string) ([]int32, error) {
	seen := map[int32]bool{}
	var levels []int32
	for part := range strings.SplitSeq(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lo, hi, isRange := strings.Cut(part, "-")
		from, err := strconv.Atoi(lo)
		if err != nil {
			return nil, fmt.Errorf("level %q is not a number", lo)
		}
		to := from
		if isRange {
			if to, err = strconv.Atoi(hi); err != nil {
				return nil, fmt.Errorf("level %q is not a number", hi)
			}
		}
		if from < 1 || to > 60 || from > to {
			return nil, fmt.Errorf("levels %q are outside 1-60", part)
		}
		for l := int32(from); l <= int32(to); l++ {
			if !seen[l] {
				seen[l] = true
				levels = append(levels, l)
			}
		}
	}
	sort.Slice(levels, func(a, b int) bool { return levels[a] < levels[b] })
	return levels, nil
}

// An EP value to four decimals: the gear picker and the UI's stat priority read no finer, and the file stays small.
func roundEP(values []float64) []float64 {
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = math.Round(v*1e4) / 1e4
	}
	return out
}

// A Defaults result as the file holds it. A request that weighs effective healing keeps the HPS EP its gear was picked by.
func levelDefaultsEntryOf(req *proto.LevelDefaultsRequest, result *proto.LevelDefaultsResult) (levelDefaultsEntry, error) {
	entry := levelDefaultsEntry{Talents: result.TalentsString, TargetArmor: result.TargetArmor, Equipment: []json.RawMessage{}}
	for _, item := range result.Equipment.GetItems() {
		encoded, err := protojson.Marshal(item)
		if err != nil {
			return entry, err
		}
		entry.Equipment = append(entry.Equipment, encoded)
	}
	ep := result.Dps.GetEpValues()
	if req.WeighEffectiveHealing {
		ep = result.Hps.GetEpValues()
	}
	if ep == nil {
		return entry, fmt.Errorf("the result holds no EP")
	}
	entry.EP = levelDefaultsEP{Stats: roundEP(ep.Stats), PseudoStats: roundEP(ep.PseudoStats)}
	return entry, nil
}

func readLevelDefaultsRequest(path string) (*proto.LevelDefaultsRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	req := &proto.LevelDefaultsRequest{}
	if err := protojson.Unmarshal(data, req); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if req.StatWeights.GetPlayer() == nil {
		return nil, fmt.Errorf("%s holds no stat weights player", path)
	}
	return req, nil
}

func readLevelDefaultsFile(path, spec string) (*levelDefaultsFile, error) {
	file := &levelDefaultsFile{Spec: spec, Levels: map[string]levelDefaultsEntry{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return file, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, file); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if file.Levels == nil {
		file.Levels = map[string]levelDefaultsEntry{}
	}
	return file, nil
}

// Writes the file whole through a temporary one, so a run stopped mid write leaves the last good file.
func writeLevelDefaultsFile(path string, file *levelDefaultsFile) error {
	encoded, err := json.Marshal(file)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(encoded, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

type levelDefaultsJob struct {
	spec  string
	req   *proto.LevelDefaultsRequest
	level int32
}

type levelDefaultsSpecState struct {
	path string
	file *levelDefaultsFile
}

// runLevelDefaults works out every request's levels the out files lack (or all of them with redo), jobs levels at
// once, and writes each spec's file as each level lands. It returns the levels that failed, by spec and level.
func runLevelDefaults(requestsDir, outDir string, specs []string, levels []int32, jobs int, iterations int32, redo bool, logf func(string, ...any)) ([]string, error) {
	if len(specs) == 0 {
		matches, err := filepath.Glob(filepath.Join(requestsDir, "*.json"))
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			specs = append(specs, strings.TrimSuffix(filepath.Base(m), ".json"))
		}
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("%s holds no requests", requestsDir)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	commit := forkCommit()

	states := map[string]*levelDefaultsSpecState{}
	var queue []levelDefaultsJob
	for _, spec := range specs {
		req, err := readLevelDefaultsRequest(filepath.Join(requestsDir, spec+".json"))
		if err != nil {
			return nil, err
		}
		if req.StatWeights.SimOptions == nil {
			req.StatWeights.SimOptions = &proto.SimOptions{}
		}
		if iterations > 0 {
			req.StatWeights.SimOptions.Iterations = iterations
		}
		if req.StatWeights.SimOptions.RandomSeed == 0 {
			req.StatWeights.SimOptions.RandomSeed = 1
		}
		path := filepath.Join(outDir, spec+".json")
		file, err := readLevelDefaultsFile(path, spec)
		if err != nil {
			return nil, err
		}
		file.CreatureArmor = creatureArmorByLevel()
		states[spec] = &levelDefaultsSpecState{path: path, file: file}
		for _, level := range levels {
			if _, done := file.Levels[strconv.Itoa(int(level))]; done && !redo {
				continue
			}
			queue = append(queue, levelDefaultsJob{spec: spec, req: req, level: level})
		}
	}
	logf("%d levels to work out across %d specs, %d at once", len(queue), len(specs), jobs)

	var mu sync.Mutex
	var failed []string
	var writeErr error
	done := 0
	work := make(chan levelDefaultsJob)
	var wg sync.WaitGroup
	for range max(1, jobs) {
		wg.Go(func() {
			for job := range work {
				start := time.Now()
				req := googleProto.Clone(job.req).(*proto.LevelDefaultsRequest)
				req.Level = job.level
				result := levelsim.Defaults(req)
				var entry levelDefaultsEntry
				var err error
				if result.Error != nil {
					err = fmt.Errorf("%s", result.Error.Message)
				} else {
					entry, err = levelDefaultsEntryOf(req, result)
				}

				mu.Lock()
				done++
				if err != nil {
					failed = append(failed, fmt.Sprintf("%s %d: %v", job.spec, job.level, err))
					logf("[%d/%d] %s level %d FAILED in %s: %v", done, len(queue), job.spec, job.level, time.Since(start).Round(time.Second), err)
				} else {
					state := states[job.spec]
					state.file.Levels[strconv.Itoa(int(job.level))] = entry
					state.file.GeneratedFrom = commit
					state.file.Date = time.Now().UTC().Format(time.RFC3339)
					if werr := writeLevelDefaultsFile(state.path, state.file); werr != nil && writeErr == nil {
						writeErr = werr
					}
					capacity := ""
					if result.HealingCapacityMultiplier > 0 {
						capacity = fmt.Sprintf(", weighed at %.2fx the level's damage", result.HealingCapacityMultiplier)
					}
					logf("[%d/%d] %s level %d in %s, %d rounds%s", done, len(queue), job.spec, job.level, time.Since(start).Round(time.Second), result.Rounds, capacity)
				}
				mu.Unlock()
			}
		})
	}
	for _, job := range queue {
		work <- job
	}
	close(work)
	wg.Wait()
	sort.Strings(failed)
	return failed, writeErr
}

func levelDefaultsMain(cmd *cobra.Command, args []string) {
	levels, err := parseLevels(levelDefaultsLevels)
	if err != nil {
		log.Fatal(err)
	}
	var specs []string
	for s := range strings.SplitSeq(levelDefaultsSpecs, ",") {
		if s = strings.TrimSpace(s); s != "" {
			specs = append(specs, s)
		}
	}
	start := time.Now()
	// The sims log every run they start, which would bury the progress lines.
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)
	logf := func(format string, a ...any) {
		fmt.Fprintf(os.Stderr, "%s %s\n", time.Since(start).Round(time.Second), fmt.Sprintf(format, a...))
	}
	failed, err := runLevelDefaults(levelDefaultsRequests, levelDefaultsOut, specs, levels, levelDefaultsJobs, levelDefaultsIterations, levelDefaultsRedo, logf)
	log.SetOutput(os.Stderr)
	if err != nil {
		log.Fatal(err)
	}
	logf("done")
	if len(failed) > 0 {
		log.Fatalf("%d levels failed:\n%s", len(failed), strings.Join(failed, "\n"))
	}
}
