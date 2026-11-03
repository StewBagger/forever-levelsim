//go:build with_db

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestParseLevels(t *testing.T) {
	got, err := parseLevels("12-14,60,1, 13")
	if err != nil {
		t.Fatal(err)
	}
	if want := []int32{1, 12, 13, 14, 60}; !reflect.DeepEqual(got, want) {
		t.Errorf("parseLevels gave %v, want %v", got, want)
	}
	for _, bad := range []string{"0-10", "50-61", "20-10", "ten"} {
		if _, err := parseLevels(bad); err == nil {
			t.Errorf("parseLevels(%q) took a level outside 1-60", bad)
		}
	}
}

// A run on the committed warrior request at two levels, at few iterations, writes one file for the spec with each
// level's talents in the level's points, gear by item, EP by stat and the level's creature armor, and a second run
// keeps what the file holds instead of working it out again.
func TestRunLevelDefaultsWritesEachLevel(t *testing.T) {
	out := t.TempDir()
	var lines []string
	logf := func(format string, a ...any) { lines = append(lines, format) }
	failed, err := runLevelDefaults(filepath.Join(root, "tools/levelsim/requests"), out, []string{"warrior_dps"}, []int32{10, 20}, 2, 100, false, logf)
	if err != nil || len(failed) > 0 {
		t.Fatalf("the run failed: %v %v", err, failed)
	}

	data, err := os.ReadFile(filepath.Join(out, "warrior_dps.json"))
	if err != nil {
		t.Fatal(err)
	}
	var file levelDefaultsFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	if file.Spec != "warrior_dps" || file.GeneratedFrom == "" || file.Date == "" {
		t.Errorf("the file names spec %q, commit %q and date %q", file.Spec, file.GeneratedFrom, file.Date)
	}
	if len(file.CreatureArmor) != 60 {
		t.Fatalf("the file holds %d creature armors, want one for each level to 60", len(file.CreatureArmor))
	}
	if len(file.Levels) != 2 {
		t.Fatalf("the file holds levels %v, want 10 and 20", file.Levels)
	}
	for _, level := range []int32{10, 20} {
		entry, ok := file.Levels[strconv.Itoa(int(level))]
		if !ok {
			t.Fatalf("the file holds no level %d", level)
		}
		points := int32(0)
		for _, c := range entry.Talents {
			if c >= '0' && c <= '9' {
				points += c - '0'
			}
		}
		if points != core.TalentPointsAtLevel(level) {
			t.Errorf("level %d talents %q spend %d points, want %d", level, entry.Talents, points, core.TalentPointsAtLevel(level))
		}
		equipment := &proto.EquipmentSpec{}
		items, _ := json.Marshal(map[string]any{"items": entry.Equipment})
		if err := protojson.Unmarshal(items, equipment); err != nil {
			t.Fatalf("level %d gear does not read as an EquipmentSpec: %v", level, err)
		}
		worn := 0
		for _, item := range equipment.Items {
			if item.Id != 0 {
				worn++
			}
		}
		if worn == 0 {
			t.Errorf("level %d wears no gear", level)
		}
		if len(entry.EP.Stats) != len(proto.Stat_name) || len(entry.EP.PseudoStats) != len(proto.PseudoStat_name) {
			t.Errorf("level %d EP holds %d stats and %d pseudo stats", level, len(entry.EP.Stats), len(entry.EP.PseudoStats))
		} else if entry.EP.Stats[proto.Stat_StatStrength] != 1 {
			t.Errorf("level %d Strength EP is %v, want 1 as the reference stat", level, entry.EP.Stats[proto.Stat_StatStrength])
		}
		if _, armor, _ := core.CreatureStatsAtLevel(level); entry.TargetArmor != armor || file.CreatureArmor[level-1] != armor {
			t.Errorf("level %d target armor is %v and its creature's %v, want %v", level, entry.TargetArmor, file.CreatureArmor[level-1], armor)
		}
	}

	lines = nil
	if _, err := runLevelDefaults(filepath.Join(root, "tools/levelsim/requests"), out, []string{"warrior_dps"}, []int32{10, 20}, 2, 100, false, logf); err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Errorf("a second run logged %d lines, want only its plan of no levels: %v", len(lines), lines)
	}
}

// A healer's file holds the HPS EP its gear was picked by, a damage dealer's the DPS EP.
func TestLevelDefaultsEntryKeepsTheWeighedEP(t *testing.T) {
	ep := func(healingPower float64) *proto.StatWeightValues {
		stats := make([]float64, len(proto.Stat_name))
		stats[proto.Stat_StatHealingPower] = healingPower
		return &proto.StatWeightValues{EpValues: &proto.UnitStats{Stats: stats}}
	}
	result := &proto.LevelDefaultsResult{Equipment: &proto.EquipmentSpec{}, Dps: ep(2), Hps: ep(1)}
	for _, tc := range []struct {
		healing bool
		want    float64
	}{{false, 2}, {true, 1}} {
		entry, err := levelDefaultsEntryOf(&proto.LevelDefaultsRequest{WeighEffectiveHealing: tc.healing}, result)
		if err != nil {
			t.Fatal(err)
		}
		if got := entry.EP.Stats[proto.Stat_StatHealingPower]; got != tc.want {
			t.Errorf("weigh effective healing %v: healing power EP %v, want %v", tc.healing, got, tc.want)
		}
	}
}
