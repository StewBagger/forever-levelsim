import { TargetDummySettings } from '@generated/proto/api';
import { DamageProfile, UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { CHARACTER_LEVEL } from '@sim/constants/mechanics';
import type { Player } from '@sim/player/player';
import {
	HEALING_GROUP_BOSS_LEVELS,
	HEALING_GROUP_DAMAGE_PROFILES,
	HEALING_GROUP_DUMMIES,
	HEALING_GROUP_DUMMY_SETTINGS,
	HEALING_GROUP_TANK_INDEX,
} from '@sim/proto/healing_group_auto_gen';
import type { Sim } from '@sim/sim';
import { batch } from '@sim/state/batch';

// The tables hold levels 1 to 60 from index 0.
const levelIndex = (level: number) => Math.min(CHARACTER_LEVEL, Math.max(1, Math.round(level))) - 1;

/** The incoming damage of a healer's fight at a level: levelsim.HealingGroupDamageProfile. */
export const healingGroupDamageProfile = (level: number): DamageProfile => DamageProfile.fromJson(HEALING_GROUP_DAMAGE_PROFILES[levelIndex(level)]);

/** The group members at a level, healable and able to die, the tank with a tank's health: levelsim.HealingGroupDummySettings. */
export const healingGroupDummySettings = (level: number): TargetDummySettings => TargetDummySettings.fromJson(HEALING_GROUP_DUMMY_SETTINGS[levelIndex(level)]);

/** The tank slot: the first target dummy, which takes the raid index after the healer's. */
export const healingGroupTanks = (): Array<UnitReference> => [UnitReference.create({ type: UnitType.Player, index: HEALING_GROUP_TANK_INDEX })];

/**
 * A healer's 5-man dungeon group at a level, as levelsim sims it: the tank and three damage dealers as healable target
 * dummies, the tank in the raid's tank slot, the level's damage profile, and a boss HEALING_GROUP_BOSS_LEVELS above the
 * level with no melee of its own, so the profile is all the damage the group takes.
 */
export const applyHealingGroup = (sim: Sim, level: number) =>
	batch(() => {
		sim.raid.setTargetDummies(HEALING_GROUP_DUMMIES);
		sim.raid.setTargetDummySettings(healingGroupDummySettings(level));
		sim.raid.setTanks(healingGroupTanks());
		sim.encounter.setDamageProfile(healingGroupDamageProfile(level));
		sim.encounter.modifyTarget(0, target => {
			target.level = level + HEALING_GROUP_BOSS_LEVELS;
			target.swingSpeed = 0;
			target.tankIndex = 0;
		});
	});

/**
 * Gives a healing spec's page the level's group when the loaded settings hold none: settings saved before the group
 * existed, or brought from another spec, would sim a healer with nobody to heal and no damage to heal.
 */
export const ensureHealingGroup = (player: Player<any>, sim: Sim) => {
	if (!player.getPlayerSpec().isHealingSpec) return;
	if (sim.raid.getTargetDummySettings()?.healable && sim.encounter.getDamageProfile()) return;
	applyHealingGroup(sim, player.getLevel());
};
