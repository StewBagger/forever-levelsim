package hunter

import (
	"reflect"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Every rank of every ladder the hunter's spell data holds, pet side included: the class's spells,
// as against item, consumable and racial ones.
func hunterLadderSpells() map[int32]*spelldata.Spell {
	spells := make(map[int32]*spelldata.Spell)
	data := reflect.ValueOf(spellData)
	for i := range data.NumField() {
		if ladder, ok := data.Field(i).Interface().(spelldata.Ladder); ok {
			ladder.Each(func(_ int32, rank *spelldata.Spell) { spells[rank.ID] = rank })
		}
	}
	return spells
}

func levelSimRequest(level int32, spec *proto.Player_Hunter) *proto.RaidSimRequest {
	target := core.NewDefaultTarget()
	target.Level = level + 3
	player := &proto.Player{
		Name:               "Leveller",
		Race:               proto.Race_RaceOrc,
		Class:              proto.Class_ClassHunter,
		Level:              level,
		Equipment:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1").GearSet,
		Consumables:        DefaultConsumables,
		Spec:               spec,
		Rotation:           core.GetAplRotation("../../ui/specs/hunter/dps/apls", "default").Rotation,
		DistanceFromTarget: 12,
	}
	return &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Duration: core.ShortDuration,
			Targets:  []*proto.Target{target},
		},
		SimOptions: &proto.SimOptions{Iterations: 2, RandomSeed: 1},
	}
}

// A hunter below 60 with no talents, each pet family and the default rotation sims, and registers
// no rank, its own or its pet's, the trainer has not taught by its level.
func TestHunterBelowMaxLevelRegistersLearnedRanks(t *testing.T) {
	ladderSpells := hunterLadderSpells()
	for _, level := range []int32{10, 20, 30, 40} {
		for _, family := range petOptions() {
			checkLearnedRanks(t, ladderSpells, level, family.Label, family.SpecOptions.(*proto.Player_Hunter))
		}
	}
}

func checkLearnedRanks(t *testing.T, ladderSpells map[int32]*spelldata.Spell, level int32, family string, spec *proto.Player_Hunter) {
	t.Helper()
	rsr := levelSimRequest(level, spec)

	result := core.RunRaidSim(rsr)
	if result.Error != nil {
		t.Fatalf("level %d %s: the sim failed: %s", level, family, result.Error.Message)
	}
	if dps := result.RaidMetrics.Dps.Avg; dps <= 0 {
		t.Errorf("level %d %s: %.1f dps", level, family, dps)
	}

	sim := core.NewSim(rsr, simsignals.CreateSignals())
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	if hunter.Pet == nil {
		t.Errorf("level %d %s: no pet, where Call Pet is taught at %d", level, family, spelldata.TaughtAt(spellData.CallPet.Highest()))
	}
	units := []*core.Unit{&hunter.Unit}
	for _, pet := range hunter.Pets {
		units = append(units, &pet.Unit)
	}
	for _, unit := range units {
		for _, spell := range unit.Spellbook {
			rank, ok := ladderSpells[spell.SpellID]
			if !ok {
				continue
			}
			if taught := spelldata.TaughtAt(rank); taught > level {
				t.Errorf("level %d: %s registers %s (%d), taught at %d", level, unit.Label, rank.RankedName(), rank.ID, taught)
			}
		}
	}
}
