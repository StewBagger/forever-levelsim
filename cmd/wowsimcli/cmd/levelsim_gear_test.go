package cmd

import (
	"testing"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// A level 30 orc hunter's set holds only items a level 30 Horde hunter can equip and get outside a
// raid: no required level above 30, no mail before 40, nothing sold only to the Alliance.
func TestBisGearKeepsToTheLevel(t *testing.T) {
	const level = 30
	byID := map[int32]*proto.UIItem{}
	for _, ui := range database.Load().Items {
		byID[ui.Id] = ui
	}
	p := newGearPicker(&proto.Player{Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc, Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{}}}, level)
	set := p.pick(map[string]float64{"StatAgility": 1, "StatStamina": 1, "PseudoStatRangedDps": 10, "PseudoStatMainHandDps": 1})

	worn := 0
	for slot, spec := range set.Items {
		if spec.GetId() == 0 {
			continue
		}
		worn++
		ui, item := byID[spec.Id], core.GetItemByID(spec.Id)
		if ui.ReqLvl > level || itemLevel(ui) > level+maxItemLevelAboveLevel {
			t.Errorf("%s: %s requires %d at item level %d", proto.ItemSlot(slot), ui.Name, ui.ReqLvl, itemLevel(ui))
		}
		if item.ArmorType > proto.ArmorType_ArmorTypeLeather {
			t.Errorf("%s: %s is %s before level 40", proto.ItemSlot(slot), ui.Name, item.ArmorType)
		}
		if !p.allowed(ui, map[int32]bool{}) {
			t.Errorf("%s: %s is not obtainable for this character", proto.ItemSlot(slot), ui.Name)
		}
		for _, src := range ui.Sources {
			if src.GetDrop().GetDifficulty() == proto.DungeonDifficulty_DifficultyRaid {
				t.Errorf("%s: %s drops in a raid", proto.ItemSlot(slot), ui.Name)
			}
		}
	}
	if worn < 14 {
		t.Errorf("only %d slots filled: %v", worn, gearNames(set))
	}
	if set.Items[proto.ItemSlot_ItemSlotTrinket1].GetId() == set.Items[proto.ItemSlot_ItemSlotTrinket2].GetId() {
		t.Errorf("the same trinket twice: %v", gearNames(set))
	}
}
