package levelsim

import (
	"os"
	"testing"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The picker reads a player's spec through the registered agents, as the sims that call it do.
func TestMain(m *testing.M) {
	sim.RegisterAll()
	os.Exit(m.Run())
}

// A level 30 orc hunter's set holds only items a level 30 Horde hunter can equip and get outside a
// raid: no required level above 30, no mail before 40, nothing sold only to the Alliance.
func TestBisGearKeepsToTheLevel(t *testing.T) {
	const level = 30
	byID := map[int32]*proto.UIItem{}
	for _, ui := range database.Load().Items {
		byID[ui.Id] = ui
	}
	p := NewGearPicker(&proto.Player{Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc, Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{}}}, level)
	ep := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
	ep.Stats[proto.Stat_StatAgility], ep.Stats[proto.Stat_StatStamina] = 1, 1
	ep.PseudoStats[proto.PseudoStat_PseudoStatRangedDps], ep.PseudoStats[proto.PseudoStat_PseudoStatMainHandDps] = 10, 1
	set := p.Pick(ep)

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
		t.Errorf("only %d slots filled: %v", worn, GearNames(set))
	}
	if set.Items[proto.ItemSlot_ItemSlotTrinket1].GetId() == set.Items[proto.ItemSlot_ItemSlotTrinket2].GetId() {
		t.Errorf("the same trinket twice: %v", GearNames(set))
	}
}

func uiItemsByID() map[int32]*proto.UIItem {
	byID := map[int32]*proto.UIItem{}
	for _, ui := range database.Load().Items {
		byID[ui.Id] = ui
	}
	return byID
}

func flatEP(stats []proto.Stat, pseudoStats []proto.PseudoStat) *proto.UnitStats {
	ep := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
	for _, s := range stats {
		ep.Stats[s] = 1
	}
	for _, s := range pseudoStats {
		ep.PseudoStats[s] = 10
	}
	return ep
}

// Placeholder rows the database holds with no source are not obtainable at any level below 60; the
// world drops and quest rewards it lost the source of are.
func TestSourcelessPlaceholdersAreNotObtainable(t *testing.T) {
	byID := uiItemsByID()
	placeholders := []int32{
		276719,                                         // DNT HEAD
		281702,                                         // Restored Mail Shoulderguards, item level 1
		288900, 288926, 288927, 288931, 288929, 288928, // Molten Core names at required level 1, item level 10
		281586, 281587, 281588, // required level 1, item level 60
	}
	real := map[int32]int32{
		1993: 31, // Ogremind Ring
		810:  49, // Hammer of the Northern Wind, a world drop epic
	}
	for _, level := range []int32{20, 40, 50, 59} {
		p := NewGearPicker(&proto.Player{Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc, Spec: &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}}}, level)
		for _, id := range placeholders {
			if ui := byID[id]; ui == nil {
				t.Errorf("%d is not in the database", id)
			} else if p.allowed(ui, map[int32]bool{}) {
				t.Errorf("level %d: %s (%d) is obtainable", level, ui.Name, id)
			}
		}
		for id, reqLvl := range real {
			if ui := byID[id]; level >= reqLvl && !p.allowed(ui, map[int32]bool{}) {
				t.Errorf("level %d: %s (%d) is not obtainable", level, ui.Name, id)
			}
		}
	}
}

// The second ring and trinket keep out the first's limit category.
func TestPairedSlotsKeepToTheLimitCategory(t *testing.T) {
	for _, level := range []int32{30, 40, 50} {
		p := NewGearPicker(&proto.Player{Class: proto.Class_ClassMage, Race: proto.Race_RaceUndead, Spec: &proto.Player_Mage{Mage: &proto.Mage{}}}, level)
		set := p.Pick(flatEP([]proto.Stat{proto.Stat_StatIntellect, proto.Stat_StatSpellDamage, proto.Stat_StatHitRating, proto.Stat_StatCritRating}, nil))
		for _, pair := range [][2]proto.ItemSlot{{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}, {proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}} {
			a, b := core.GetItemByID(set.Items[pair[0]].GetId()), core.GetItemByID(set.Items[pair[1]].GetId())
			if a == nil || b == nil {
				continue
			}
			if a.LimitCategory != 0 && a.LimitCategory == b.LimitCategory {
				t.Errorf("level %d: %s and %s share limit category %d", level, a.Name, b.Name, a.LimitCategory)
			}
		}
	}
}

// A protection warrior wears a one-handed weapon and a shield, and mail before level 40, plate from it.
func TestTankWearsAShieldAndItsArmor(t *testing.T) {
	ep := flatEP([]proto.Stat{proto.Stat_StatStrength, proto.Stat_StatAgility, proto.Stat_StatAttackPower, proto.Stat_StatHitRating, proto.Stat_StatCritRating},
		[]proto.PseudoStat{proto.PseudoStat_PseudoStatMainHandDps})
	for level, armorType := range map[int32]proto.ArmorType{30: proto.ArmorType_ArmorTypeMail, 50: proto.ArmorType_ArmorTypePlate} {
		p := NewGearPicker(&proto.Player{Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc, Spec: &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}}}, level)
		set := p.Pick(ep)
		mainHand, offHand := core.GetItemByID(set.Items[proto.ItemSlot_ItemSlotMainHand].GetId()), core.GetItemByID(set.Items[proto.ItemSlot_ItemSlotOffHand].GetId())
		if mainHand == nil || mainHand.HandType == proto.HandType_HandTypeTwoHand {
			t.Errorf("level %d: main hand is not one-handed: %v", level, GearNames(set))
		}
		if offHand == nil || offHand.WeaponType != proto.WeaponType_WeaponTypeShield {
			t.Errorf("level %d: off hand is not a shield: %v", level, GearNames(set))
		}
		for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist,
			proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet} {
			if item := core.GetItemByID(set.Items[slot].GetId()); item != nil && item.ArmorType != armorType {
				t.Errorf("level %d: %s: %s is %s", level, slot, item.Name, item.ArmorType)
			}
		}
	}
}

