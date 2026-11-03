package priest

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{18, 17, 18}

type Priest struct {
	core.Character
	SelfBuffs
	Talents              *proto.PriestTalents
	PrayerOfHealingMode  proto.PrayerOfHealingMode
	PrayerOfMendingJumps int32
	PrayerOfMendingDelay float64
	HolyNovaHealTargets  int32

	Latency float64

	InnerFocusAura *core.Aura

	VampiricEmbrace *core.Spell

	powerInLightAuras core.AuraArray
	weakenedSoulAuras core.AuraArray
}

type SelfBuffs struct {
	PreShadowform bool
	Armor         proto.PriestOptions_Armor
}

func (priest *Priest) GetCharacter() *core.Character {
	return &priest.Character
}
func (priest *Priest) GetPriest() *Priest {
	return priest
}

func (priest *Priest) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (priest *Priest) Initialize() {
	priest.eachLearnedRank(spellData.HolyFire, priest.registerHolyFireSpell)
	priest.eachLearnedRank(MindBlastRankMap, priest.registerMindBlastSpell)
	priest.eachLearnedRank(ShadowWordPainRankMap, priest.registerShadowWordPainSpell)
	priest.eachLearnedRank(ShadowWordDeathRankMap, priest.registerShadowWordDeathSpell)
	priest.eachLearnedRank(SmiteRankMap, priest.registerSmiteSpell)
	priest.eachLearnedRank(DevouringPlagueRankMap, priest.registerDevouringPlagueSpell)
	priest.eachLearnedRank(spellData.InnerFire, priest.registerInnerFire)
	priest.eachLearnedRank(spellData.Fade, priest.registerFade)
	shared.RegisterShoot(&priest.Character)

	switch priest.Race {
	case proto.Race_RaceNightElf:
		priest.eachLearnedRank(StarshardsRankMap, priest.registerStarshardsSpell)
		priest.eachLearnedRank(spellData.ElunesGrace, priest.registerElunesGrace)
	case proto.Race_RaceTroll:
		priest.eachLearnedRank(spellData.HexOfWeakness, priest.registerHexOfWeakness)
		priest.eachLearnedRank(spellData.Shadowguard, priest.registerShadowguard)
	case proto.Race_RaceUndead:
		priest.eachLearnedRank(spellData.TouchOfWeakness, priest.registerTouchOfWeakness)
		priest.eachLearnedRank(spellData.DarkSacrifice, priest.registerDarkSacrifice)
	case proto.Race_RaceDwarf:
		priest.eachLearnedRank(spellData.Chastise, priest.registerChastise)
		priest.eachLearnedRank(spellData.DesperatePrayer, priest.registerDesperatePrayer)
	}

	priest.registerPrayerOfFortitude()
	priest.registerPrayerOfSpirit()
	priest.registerPrayerOfShadowProtection()
}

func (priest *Priest) InitializeHealingSpells() {
	priest.registerPowerWordShield()
	priest.eachLearnedRank(spellData.FlashHeal, priest.registerFlashHeal)
	priest.eachLearnedRank(spellData.Heal, priest.registerHeal)
	priest.eachLearnedRank(spellData.LesserHeal, priest.registerLesserHeal)
	priest.eachLearnedRank(spellData.GreaterHeal, priest.registerGreaterHeal)
	priest.eachLearnedRank(spellData.PrayerOfHealing, priest.registerPrayerOfHealing)
	priest.eachLearnedRank(spellData.Renew, priest.registerRenew)
}

// Registers each rank of the ladder the priest knows at its level. A rank above that is aliased to the
// highest known one, so a rotation naming it casts the known rank.
func (priest *Priest) eachLearnedRank(ladder spelldata.Ladder, register func(int32, *spelldata.Spell)) {
	ladder.Each(func(n int32, rank *spelldata.Spell) {
		if spelldata.LearnedRank(&priest.Unit, ladder, rank) == rank {
			register(n, rank)
		}
	})
}

// The highest rank of the ladder the priest knows at its level, or spelldata.Nil below its first rank.
func (priest *Priest) learnedTopRank(ladder spelldata.Ladder) *spelldata.Spell {
	return spelldata.LearnedRank(&priest.Unit, ladder, ladder.Highest())
}

func (priest *Priest) registerSelfBuffSpell(rank *spelldata.Spell, aura *core.Aura) *core.Spell {
	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
		if aura.MaxStacks > 0 {
			aura.SetStacks(sim, aura.MaxStacks)
		}
	}
	config.RelatedSelfBuff = aura

	return priest.RegisterSpell(config)
}

