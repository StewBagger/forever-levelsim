package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{17, 17, 17}

type Paladin struct {
	core.Character

	Talents *proto.PaladinTalents

	Forbearance       *core.Aura
	RighteousFuryAura *core.Aura

	Judgement *core.Spell

	// The seal the paladin is under, and the ranks it has cast so far this iteration.
	currentSeal *sealConfig

	// The judgement effects this paladin can put on enemies; its melee strikes refresh them.
	JudgementAuras []core.AuraArray

	// Unsealed Wrath: one Echo per seal that leaves one.
	echoes []*sealEcho

	// The mark Consecration lays on its bonus targets, one aura per enemy, which carries Consecrated Ground.
	consecrationMarks core.AuraArray
	// Consecrated Ground: the talent rank's points for the mark's damage from caster mod; nil without it.
	consecratedGroundPoints *spelldata.Effect

	// Light's Vigil: one entry per rank, so Holy Shock can find the vigil it consumes.
	lightsVigils []*lightsVigil
	// Oath of Light: nil without the talent.
	oathOfLight *oathOfLight

	// What gear adds to numbers the spells read as they register. Item effects and set bonuses
	// apply before Initialize, so the spells pick these up.
	sealOfTheCrusaderBonusAttackPower float64
	judgementOfTheCrusaderBonus       float64
	flashOfLightBonusHealing          float64
	holyShieldBlockValueMultiplier    float64
	forbearanceReduction              time.Duration

	sealOfFuryTaunt *core.Spell
}

// Implemented by each Paladin spec.
type PaladinAgent interface {
	GetPaladin() *Paladin
}

func (paladin *Paladin) GetCharacter() *core.Character {
	return &paladin.Character
}

func (paladin *Paladin) GetPaladin() *Paladin {
	return paladin
}

func (paladin *Paladin) AddRaidBuffs(_ *proto.RaidBuffs) {
}

func (paladin *Paladin) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (paladin *Paladin) Initialize() {
	paladin.registerForbearance()
	paladin.registerRighteousFury()

	paladin.registerJudgement()
	paladin.registerSeals()
	paladin.registerAuras()

	paladin.eachLearnedRank(HolyStrikeRankMap, paladin.registerHolyStrike)
	paladin.registerConsecrationMark()
	paladin.eachLearnedRank(ConsecrationRankMap, paladin.registerConsecration)
	paladin.eachLearnedRank(ExorcismRankMap, paladin.registerExorcism)
	paladin.eachLearnedRank(HammerOfWrathRankMap, paladin.registerHammerOfWrath)
	paladin.eachLearnedRank(HolyWrathRankMap, paladin.registerHolyWrath)

	paladin.eachLearnedRank(HolyLightRankMap, paladin.registerHolyLight)
	paladin.eachLearnedRank(FlashOfLightRankMap, paladin.registerFlashOfLight)
	paladin.eachLearnedRank(LayOnHandsRankMap, paladin.registerLayOnHands)
}

// Registers each rank of the ladder the paladin has learned at its level. A rank above the level is
// aliased to the highest learned one below it, so a rotation naming that rank casts the learned one.
func (paladin *Paladin) eachLearnedRank(ladder spelldata.Ladder, register func(n int32, rank *spelldata.Spell)) {
	ladder.Each(func(n int32, rank *spelldata.Spell) {
		if spelldata.TaughtAt(rank) <= paladin.Level {
			register(n, rank)
			return
		}
		spelldata.LearnedRank(&paladin.Unit, ladder, rank)
	})
}

// The single-rank spell if the paladin has learned it at its level, or nil.
func (paladin *Paladin) learned(ladder spelldata.Ladder) *spelldata.Spell {
	if rank := spelldata.LearnedRank(&paladin.Unit, ladder, ladder.Highest()); rank != spelldata.Nil {
		return rank
	}
	return nil
}

func (paladin *Paladin) Reset(_ *core.Simulation) {
	paladin.currentSeal = nil
	for _, echo := range paladin.echoes {
		echo.seal = nil
	}
}

func (paladin *Paladin) OnEncounterStart(_ *core.Simulation) {
}

func NewPaladin(character *core.Character, talentsStr string, _ *proto.PaladinOptions) *Paladin {
	paladin := &Paladin{
		Character: *character,
		Talents:   &proto.PaladinTalents{},

		holyShieldBlockValueMultiplier: 1,
	}

	paladin.FillTalentsProto(paladin.Talents.ProtoReflect(), talentsStr, TalentTreeSizes)

	paladin.PseudoStats.CanParry = true

	paladin.EnableManaBar()

	paladin.EnableAutoAttacks(paladin, core.AutoAttackOptions{
		MainHand:       paladin.WeaponFromMainHand(),
		AutoSwingMelee: true,
	})

	paladin.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	paladin.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiAtLevel(character.Class, character.Level))
	paladin.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	return paladin
}
