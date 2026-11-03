package core

// AliasRank records that the unit registered the rank `learned` where its class code names the rank
// `named`. A rotation, a gear set and the UI name the rank the class registers at max level; a
// character below that level knows a lower rank of the same spell, and every lookup by the named
// spell id answers with the learned one.
func (unit *Unit) AliasRank(named int32, learned int32) {
	if named == learned {
		return
	}
	if unit.rankAliases == nil {
		unit.rankAliases = make(map[int32]int32)
	}
	unit.rankAliases[named] = learned
}

// resolveRankAlias turns an action naming an aliased rank into the rank registered in its place. The
// unit's own aliases are read first, then every raid member's, since a debuff on a target carries
// the spell id of the caster's rank.
func (unit *Unit) resolveRankAlias(actionID ActionID) ActionID {
	if actionID.SpellID == 0 {
		return actionID
	}
	if learned, ok := unit.rankAliases[actionID.SpellID]; ok {
		actionID.SpellID = learned
		return actionID
	}
	if unit.Env == nil || unit.Env.Raid == nil {
		return actionID
	}
	for _, other := range unit.Env.Raid.AllUnits {
		if learned, ok := other.rankAliases[actionID.SpellID]; ok {
			actionID.SpellID = learned
			return actionID
		}
	}
	return actionID
}

// GetAuraByID finds an aura by action, reading an aliased rank's id as the rank registered in its
// place.
func (unit *Unit) GetAuraByID(actionID ActionID) *Aura {
	if aura := unit.auraTracker.GetAuraByID(actionID); aura != nil {
		return aura
	}
	if resolved := unit.resolveRankAlias(actionID); resolved != actionID {
		return unit.auraTracker.GetAuraByID(resolved)
	}
	return nil
}
