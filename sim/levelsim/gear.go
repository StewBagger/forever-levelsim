package levelsim

import (
	"fmt"
	"regexp"
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

// The specs that tank. Their EP is the run's DPS EP, which values neither armor nor a shield, so their
// set keeps to what a tank wears: the best armor type the class has learned, and a one-handed weapon
// and a shield for a class that has shields.
var tankSpecs = map[proto.Spec]bool{
	proto.Spec_SpecProtectionWarrior: true,
	proto.Spec_SpecProtectionPaladin: true,
	proto.Spec_SpecFeralBearDruid:    true,
}

// Names the database carries for items that are not in the game: "DNT HEAD", "Template Item".
var placeholderName = regexp.MustCompile(`\b(DNT|PH|Test|TEST|Deprecated|DEPRECATED|Template)\b`)

// The lowest required level of an epic with a source: the level 40 bracers the database sells for
// reputation. A source-less epic below it is a placeholder copy of a raid item.
const minEpicRequiredLevel = 40

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
	tank       bool
	candidates []gearCandidate
}

// The weapons a set is picked with.
type hands int

const (
	handsByEP hands = iota // whichever of a two-hander or a main and off hand the EP values more
	handsTwo               // a two-hander
	handsOne               // a one-handed main hand and an off hand
)

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
	p.tank = tankSpecs[p.spec]

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
	if ui.Quality >= proto.ItemQuality_ItemQualityLegendary || placeholderName.MatchString(ui.Name) {
		return false
	}
	if len(ui.Sources) == 0 && !plausibleWithoutSource(ui) {
		return false
	}
	// An item with no required level is judged by its item level, so an unrestricted item made for
	// level 60 does not reach a level 20 set. So is a source-less item whose item level sits more than
	// maxItemLevelAboveLevel above its required level: the row's required level is not the one the
	// item is had at (Elderwild Construction Hammer: required level 1, item level 60).
	judgedByItemLevel := ui.ReqLvl == 0 || (len(ui.Sources) == 0 && itemLevel(ui) > ui.ReqLvl+maxItemLevelAboveLevel)
	if ui.ReqLvl > p.level || (judgedByItemLevel && itemLevel(ui) > p.level) {
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

// Whether an item the database holds no source for can be a real item. Of the 892 equippable
// source-less rows, most are world drops and quest rewards the database lost the source of (Ogremind
// Ring, Hammer of the Northern Wind), at an item level 5 above their required level like any other.
// The rest are not items a character gets:
//   - 76 at item level 1 or none: the "Restored" scaling set, the "Hidden" transmog slots, the "Gilded
//     Blademaster's" set and costumes, whose stats are a placeholder's (Restored Mail Shoulderguards: 29
//     armor and nothing else).
//   - 7 epics at required level 1, item level 10 with Molten Core names (Earthshaker, Band of Sulfuras):
//     no epic with a source requires less than minEpicRequiredLevel.
//
// A source-less item with a required level far below its item level is kept, but judged by its item
// level (allowed).
func plausibleWithoutSource(ui *proto.UIItem) bool {
	if itemLevel(ui) <= 1 {
		return false
	}
	return ui.Quality < proto.ItemQuality_ItemQualityEpic || ui.ReqLvl >= minEpicRequiredLevel
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

// The best armor type the class has learned at the level: plate for a warrior from 40, mail before.
func (p *GearPicker) bestArmorType() proto.ArmorType {
	for _, armorType := range core.ClassArmorTypeCapabilities[p.class] {
		if learnedAt, ok := armorTypeLevels[p.class][armorType]; !ok || p.level >= learnedAt {
			return armorType
		}
	}
	return proto.ArmorType_ArmorTypeUnknown
}

// Whether the run tanks with a shield: a tank of a class that has shields.
func (p *GearPicker) shieldTank() bool {
	_, hasShields := core.ClassWeaponTypeCapabilities[p.class][proto.WeaponType_WeaponTypeShield]
	return p.tank && hasShields
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

// The best set for the EP: per slot the highest-valued item, and whichever of a two-hander or a main
// and off hand is worth more.
func (p *GearPicker) Pick(ep *proto.UnitStats) *proto.EquipmentSpec {
	return p.pick(ep, handsByEP)
}

// The best set for the EP with the weapons given. A tank wears the best armor type it has learned
// in every armor slot that has one, and a shield tank a one-handed weapon and a shield whatever the
// weapons asked.
func (p *GearPicker) pick(ep *proto.UnitStats, weapons hands) *proto.EquipmentSpec {
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
	for _, cands := range bySlot {
		sort.Slice(cands, func(i, j int) bool { return cands[i].ui.Id < cands[j].ui.Id })
	}
	// The highest-valued item the filter keeps, of positive value. Filling, the slot takes the best
	// there is even at no value, the higher item level first, so an item level 1 piece loses to a
	// real one.
	best := func(slot proto.ItemSlot, filter func(*core.Item) bool, fill bool) (gearCandidate, float64) {
		var pick gearCandidate
		top := 0.0
		for _, c := range bySlot[slot] {
			if filter != nil && !filter(c.item) {
				continue
			}
			v := itemEP(c.item, ep, slot)
			switch {
			case pick.ui == nil:
				if v <= 0 && !fill {
					continue
				}
			case v < top, v == top && (!fill || itemLevel(c.ui) <= itemLevel(pick.ui)):
				continue
			}
			pick, top = c, v
		}
		return pick, top
	}
	armorType := p.bestArmorType()
	wearable := func(item *core.Item) bool {
		return !p.tank || item.ArmorType == proto.ArmorType_ArmorTypeUnknown || item.Type == proto.ItemType_ItemTypeBack || item.ArmorType == armorType
	}
	ofArmorType := func(item *core.Item) bool {
		return item.Type != proto.ItemType_ItemTypeBack && item.ArmorType != proto.ArmorType_ArmorTypeUnknown && item.ArmorType == armorType
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
		case proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotTrinket1:
			// A ring or trinket the EP gives no value stays out: below 40 there is often one useful
			// trinket or none, and the batch sim groups one item with an empty slot.
			c, _ := best(slot, nil, false)
			set(slot, c)
		case proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotTrinket2:
			// A trinket is never doubled, since nearly every one is unique-equipped in game whatever
			// the row says.
			c, _ := best(slot, pairedWith(items[slot-1], slot == proto.ItemSlot_ItemSlotTrinket2), false)
			set(slot, c)
		default:
			c, _ := best(slot, wearable, false)
			if c.ui == nil && p.tank {
				// A piece of the armor type the EP gives no value still beats one of another type or
				// none; with none of the type for the slot, the slot takes the best of any.
				if c, _ = best(slot, ofArmorType, true); c.ui == nil {
					c, _ = best(slot, nil, false)
				}
			}
			set(slot, c)
		}
	}

	shield := p.shieldTank()
	if shield {
		weapons = handsOne
	}
	twoHander, twoHandValue := best(proto.ItemSlot_ItemSlotMainHand, func(i *core.Item) bool { return i.HandType == proto.HandType_HandTypeTwoHand }, false)
	mainHand, mainValue := best(proto.ItemSlot_ItemSlotMainHand, func(i *core.Item) bool { return i.HandType != proto.HandType_HandTypeTwoHand }, false)
	var mainSpec *proto.ItemSpec
	if mainHand.ui != nil {
		mainSpec = &proto.ItemSpec{Id: mainHand.ui.Id}
	}
	offFilter := pairedWith(mainSpec, false)
	if shield {
		paired := offFilter
		offFilter = func(i *core.Item) bool {
			return i.WeaponType == proto.WeaponType_WeaponTypeShield && (paired == nil || paired(i))
		}
	}
	offHand, offValue := best(proto.ItemSlot_ItemSlotOffHand, offFilter, shield)
	useTwoHander := twoHandValue > mainValue+offValue
	switch weapons {
	case handsTwo:
		useTwoHander = twoHander.ui != nil
	case handsOne:
		useTwoHander = mainHand.ui == nil
	}
	if useTwoHander {
		set(proto.ItemSlot_ItemSlotMainHand, twoHander)
	} else {
		set(proto.ItemSlot_ItemSlotMainHand, mainHand)
		set(proto.ItemSlot_ItemSlotOffHand, offHand)
	}
	// Every character holds a weapon where one can be equipped: with none the EP values, or none of
	// the kind the weapons asked, the main hand takes the best there is, the higher item level first.
	// A run with an empty hand weighs its weapon damage at nothing, so the next round would leave
	// the hand empty too.
	if items[proto.ItemSlot_ItemSlotMainHand] == nil {
		oneHanded := func(i *core.Item) bool { return i.HandType != proto.HandType_HandTypeTwoHand }
		if shield {
			if c, _ := best(proto.ItemSlot_ItemSlotMainHand, oneHanded, true); c.ui != nil {
				set(proto.ItemSlot_ItemSlotMainHand, c)
				set(proto.ItemSlot_ItemSlotOffHand, offHand)
			}
		}
		if items[proto.ItemSlot_ItemSlotMainHand] == nil {
			c, _ := best(proto.ItemSlot_ItemSlotMainHand, nil, true)
			set(proto.ItemSlot_ItemSlotMainHand, c)
		}
	}

	for i := range items {
		if items[i] == nil {
			items[i] = &proto.ItemSpec{}
		}
	}
	return &proto.EquipmentSpec{Items: items}
}

// What the other slot of a pair may hold beside the item worn in one: not a second copy of a unique
// item, nor another item of its limit category (the two Runes of Perfection share one, and the game
// equips one at a time).
func pairedWith(worn *proto.ItemSpec, unique bool) func(*core.Item) bool {
	if worn.GetId() == 0 {
		return nil
	}
	first := core.GetItemByID(worn.Id)
	return func(item *core.Item) bool {
		if item.ID == first.ID && (unique || first.Unique) {
			return false
		}
		return first.LimitCategory == 0 || item.LimitCategory != first.LimitCategory
	}
}

// The weapons a set wears: a two-hander, a main hand with or without an off hand, or none.
func handsWorn(e *proto.EquipmentSpec) hands {
	if e == nil || len(e.Items) <= int(proto.ItemSlot_ItemSlotMainHand) {
		return handsByEP
	}
	mainHand := core.GetItemByID(e.Items[proto.ItemSlot_ItemSlotMainHand].GetId())
	switch {
	case mainHand == nil:
		return handsByEP
	case mainHand.HandType == proto.HandType_HandTypeTwoHand:
		return handsTwo
	}
	return handsOne
}

// Whether a set holds a weapon in each hand.
func dualWields(e *proto.EquipmentSpec) bool {
	if e == nil || len(e.Items) <= int(proto.ItemSlot_ItemSlotOffHand) {
		return false
	}
	offHand := core.GetItemByID(e.Items[proto.ItemSlot_ItemSlotOffHand].GetId())
	return offHand != nil && offHand.Type == proto.ItemType_ItemTypeWeapon &&
		offHand.WeaponType != proto.WeaponType_WeaponTypeShield && offHand.WeaponType != proto.WeaponType_WeaponTypeOffHand
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
