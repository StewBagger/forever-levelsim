package rogue

import (
	"reflect"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Below 60 the rogue registers only the ranks taught at its level, and the default rotation sims
// through the abilities it has not learned yet. Only the class's own spells are checked: a gear
// proc's row states the item's level, not a level the rogue learns it at.
func TestLevelledRogueKnowsOnlyTaughtRanks(t *testing.T) {
	classSpells := map[int32]bool{}
	ladders := reflect.ValueOf(spellData)
	for i := range ladders.NumField() {
		ladders.Field(i).Interface().(spelldata.Ladder).Each(func(_ int32, s *spelldata.Spell) {
			classSpells[s.ID] = true
		})
	}

	rotation := core.GetAplRotation("../../ui/specs/rogue/dps/apls", "default").Rotation
	// The last case spends 31 points in Assassination for Mutilate, taught at 30, and Cold Blood.
	cases := []struct {
		level   int32
		talents string
	}{{10, ""}, {20, ""}, {30, ""}, {40, ""}, {40, "00532310151501031"}}
	for _, c := range cases {
		level := c.level
		player := core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassRogue,
			Race:          proto.Race_RaceHuman,
			Level:         level,
			TalentsString: c.talents,
			Equipment:     core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "p1_muti").GearSet,
			Consumables:   DefaultConsumables,
			Rotation:      rotation,
		}, DefaultOptions)
		request := &proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 5},
		}

		result := core.RunRaidSim(request)
		if result.Error != nil {
			t.Fatalf("level %d: the sim failed: %s", level, result.Error.Message)
		}
		if dps := result.RaidMetrics.Dps.Avg; dps <= 0 {
			t.Errorf("level %d: the rogue deals %.0f dps, want some", level, dps)
		}

		sim := core.NewSim(request, simsignals.CreateSignals())
		rogue := sim.Raid.Parties[0].Players[0].(*Rogue)
		for _, spell := range rogue.Spellbook {
			id := spell.ActionID.SpellID
			if !classSpells[id] || spell.ActionID.ItemID != 0 {
				continue
			}
			row := spelldata.MustFind(id)
			if taught := spelldata.TaughtAt(row); taught > level {
				t.Errorf("level %d: %s (%d) is registered, taught at %d", level, row.RankedName(), id, taught)
			}
		}
	}
}
