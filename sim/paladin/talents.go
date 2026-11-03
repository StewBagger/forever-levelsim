package paladin

func (paladin *Paladin) ApplyTalents() {
	paladin.registerTalentSpells()

	paladin.registerHolyTalents()
	paladin.registerProtectionTalents()
	paladin.registerRetributionTalents()
}

// The abilities a talent point buys.
func (paladin *Paladin) registerTalentSpells() {
	// Holy
	if paladin.Talents.DivineFavor {
		paladin.registerDivineFavor()
	}
	if paladin.Talents.OathOfLight {
		paladin.registerOathOfLight()
	}
	if paladin.Talents.HolyShock {
		paladin.registerHolyShock()
	}
	if paladin.Talents.LightsVigil {
		paladin.registerLightsVigil()
	}

	// Protection
	if paladin.Talents.SwiftJudgement {
		paladin.registerSwiftJudgement()
	}
	if paladin.Talents.TemplarsBulwark {
		paladin.registerTemplarsBulwark()
	}
	if paladin.Talents.HolyShield {
		paladin.eachLearnedRank(HolyShieldRankMap, paladin.registerHolyShield)
	}

	// Retribution
	if paladin.Talents.SealOfCommand {
		paladin.eachLearnedRank(SealOfCommandRankMap, paladin.registerSealOfCommand)
	}
}
