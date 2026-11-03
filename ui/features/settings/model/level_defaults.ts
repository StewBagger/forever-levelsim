import { maxTalentPoints } from '@features/talents/model/can_set_points';
import { Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BOSS_LEVEL, CHARACTER_LEVEL } from '@sim/constants/mechanics';
import type { Player } from '@sim/player/player';
import { Stats } from '@sim/proto/stats';
import { getTalentPoints } from '@sim/proto/utils';
import { Encounter } from '@sim/raid/encounter';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';
import { SimRunKind } from '@sim/state/sim_store';
import { errorHeadline } from '@sim/utils/format';
import { toastManager } from '@ui-kit/Toast';

// Sets the character's level, kept between the spec's minimum (`minLevel`, 1 by default) and 60. A first target at the
// boss level, or at or 3 above the old level, moves with the character: to the same level below 60 and back to the boss
// level at 60. Any other target level was picked on purpose and stays. Talents that spend more points than the new level
// has are cleared in the same write, so no stats request goes out for a build the level cannot hold. Null when the level
// is unchanged, else whether the target moved.
export const changeLevel = (player: Player<any>, newValue: number, minLevel = 1): { level: number; targetFollows: boolean } | null => {
	const oldLevel = player.getLevel();
	const level = Math.min(CHARACTER_LEVEL, Math.max(minLevel, Math.round(newValue)));
	if (level == oldLevel) return null;
	const encounter = player.sim.encounter;
	const targetLevel = encounter.getTarget(0)?.level;
	const targetFollows = targetLevel !== undefined && [BOSS_LEVEL, oldLevel, oldLevel + 3].includes(targetLevel);
	batch(() => {
		player.setLevel(level);
		if (getTalentPoints(player.getTalentsString()) > maxTalentPoints(level)) player.setTalentsString('');
		if (!targetFollows) return;
		encounter.modifyTarget(0, target => {
			target.level = level == CHARACTER_LEVEL ? BOSS_LEVEL : level;
		});
	});
	return { level, targetFollows };
};

const setTargetArmor = (host: IndividualSimHost<any>, armor: number) =>
	host.sim.encounter.modifyTarget(0, target => {
		target.stats[Stat.StatArmor] = armor;
	});

// The spec's default spec and class options and consumables, set after a level's talents and gear. Each spec's
// `derivedSettings` then settles them against those talents and gear, so a choice a lower level forced off (a talent
// it lacks, an imbue with no weapon to take it) comes back once the level allows it again.
const applyDefaultOptions = (host: IndividualSimHost<any>) => {
	const { player, individualConfig: config } = host;
	player.setSpecOptions(config.defaults.specOptions);
	player.setConsumes(config.defaults.consumables);
};

// The spec's own talents, gear, EP, options and consumables, which a level 60 character starts from.
const applySpecDefaults = (host: IndividualSimHost<any>, targetFollows: boolean) => {
	const { player, sim, individualConfig: config } = host;
	batch(() => {
		player.setTalentsString(config.defaults.talents.talentsString);
		player.setGear(sim.db.lookupEquipmentSpec(config.defaults.gear));
		player.setEpWeights(config.defaults.epWeights);
		applyDefaultOptions(host);
		if (targetFollows) {
			const target = config.defaults.encounter?.encounter?.targets[0] ?? Encounter.defaultTargetProto();
			setTargetArmor(host, target.stats[Stat.StatArmor]);
		}
	});
};

// The "working out" toast of the level load in flight, per player. A newer level edit closes it at once, so only the
// load whose result can still land is on screen.
const busyToasts = new WeakMap<Player<any>, string>();

const closeBusyToast = (player: Player<any>) => {
	const busy = busyToasts.get(player);
	if (busy === undefined) return;
	toastManager.close(busy);
	busyToasts.delete(player);
};

// A level's starting talents, BIS gear and EP, worked out by the sim from the spec's default talents and the stat weights
// request. Its result is dropped when the level has changed again before it lands.
const loadLevelDefaults = async (host: IndividualSimHost<any>, level: number, targetFollows: boolean) => {
	const { player, sim, individualConfig: config } = host;
	const busy = toastManager.add({ variant: 'info', body: i18n.t('settings_tab.player.level.loading', { level }), autohide: false, canClose: false });
	busyToasts.set(player, busy);
	try {
		const result = await sim.runs.start(SimRunKind.StatWeights, () =>
			sim.levelDefaults(
				player,
				level,
				config.defaults.talents.talentsString,
				config.epStats,
				config.epPseudoStats ?? [],
				player.getRefStat('dpsRefStat') ?? config.epReferenceStat,
			),
		);
		if (player.getLevel() != level) return;
		batch(() => {
			player.setTalentsString(result.talentsString);
			if (result.equipment) player.setGear(sim.db.lookupEquipmentSpec(result.equipment));
			if (result.dps?.epValues) player.setEpWeights(Stats.fromProto(result.dps.epValues));
			applyDefaultOptions(host);
			if (targetFollows) setTargetArmor(host, result.targetArmor);
		});
		toastManager.add({ variant: 'success', body: i18n.t('settings_tab.player.level.loaded', { level }) });
	} catch (error) {
		// A dropped load's failure is not news: the level it was for is gone.
		if (player.getLevel() == level) toastManager.add({ variant: 'error', body: errorHeadline(error) });
	} finally {
		toastManager.close(busy);
		if (busyToasts.get(player) === busy) busyToasts.delete(player);
	}
};

// A user's level edit: the level and the target change at once, then the level's talents, gear and EP follow. Resolves
// once they have landed, or once loading them has failed and said so; it never rejects.
export const applyLevelEdit = (host: IndividualSimHost<any>, newValue: number): Promise<void> => {
	const change = changeLevel(host.player, newValue, host.individualConfig.minLevel);
	if (!change) return Promise.resolve();
	closeBusyToast(host.player);
	if (change.level == CHARACTER_LEVEL) {
		applySpecDefaults(host, change.targetFollows);
		return Promise.resolve();
	}
	return loadLevelDefaults(host, change.level, change.targetFollows);
};