// A prayer's own copy on every player in the raid, up once the priest casts it. It bids in the same
// category as the raid buff of its stat, so the two count once. The cast waits until nothing as
// strong holds the category, which a raid buff as strong, permanent from the start, always does.
func (priest *Priest) registerPrayer(rank *spelldata.Spell, newAura func(*core.Unit, bool, int32) *core.Aura) {
	if rank == spelldata.Nil {
		return
	}
	auras := priest.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		return newAura(unit, true, 0)
	})
	ownAura := auras.Get(&priest.Unit)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ExtraCastCondition = func(sim *core.Simulation, _ *core.Unit) bool {
		return ownAura.ShouldRefreshExclusiveEffects(sim, 0)
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		auras.ActivateAllPlayers(sim)
	}
	config.RelatedAuraArrays = auras.ToMap()

	priest.RegisterSpell(config)
}

func (priest *Priest) periodicDamageSpellConfig(rank *spelldata.Spell, classSpellMask int64) core.SpellConfig {
	tick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = classSpellMask

	config.Dot = spelldata.DotConfig(rank, tick, spelldata.RankLabel(rank, &priest.Unit))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		if rank.PeriodicCanCrit() {
			return spell.CalcPeriodicDamage(sim, target, tick.AverageAt(priest.Level), spell.OutcomeExpectedMagicCrit)
		}
		return spell.CalcPeriodicDamage(sim, target, tick.AverageAt(priest.Level), spell.OutcomeExpectedMagicAlwaysHit)
	}

	return config
}

func (priest *Priest) directDamageSpellConfig(rank *spelldata.Spell, classSpellMask int64) core.SpellConfig {
	baseDamage := rank.DamageEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = classSpellMask

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, baseDamage.AverageAt(priest.Level), spell.OutcomeMagicHitAndCrit)
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.AverageAt(priest.Level), spell.OutcomeExpectedMagicCrit)
	}

	return config
}

func (priest *Priest) directHealSpellConfig(rank *spelldata.Spell, classSpellMask int64) core.SpellConfig {
	heal := rank.HealEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = classSpellMask
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		target = priest.healTarget(target)
		priest.withRenewedHope(sim, target, spell, func() {
			spell.CalcAndDealHealing(sim, target, heal.AverageAt(priest.Level), spell.OutcomeHealingCrit)
		})
	}

	return config
}

func (priest *Priest) healTarget(target *core.Unit) *core.Unit {
	if target == nil || target.Type == core.EnemyUnit {
		return &priest.Unit
	}
	return target
}

func (priest *Priest) Reset(_ *core.Simulation) {

}

func (priest *Priest) OnEncounterStart(sim *core.Simulation) {
}

func New(char *core.Character, selfBuffs SelfBuffs, talents string) *Priest {
	priest := &Priest{
		Character: *char,
		SelfBuffs: selfBuffs,
		Talents:   &proto.PriestTalents{},
	}

	priest.FillTalentsProto(priest.Talents.ProtoReflect(), talents, TalentTreeSizes)
	priest.EnableManaBar()
	priest.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	priest.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiAtLevel(char.Class, char.Level))

	return priest
}

// Agent is a generic way to access underlying priest on any of the agents.
type PriestAgent interface {
	GetPriest() *Priest
}

func NewPriest(character *core.Character, options *proto.Player) *Priest {
	classOptions := options.GetDpsPriest().GetOptions().GetClassOptions()
	selfBuffs := SelfBuffs{
		PreShadowform: classOptions.GetPreShadowform(),
		Armor:         classOptions.GetArmor(),
	}

	basePriest := New(character, selfBuffs, options.TalentsString)
	basePriest.Latency = float64(basePriest.ChannelClipDelay.Milliseconds())

	return basePriest
}

func RegisterPriest() {
	core.RegisterAgentFactory(
		proto.Player_DpsPriest{},
		proto.Spec_SpecDpsPriest,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewPriest(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DpsPriest)
			if !ok {
				panic("Invalid spec value for Priest!")
			}
			player.Spec = playerSpec
		},
	)
}

const (
	PriestSpellDevouringPlague int64 = 1 << iota
	PriestSpellHolyFire
	PriestSpellMindBlast
	PriestSpellMindFlay
	PriestSpellPenance
	PriestSpellShadowWordDeath
	PriestSpellShadowWordPain
	PriestSpellSmite

	PriestSpellFlashHeal
	PriestSpellLesserHeal
	PriestSpellHeal
	PriestSpellGreaterHeal
	PriestSpellRenew
	PriestSpellPowerWordShield
	PriestSpellHolyNova
	PriestSpellLightwell
	PriestSpellPrayerOfHealing
	PriestSpellPrayerOfMending

	PriestSpellCureDisease
	PriestSpellAbolishDisease

	PriestShadowSpells = PriestSpellDevouringPlague |
		PriestSpellShadowWordDeath |
		PriestSpellShadowWordPain |
		PriestSpellMindFlay |
		PriestSpellMindBlast
)