// BisGear keeps the weapons the player wears: a two-hander for one wearing a two-hander, a main and
// off hand for one dual wielding.
func TestBisGearKeepsTheWeaponsWorn(t *testing.T) {
	const level = 30
	ep := flatEP([]proto.Stat{proto.Stat_StatAgility, proto.Stat_StatAttackPower, proto.Stat_StatRangedAttackPower},
		[]proto.PseudoStat{proto.PseudoStat_PseudoStatRangedDps, proto.PseudoStat_PseudoStatMainHandDps, proto.PseudoStat_PseudoStatOffHandDps})
	player := &proto.Player{Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc, Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{}}}
	p := NewGearPicker(player, level)
	twoHand, oneHand := p.pick(ep, handsTwo), p.pick(ep, handsOne)
	if handsWorn(twoHand) != handsTwo || !dualWields(oneHand) {
		t.Fatalf("no two-hander or no dual wield pair at level %d: %v / %v", level, GearNames(twoHand), GearNames(oneHand))
	}
	for _, worn := range []*proto.EquipmentSpec{twoHand, oneHand} {
		player.Equipment = worn
		res := BisGear(&proto.BisGearRequest{Player: player, Level: level, Ep: ep})
		if res.Error != nil {
			t.Fatal(res.Error.Message)
		}
		if handsWorn(res.Equipment) != handsWorn(worn) {
			t.Errorf("wearing %v, BisGear gave %v", GearNames(worn), GearNames(res.Equipment))
		}
	}
}

// With an EP that values nothing an item holds, a level 10 warrior and paladin still hold a weapon,
// the higher item level first, and leave the rings, trinkets and armor the EP gives no value empty.
func TestEveryCharacterHoldsAWeapon(t *testing.T) {
	const level = 10
	byID := uiItemsByID()
	none := flatEP(nil, nil)
	for _, player := range []*proto.Player{
		{Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc, Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}}},
		{Class: proto.Class_ClassPaladin, Race: proto.Race_RaceHuman, Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{}}},
	} {
		p := NewGearPicker(player, level)
		set := p.Pick(none)
		mainHand := set.Items[proto.ItemSlot_ItemSlotMainHand].GetId()
		if mainHand == 0 {
			t.Errorf("%s: no main hand: %v", player.Class, GearNames(set))
			continue
		}
		for _, c := range p.candidates {
			if c.item.Type == proto.ItemType_ItemTypeWeapon && c.item.WeaponType != proto.WeaponType_WeaponTypeShield &&
				c.item.WeaponType != proto.WeaponType_WeaponTypeOffHand && itemLevel(c.ui) > itemLevel(byID[mainHand]) {
				t.Errorf("%s: %s (item level %d) over %s (item level %d)", player.Class, byID[mainHand].Name, itemLevel(byID[mainHand]), c.ui.Name, itemLevel(c.ui))
				break
			}
		}
		for slot, spec := range set.Items {
			if spec.GetId() != 0 && proto.ItemSlot(slot) != proto.ItemSlot_ItemSlotMainHand {
				t.Errorf("%s: %s holds %s, which the EP gives no value", player.Class, proto.ItemSlot(slot), byID[spec.Id].Name)
			}
		}
	}
}

// A level 10 protection warrior wears pieces of its armor type the EP gives no value, but no
// armor-slot item without one: the item level 1 Barrens Headband is not a helm.
func TestTankFillsOnlyWithItsArmorType(t *testing.T) {
	p := NewGearPicker(&proto.Player{Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc, Spec: &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}}}, 10)
	set := p.Pick(flatEP(nil, nil))
	worn := 0
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist,
		proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet} {
		if item := core.GetItemByID(set.Items[slot].GetId()); item != nil {
			worn++
			if item.ArmorType != proto.ArmorType_ArmorTypeMail {
				t.Errorf("%s: %s is %s", slot, item.Name, item.ArmorType)
			}
		}
	}
	if worn == 0 {
		t.Errorf("no mail worn: %v", GearNames(set))
	}
	if offHand := core.GetItemByID(set.Items[proto.ItemSlot_ItemSlotOffHand].GetId()); offHand == nil || offHand.WeaponType != proto.WeaponType_WeaponTypeShield {
		t.Errorf("off hand is not a shield: %v", GearNames(set))
	}
	if set.Items[proto.ItemSlot_ItemSlotMainHand].GetId() == 0 {
		t.Errorf("no main hand: %v", GearNames(set))
	}
}
