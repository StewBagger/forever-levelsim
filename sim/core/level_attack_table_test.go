package core

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// The level gaps the tables below are checked at, target level minus character level.
var levelTableGaps = []int32{-3, -2, -1, 0, 1, 2, 3}

const (
	levelTableArmor          = 1500
	levelTableFireResistance = 75
)

// The weapon tester at a level, against one target per level gap and a last target with no level, which
// takes the default raid boss level. offHand 0 leaves it with one weapon.
func levelTableRaidSimRequest(level int32, offHand int32) *proto.RaidSimRequest {
	ensureWeaponTester()
	targetStats := stats.Stats{stats.Armor: levelTableArmor, stats.FireResistance: levelTableFireResistance}.ToProtoArray()
	var targets []*proto.Target
	for _, gap := range levelTableGaps {
		targets = append(targets, &proto.Target{Level: level + gap, Stats: targetStats, MobType: proto.MobType_MobTypeHumanoid})
	}
	targets = append(targets, &proto.Target{Stats: targetStats, MobType: proto.MobType_MobTypeHumanoid})

	player := &proto.Player{
		Name:      "Level Tester",
		Class:     proto.Class_ClassRogue,
		Race:      proto.Race_RaceUndead,
		Level:     level,
		Buffs:     &proto.IndividualBuffs{},
		Spec:      &proto.Player_Rogue{},
		Equipment: &proto.EquipmentSpec{Items: WeaponTestSlots(WeaponTestSwordID, offHand, 0)},
		Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: targets, Duration: 180},
	}
}

