package levelsim

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/druid"
	"github.com/wowsims/forever/sim/druid/balance"
	"github.com/wowsims/forever/sim/druid/feralbear"
	"github.com/wowsims/forever/sim/druid/feralcat"
	"github.com/wowsims/forever/sim/druid/restoration"
)

func init() {
	balance.RegisterBalanceDruid()
	feralbear.RegisterFeralBearDruid()
	feralcat.RegisterFeralCatDruid()
	restoration.RegisterRestorationDruid()
	common.RegisterAllEffects()
}

type levelSpec struct {
	name     string
	spec     interface{}
	gear     string
	rotation string
	form     druid.DruidForm // The form the spec fights in, which it needs to have learned.
}

var levelSpecs = []levelSpec{
	{"Balance", &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{
		Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}}}},
		"balance/gear_sets/p1", "balance/apls/default", druid.Humanoid},
	{"FeralCat", &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
		Options: &proto.FeralCatDruid_Options{ClassOptions: &proto.DruidOptions{}}}},
		"feralcat/gear_sets/default", "feralcat/apls/adaptive", druid.Cat},
	{"FeralBear", &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
		Options: &proto.FeralBearDruid_Options{ClassOptions: &proto.DruidOptions{}}}},
		"feralbear/gear_sets/default", "feralbear/apls/adaptive", druid.Bear},
	{"Restoration", &proto.Player_RestorationDruid{RestorationDruid: &proto.RestorationDruid{
		Options: &proto.RestorationDruid_Options{ClassOptions: &proto.DruidOptions{}}}},
		"balance/gear_sets/p1", "", druid.Humanoid},
}

const specDir = "../../../ui/specs/druid/"

// Spells sim/core registers on the druid at their level 60 rank, which no druid code picks.
var coreRegistered = map[int32]bool{
	9910: true, // Thorns rank 6's damage, from buffs.NewSpellDataThornsAuraWithSpellPower.
}

// The panic message building the sim raises, or "" when it builds.
func buildPanic(rsr *proto.RaidSimRequest) (msg string) {
	defer func() {
		if err := recover(); err != nil {
			msg = fmt.Sprint(err)
		}
	}()
	core.NewSim(rsr, simsignals.CreateSignals())
	return ""
}

func levelRequest(s levelSpec, level int32) *proto.RaidSimRequest {
	dir, file, _ := strings.Cut(s.gear, "/gear_sets/")
	player := &proto.Player{
		Class:       proto.Class_ClassDruid,
		Race:        proto.Race_RaceNightElf,
		Level:       level,
		Equipment:   core.GetGearSet(specDir+dir+"/gear_sets", file).GearSet,
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	if s.rotation != "" {
		dir, file, _ := strings.Cut(s.rotation, "/apls/")
		player.Rotation = core.GetAplRotation(specDir+dir+"/apls", file).Rotation
	}
	core.WithSpec(player, s.spec)
	return &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Duration: 60,
			// Level 0 takes the raid's level plus three.
			Targets: []*proto.Target{{MobType: proto.MobType_MobTypeMechanical, SwingSpeed: 2, MinBaseDamage: 100}},
		},
		SimOptions: &proto.SimOptions{Iterations: 5, IsTest: true, RandomSeed: 101},
	}
}

// Below 60 each spec sims with the spells its level has taught, and a feral spec below its form's
// level fails with a message naming the form.
func TestDruidSpecsSimBelowSixty(t *testing.T) {
	for _, s := range levelSpecs {
		for _, level := range []int32{10, 20, 30, 40} {
			if level < druid.FormLevel(s.form) {
				if msg := buildPanic(levelRequest(s, level)); !strings.Contains(msg, "Form, which a druid learns at level") {
					t.Errorf("%s at %d: want a failure naming the form, got %q", s.name, level, msg)
				}
				continue
			}
			result := core.RunSim(levelRequest(s, level), nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Errorf("%s at %d: %s", s.name, level, result.Error.Message)
				continue
			}

			sim := core.NewSim(levelRequest(s, level), simsignals.CreateSignals())
			character := sim.Raid.Parties[0].Players[0].GetCharacter()
			if character.Level != level {
				t.Fatalf("%s: the character is level %d, want %d", s.name, character.Level, level)
			}
			for _, spell := range character.Spellbook {
				if spell.ItemID != 0 || spell.SpellID == 0 {
					continue
				}
				row := spelldata.Find(spell.SpellID)
				if row == spelldata.Nil || coreRegistered[row.ID] {
					continue
				}
				if taught := spelldata.TaughtAt(row); taught > level {
					t.Errorf("%s at %d registers %d %s, taught at %d", s.name, level, row.ID, row.Name, taught)
				}
			}
		}
	}
}
