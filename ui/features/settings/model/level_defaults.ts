import { Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BOSS_LEVEL, CHARACTER_LEVEL } from '@sim/constants/mechanics';
import type { Player } from '@sim/player/player';
import { Stats } from '@sim/proto/stats';
import { Encounter } from '@sim/raid/encounter';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';
import { SimRunKind } from '@sim/state/sim_store';
import { toastManager } from '@ui-kit/Toast';

// Sets the character's level, kept between 1 and 60. A first target at the boss level, or at or 3 above the old level,
// moves with the character: to the same level below 60 and back to the boss level at 60. Any other target level was
// picked on purpose and stays. Null when the level is unchanged, else whether the target moved.
export const changeLevel = (player: Player<any>, newValue: number): { level: number; targetFollows: boolean } | null => {
	const oldLevel = player.getLevel();
	const level = Math.min(CHARACTER_LEVEL, Math.max(1, Math.round(newValue)));
	if (level == oldLevel) return null;
	const encounter = player.sim.encounter;
	const targetLevel = encounter.getTarget(0)?.level;
	const targetFollows = targetLevel !== undefined && [BOSS_LEVEL, oldLevel, oldLevel + 3].includes(targetLevel);
	batch(() => {
		player.setLevel(level);
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

// The spec's own talents, gear and EP, which a level 60 character starts from.
const applySpecDefaults = (host: IndividualSimHost<any>, targetFollows: boolean) => {
	const { player, sim, individualConfig: config } = host;
	batch(() => {
		player.setTalentsString(config.defaults.talents.talentsString);
		player.setGear(sim.db.lookupEquipmentSpec(config.defaults.gear));
		player.setEpWeights(config.defaults.epWeights);
		if (targetFollows) {
			const target = config.defaults.encounter?.encounter?.targets[0] ?? Encounter.defaultTargetProto();
			setTargetArmor(host, target.stats[Stat.StatArmor]);
		}
	});
};

// A level's starting talents, BIS gear and EP, worked out by the sim from the spec's default talents and the stat weights
// request. Its result is dropped when the level has changed again before it lands.
const loadLevelDefaults = async (host: IndividualSimHost<any>, level: number, targetFollows: boolean) => {
	const { player, sim, individualConfig: config } = host;
	const busy = toastManager.add({ variant: 'info', body: i18n.t('settings_tab.player.level.loading', { level }), autohide: false, canClose: false });
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
			if (targetFollows) setTargetArmor(host, result.targetArmor);
		});
		toastManager.add({ variant: 'success', body: i18n.t('settings_tab.player.level.loaded', { level }) });
	} catch (error) {
		toastManager.add({ variant: 'error', body: error instanceof Error ? error.message : String(error) });
	} finally {
		toastManager.close(busy);
	}
};

// A user's level edit: the level and the target change at once, then the level's talents, gear and EP follow. Resolves
// once they have landed, or once loading them has failed and said so; it never rejects.
export const applyLevelEdit = (host: IndividualSimHost<any>, newValue: number): Promise<void> => {
	const change = changeLevel(host.player, newValue);
	if (!change) return Promise.resolve();
	if (change.level == CHARACTER_LEVEL) {
		applySpecDefaults(host, change.targetFollows);
		return Promise.resolve();
	}
	return loadLevelDefaults(host, change.level, change.targetFollows);
};
