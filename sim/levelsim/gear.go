package levelsim

import (
	"fmt"
	"slices"
	"sort"
	"sync"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The gear a levelling run wears is the best in slot for its level with no raid gear (owner's rule,
// 2026-11-03): every item a character of that level can equip, from any source but a raid, valued by
// the run's own EP at that level. The set and the EP depend on each other, so the run alternates the
// two until the set stops changing.
//
// What the EP cannot see is not valued: an item's procs and on-use effects, set bonuses and enchants.
// Gear carries no enchants.

// The levels a class learns an armor type or dual wield at, below which it does not have it.
var armorTypeLevels = map[proto.Class]map[proto.ArmorType]int32{
	proto.Class_ClassWarrior: {proto.ArmorType_ArmorTypePlate: 40},
	proto.Class_ClassPaladin: {proto.ArmorType_ArmorTypePlate: 40},
	proto.Class_ClassHunter:  {proto.ArmorType_ArmorTypeMail: 40},
	proto.Class_ClassShaman:  {proto.ArmorType_ArmorTypeMail: 40},
}

var dualWieldLevels = map[proto.Class]int32{
	proto.Class_ClassWarrior: 20,
	proto.Class_ClassHunter:  20,
	proto.Class_ClassRogue:   10,
}

// How far an item's level may sit above the character's: a dungeon blue for its level is 5-10 above.
const maxItemLevelAboveLevel = 10

var hordeRaces = []proto.Race{proto.Race_RaceOrc, proto.Race_RaceTauren, proto.Race_RaceTroll, proto.Race_RaceUndead, proto.Race_RaceBloodElf}
var allianceRaces = []proto.Race{proto.Race_RaceHuman, proto.Race_RaceDwarf, proto.Race_RaceGnome, proto.Race_RaceNightElf}

type gearCandidate struct {
	ui   *proto.UIItem
	item *core.Item
}

type GearPicker struct {
	level      int32
	class      proto.Class
	spec       proto.Spec
	race       proto.Race
	profs      []proto.Profession
	candidates []gearCandidate
}

var (
	uiDatabase     *proto.UIDatabase
	uiDatabaseOnce sync.Once
)

// The web server and the wasm sim are built without the with_db tag, so their sim database holds only
// what requests carried; the picker needs every item, so it adds them on first use.
func loadDatabase() {
	uiDatabase = database.Load()
	if !core.WITH_DB {
		core.AddToDatabase(core.SimDatabaseFromUI(uiDatabase))
	}
}

func NewGearPicker(player *proto.Player, level int32) *GearPicker {
	uiDatabaseOnce.Do(loadDatabase)
	p := &GearPicker{
		level: level,
		class: player.Class,
		spec:  core.PlayerProtoToSpec(player),
		race:  player.Race,
		profs: []proto.Profession{player.Profession1, player.Profession2},
	}

	// A raid is any zone a raid-difficulty drop comes from.
	raidZones := map[int32]bool{}
	for _, ui := range uiDatabase.Items {
		for _, src := range ui.Sources {
			if d := src.GetDrop(); d != nil && d.Difficulty == proto.DungeonDifficulty_DifficultyRaid && d.ZoneId != 0 {
				raidZones[d.ZoneId] = true
			}
		}
	}

	for _, ui := range uiDatabase.Items {
		if !p.allowed(ui, raidZones) {
			continue
		}
		if core.GetItemByID(ui.Id) == nil {
			continue
		}
		// The item as worn: NewItem scales its stats to its item level.
		item := core.NewItem(core.ItemSpec{ID: ui.Id})
		if !p.canEquip(&item) {
			continue
		}
		p.candidates = append(p.candidates, gearCandidate{ui: ui, item: &item})
	}
	return p
}

func (p *GearPicker) allowed(ui *proto.UIItem, raidZones map[int32]bool) bool {
	if ui.Quality >= proto.ItemQuality_ItemQualityLegendary {
		return false
	}
	// An item with no required level is judged by its item level, so an unrestricted item made for
	// level 60 does not reach a level 20 set.
	if ui.ReqLvl > p.level || (ui.ReqLvl == 0 && itemLevel(ui) > p.level) {
		return false
	}
	// An item far above the level it can be equipped at is one whose stats the database holds at a
	// fixed item level the character would not get (Youngster's First Bow: required level 1, item
	// level 54); its stats are not the ones a character of this level wears.
	if itemLevel(ui) > p.level+maxItemLevelAboveLevel {
		return false
	}
	for _, src := range ui.Sources {
		if d := src.GetDrop(); d != nil && (d.Difficulty == proto.DungeonDifficulty_DifficultyRaid || raidZones[d.ZoneId]) {
			return false
		}
	}
	if len(ui.ClassAllowlist) > 0 && !slices.Contains(ui.ClassAllowlist, p.class) {
		return false
	}
	if ui.RequiredProfession != proto.Profession_ProfessionUnknown && !slices.Contains(p.profs, ui.RequiredProfession) {
		return false
	}
	faction := p.faction()
	switch ui.FactionRestriction {
	case proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY:
		return faction != proto.Faction_Horde
	case proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY:
		return faction != proto.Faction_Alliance
	}
	// The item row often carries no restriction where its sources do: a battleground quartermaster or
	// reputation sells to one side. An item every source of which is the other side's is not obtainable.
	if faction == proto.Faction_Unknown || len(ui.Sources) == 0 {
		return true
	}
	for _, src := range ui.Sources {
		side := proto.Faction_Unknown
		switch {
		case src.GetRep() != nil:
			side = src.GetRep().FactionId
		case src.GetSoldBy() != nil:
			side = src.GetSoldBy().Faction
		case src.GetQuest() != nil:
			side = src.GetQuest().Faction
		case src.GetPvp() != nil:
			continue
		}
		if side == proto.Faction_Unknown || side == faction {
			return true
		}
	}
	return false
}

// The player's side, or Unknown for a race this table does not place.
func (p *GearPicker) faction() proto.Faction {
	switch {
	case slices.Contains(hordeRaces, p.race):
		return proto.Faction_Horde
	case slices.Contains(allianceRaces, p.race):
		return proto.Faction_Alliance
	}
	return proto.Faction_Unknown
}

func itemLevel(ui *proto.UIItem) int32 {
	if opt, ok := ui.ScalingOptions[0]; ok {
		return opt.Ilvl
	}
	return 0
}

func (p *GearPicker) canEquip(item *core.Item) bool {
	switch item.Type {
	case proto.ItemType_ItemTypeFinger, proto.ItemType_ItemTypeTrinket, proto.ItemType_ItemTypeNeck, proto.ItemType_ItemTypeBack:
		return true
	case proto.ItemType_ItemTypeWeapon:
		eligible, ok := core.ClassWeaponTypeCapabilities[p.class][item.WeaponType]
		if !ok {
			return false
		}
		return item.HandType != proto.HandType_HandTypeTwoHand || eligible.CanUseTwoHand
	case proto.ItemType_ItemTypeRanged:
		return slices.Contains(core.ClassRangedWeaponTypeCapabilities[p.class], item.RangedWeaponType)
	}
	if learnedAt, ok := armorTypeLevels[p.class][item.ArmorType]; ok && p.level < learnedAt {
		return false
	}
	armorTypes := core.ClassArmorTypeCapabilities[p.class]
	return len(armorTypes) > 0 && item.ArmorType <= armorTypes[0]
}

func (p *GearPicker) canDualWield() bool {
	if !core.SpecCanDualWieldCapabilities[p.spec] {
		return false
	}
	learnedAt, ok := dualWieldLevels[p.class]
	return !ok || p.level >= learnedAt
}

// The value of an item to the run, by its EP.
func itemEP(item *core.Item, ep *proto.UnitStats, slot proto.ItemSlot) float64 {
	value := 0.0
	for s := 0; s < int(stats.ProtoStatsLen); s++ {
		if v := item.Stats[s]; v != 0 {
			value += v * at(ep.GetStats(), s)
		}
	}
	if item.WeaponDamageMax > 0 && item.SwingSpeed > 0 {
		dps := (item.WeaponDamageMin + item.WeaponDamageMax) / 2 / item.SwingSpeed
		switch {
		case item.Type == proto.ItemType_ItemTypeRanged:
			value += dps * at(ep.GetPseudoStats(), int(proto.PseudoStat_PseudoStatRangedDps))
		case slot == proto.ItemSlot_ItemSlotOffHand:
			value += dps * at(ep.GetPseudoStats(), int(proto.PseudoStat_PseudoStatOffHandDps))
		default:
			value += dps * at(ep.GetPseudoStats(), int(proto.PseudoStat_PseudoStatMainHandDps))
		}
	}
	return value
}

// The best set for the EP: per slot the highest-valued item, two different ones in the paired slots
// where an item is unique, and whichever of a two-hander or a main and off hand is worth more.
func (p *GearPicker) Pick(ep *proto.UnitStats) *proto.EquipmentSpec {
	bySlot := map[proto.ItemSlot][]gearCandidate{}
	for _, c := range p.candidates {
		for _, slot := range core.EligibleSlotsForItem(c.item) {
			if slot == proto.ItemSlot_ItemSlotOffHand && c.item.Type == proto.ItemType_ItemTypeWeapon &&
				c.item.WeaponType != proto.WeaponType_WeaponTypeShield && c.item.WeaponType != proto.WeaponType_WeaponTypeOffHand && !p.canDualWield() {
				continue
			}
			bySlot[slot] = append(bySlot[slot], c)
		}
	}
	best := func(slot proto.ItemSlot, exclude int32, filter func(*core.Item) bool) (gearCandidate, float64) {
		var pick gearCandidate
		top := 0.0
		cands := bySlot[slot]
		sort.Slice(cands, func(i, j int) bool { return cands[i].ui.Id < cands[j].ui.Id })
		for _, c := range cands {
			if c.ui.Id == exclude || (filter != nil && !filter(c.item)) {
				continue
			}
			if v := itemEP(c.item, ep, slot); v > top {
				pick, top = c, v
			}
		}
		return pick, top
	}

	items := make([]*proto.ItemSpec, len(proto.ItemSlot_name))
	set := func(slot proto.ItemSlot, c gearCandidate) {
		if c.ui != nil {
			items[slot] = &proto.ItemSpec{Id: c.ui.Id}
		}
	}
	for slot := proto.ItemSlot_ItemSlotHead; slot <= proto.ItemSlot_ItemSlotRanged; slot++ {
		switch slot {
		case proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand:
			continue
		case proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotTrinket2:
			// A second copy is taken only of a ring the database does not mark unique; a trinket
			// is never doubled, since nearly every one is unique-equipped in game whatever the row says.
			first := items[slot-1]
			exclude := int32(0)
			if first != nil && (slot == proto.ItemSlot_ItemSlotTrinket2 || core.GetItemByID(first.Id).Unique) {
				exclude = first.Id
			}
			c, _ := best(slot, exclude, nil)
			set(slot, c)
		default:
			c, _ := best(slot, 0, nil)
			set(slot, c)
		}
	}

	twoHander, twoHandValue := best(proto.ItemSlot_ItemSlotMainHand, 0, func(i *core.Item) bool { return i.HandType == proto.HandType_HandTypeTwoHand })
	mainHand, mainValue := best(proto.ItemSlot_ItemSlotMainHand, 0, func(i *core.Item) bool { return i.HandType != proto.HandType_HandTypeTwoHand })
	exclude := int32(0)
	if mainHand.item != nil && mainHand.item.Unique {
		exclude = mainHand.ui.Id
	}
	offHand, offValue := best(proto.ItemSlot_ItemSlotOffHand, exclude, nil)
	if twoHandValue > mainValue+offValue {
		set(proto.ItemSlot_ItemSlotMainHand, twoHander)
	} else {
		set(proto.ItemSlot_ItemSlotMainHand, mainHand)
		set(proto.ItemSlot_ItemSlotOffHand, offHand)
	}

	for i := range items {
		if items[i] == nil {
			items[i] = &proto.ItemSpec{}
		}
	}
	return &proto.EquipmentSpec{Items: items}
}

func at(values []float64, i int) float64 {
	if i < len(values) {
		return values[i]
	}
	return 0
}

func SameGear(a, b *proto.EquipmentSpec) bool {
	if a == nil || b == nil || len(a.Items) != len(b.Items) {
		return false
	}
	for i := range a.Items {
		if a.Items[i].GetId() != b.Items[i].GetId() {
			return false
		}
	}
	return true
}

func GearNames(e *proto.EquipmentSpec) []string {
	var names []string
	for slot, spec := range e.Items {
		if spec.GetId() == 0 {
			continue
		}
		name := fmt.Sprint(spec.Id)
		if item := core.GetItemByID(spec.Id); item != nil {
			name = item.Name
		}
		names = append(names, fmt.Sprintf("%s: %s (%d)", proto.ItemSlot(slot).String()[len("ItemSlot"):], name, spec.Id))
	}
	return names
}
