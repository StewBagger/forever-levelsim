package warrior_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Below 60 the warrior registers only the ranks taught at its level, and each spec's default
// rotation sims through the abilities it has not learned yet, Berserker Stance included. The
// warrior's own spells are the ones in its class family; a gear proc's row states the item's level,
// not a level the warrior learns it at, so the warrior wears only its weapons.
func TestLevelledWarriorKnowsOnlyTaughtRanks(t *testing.T) {
	warriorFamily := spelldata.MustFind(78).ClassFlags.Family // Heroic Strike Rank 1

	specs := []struct {
		name     string
		spec     any
		gearDir  string
		gearSet  string
		aplDir   string
		rotation string
	}{
		{"Fury", dpsOptions, "../../ui/specs/warrior/dps/gear_sets", "p1_fury", "../../ui/specs/warrior/dps/apls", "fury"},
		{"Arms", dpsOptions, "../../ui/specs/warrior/dps/gear_sets", "p1_arms", "../../ui/specs/warrior/dps/apls", "arms"},
		{"Protection", protectionOptions, "../../ui/specs/warrior/protection/gear_sets", "p1_bis", "../../ui/specs/warrior/protection/apls", "default"},
	}

	for _, s := range specs {
		weapons := &proto.EquipmentSpec{Items: make([]*proto.ItemSpec, len(proto.ItemSlot_name))}
		for slot, item := range core.GetGearSet(s.gearDir, s.gearSet).GearSet.Items {
			if slot == int(proto.ItemSlot_ItemSlotMainHand) || slot == int(proto.ItemSlot_ItemSlotOffHand) {
				weapons.Items[slot] = &proto.ItemSpec{Id: item.Id}
			} else {
				weapons.Items[slot] = &proto.ItemSpec{}
			}
		}

		for _, level := range []int32{10, 20, 30, 40} {
			player := &proto.Player{
				Name:      s.name,
				Race:      proto.Race_RaceOrc,
				Class:     proto.Class_ClassWarrior,
				Level:     level,
				Equipment: weapons,
				Rotation:  core.GetAplRotation(s.aplDir, s.rotation).Rotation,
			}
			switch spec := s.spec.(type) {
			case *proto.Player_DpsWarrior:
				player.Spec = spec
			case *proto.Player_ProtectionWarrior:
				player.Spec = spec
			}
			request := &proto.RaidSimRequest{
				Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
				Encounter:  core.MakeSingleTargetEncounter(0),
				SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 5},
			}

			result := core.RunRaidSim(request)
			if result.Error != nil {
				t.Fatalf("%s level %d: the sim failed: %s", s.name, level, result.Error.Message)
			}
			if dps := result.RaidMetrics.Dps.Avg; dps <= 0 {
				t.Errorf("%s level %d: the warrior deals %.0f dps, want some", s.name, level, dps)
			}

			sim := core.NewSim(request, simsignals.CreateSignals())
			unit := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
			for _, spell := range unit.Spellbook {
				id := spell.ActionID.SpellID
				if id == 0 || spell.ActionID.ItemID != 0 {
					continue
				}
				row := spelldata.Find(id)
				if row == spelldata.Nil || row.ClassFlags.Family != warriorFamily {
					continue
				}
				if taught := spelldata.TaughtAt(row); taught > level {
					t.Errorf("%s level %d: %s (%d) is registered, taught at %d", s.name, level, row.RankedName(), id, taught)
				}
			}
		}
	}
}