func levelTableSim(t *testing.T, level int32, offHand int32) (*Simulation, *Character) {
	t.Helper()
	sim := NewSim(levelTableRaidSimRequest(level, offHand), simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// The Classic (1.12) player-against-mob table at a defense minus weapon skill of d, with the weapon skill
// at its level-based maximum of 5 a level and a mob's defense at 5 a level, so d is 5 per level of gap.
type classicMeleeTable struct {
	miss, dodge, parry, glance, glanceMin, glanceMax, critSuppression float64
}

func classicMeleeTableAt(d float64) classicMeleeTable {
	table := classicMeleeTable{
		// Dodge: 5% + 0.1% a point either way.
		dodge: 0.05 + 0.001*d,
		// Glancing blows: 10% + 2% a point of defense over the capped skill, never below 0.
		glance: math.Max(0, 0.1+0.02*d),
		// Glancing damage: a roll between these, low 1.3 - 0.05d within [0.01, 0.91], high 1.2 - 0.03d
		// within [0.2, 0.99].
		glanceMin: Clamp(1.3-0.05*d, 0.01, 0.91),
		glanceMax: Clamp(1.2-0.03*d, 0.2, 0.99),
	}
	// Miss: 5% + 0.1% a point, 0.2% a point past a gap of 10.
	if d > 10 {
		table.miss = 0.05 + 0.002*d
	} else {
		table.miss = 0.05 + 0.001*d
	}
	// Parry: 5% + 0.1% a point, 0.6% a point past a gap of 10.
	if d > 10 {
		table.parry = 0.05 + 0.006*d
	} else {
		table.parry = 0.05 + 0.001*d
	}
	// Crit: 0.2% less a point of defense over skill, 0.04% more a point of skill over defense.
	if d > 0 {
		table.critSuppression = 0.002 * d
	} else {
		table.critSuppression = 0.0004 * d
	}
	return table
}

// Classic spell hit against a mob: 96% at equal level, a point less for each of the first two levels above
// and 11 points less at the third, a point more for each level below. Forever has no 1% miss floor.
func classicSpellMissAt(gap int32) float64 {
	if gap >= 3 {
		return 0.17
	}
	return math.Max(0, 0.04+0.01*float64(gap))
}

// The attack tables of a character below 60 follow the level gap alone, as at 60: the weapon skill is 5 a
// character level and a target's defense 5 a target level, so each gap gives the Classic table whatever the
// character's level. Armor and partial resists are the level-dependent parts, through the attacker's level.
func TestAttackTablesBelowLevel60FollowTheLevelGap(t *testing.T) {
	for _, level := range []int32{10, 20, 30, 40, 50, 60} {
		sim, character := levelTableSim(t, level, 0)
		_, dualWielder := levelTableSim(t, level, WeaponTestDaggerID)

		white := &Spell{Unit: &character.Unit, ProcMask: ProcMaskMeleeMHAuto}
		special := &Spell{Unit: &character.Unit, ProcMask: ProcMaskMeleeMHSpecial}
		dualWieldWhite := &Spell{Unit: &dualWielder.Unit, ProcMask: ProcMaskMeleeMHAuto}
		dualWieldSpecial := &Spell{Unit: &dualWielder.Unit, ProcMask: ProcMaskMeleeMHSpecial}
		auraCritWhite := &Spell{Unit: &character.Unit, ProcMask: ProcMaskMeleeMHAuto, BonusCritPercent: 5}
		fire := &Spell{Unit: &character.Unit, SpellSchool: SpellSchoolFire, SchoolIndex: stats.SchoolIndexFire}
		physical := &Spell{Unit: &character.Unit, SpellSchool: SpellSchoolPhysical, SchoolIndex: stats.SchoolIndexPhysical}

		targets := sim.Encounter.AllTargetUnits
		equalLevel := character.AttackTables[targets[3].UnitIndex]
		critAtEqualLevel := white.physicalCritChanceUnclamped(equalLevel, 0)
		auraCritAtEqualLevel := auraCritWhite.physicalCritChanceUnclamped(equalLevel, 0)

		for i, target := range targets {
			gap := target.Level - level
			if i == len(levelTableGaps) {
				if gap != 3 {
					t.Errorf("level %d: a target with no level is %+d, want the default raid boss at +3", level, gap)
				}
			} else if gap != levelTableGaps[i] {
				t.Fatalf("level %d: target %d is %+d, want %+d", level, i, gap, levelTableGaps[i])
			}

			table := character.AttackTables[target.UnitIndex]
			dualWieldTable := dualWielder.AttackTables[target.UnitIndex]
			want := classicMeleeTableAt(5 * float64(gap))
			auraCritSuppression := 0.0
			if gap >= 3 {
				auraCritSuppression = 0.018
			}
			armor := float64(levelTableArmor)
			resistance := float64(levelTableFireResistance)

			checkAttackTableValues(t, fmt.Sprintf("level %d against %+d", level, gap), []attackTableValue{
				{"white miss", white.GetPhysicalMissChance(table), want.miss},
				{"dual-wield white miss", dualWieldWhite.GetPhysicalMissChance(dualWieldTable), want.miss + 0.19},
				{"special miss", special.physicalMissChanceNoDWPenalty(table), want.miss},
				{"dual-wield special miss", dualWieldSpecial.physicalMissChanceNoDWPenalty(dualWieldTable), want.miss},
				{"dodge", target.GetTotalDodgeChanceAsDefender(white, table), want.dodge},
				{"parry", target.GetTotalParryChanceAsDefender(white, table), want.parry},
				{"glance chance", table.Hand(white).BaseGlanceChance, want.glance},
				{"glance multiplier min", table.Hand(white).GlanceMultiplierMin, want.glanceMin},
				{"glance multiplier max", table.Hand(white).GlanceMultiplierMax, want.glanceMax},
				{"crit lost to the gap", critAtEqualLevel - white.physicalCritChanceUnclamped(table, 0), want.critSuppression},
				{"crit with aura crit lost to the gap", auraCritAtEqualLevel - auraCritWhite.physicalCritChanceUnclamped(table, 0), want.critSuppression + auraCritSuppression},
				{"spell miss", fire.SpellChanceToMiss(table), classicSpellMissAt(gap)},
				{"spell crit lost to the gap", fire.SpellCritChance(targets[3]) - fire.SpellCritChance(target), 0},
				{"armor damage taken", table.GetArmorDamageModifier(physical), 1 - armor/(armor+400+85*float64(level))},
				{"average fire resist", target.averageResist(fire, &character.Unit, false), resistance / (resistance + 9*float64(level))},
			})
		}
	}
}
