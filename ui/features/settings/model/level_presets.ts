import { maxTalentPoints } from '@features/talents/model/can_set_points';
import { trimTalents } from '@features/talents/model/trim_talents';
import { LevelDefaultsResult, StatWeightValues } from '@generated/proto/api';
import { EquipmentSpec, UnitStats } from '@generated/proto/common';
import type { JsonObject, JsonValue } from '@protobuf-ts/runtime';
import { SITE_BASE } from '@sim/constants/other';
import type { Player } from '@sim/player/player';
import { Database } from '@sim/proto/database';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';

// One level's defaults as `wowsimcli levelsim defaults` writes them: the trimmed talents, the gear as ItemSpecs and the
// EP as a UnitStats, both in protojson, and the target's armor.
interface LevelPreset {
	talents: string;
	equipment: JsonValue[];
	ep: JsonObject;
	targetArmor: number;
}

// A spec page's worked-out levels, and a creature's armor at each level from 1 to 60.
export interface LevelPresetFile {
	creatureArmor: number[];
	levels: Record<string, LevelPreset>;
}

// Each spec page's file, fetched once. A failed fetch, a missing file's included, is dropped, so the next level edit asks
// again.
const files = new Map<string, Promise<LevelPresetFile>>();

// The file of a spec page: `/forever/warrior/dps/` reads `assets/levelsim/defaults/warrior_dps.json`.
const levelPresetUrl = (simLink: string): string => {
	const page = simLink.startsWith(SITE_BASE) ? simLink.slice(SITE_BASE.length) : simLink;
	return `${SITE_BASE}assets/levelsim/defaults/${page.split('/').filter(Boolean).join('_')}.json`;
};

// The spec page's file of worked-out levels, or undefined when it could not be read, and the sim works a level out
// instead.
export const levelPresetFile = async (player: Player<any>): Promise<LevelPresetFile | undefined> => {
	const url = levelPresetUrl(player.playerSpec.simLink);
	let file = files.get(url);
	if (!file) {
		file = fetch(url).then(async response => {
			if (!response.ok) throw new Error(`${url}: ${response.status}`);
			const read = (await response.json()) as LevelPresetFile;
			if (!read.levels || !Object.keys(read.levels).length || read.creatureArmor?.length != 60) throw new Error(`${url} holds no levels`);
			return read;
		});
		file.catch(() => files.delete(url));
		files.set(url, file);
	}
	return file.catch(() => undefined);
};

// The gear by item, with the common items db.json leaves to the leftovers (a level 10 Claymore) loaded.
const presetGear = async (preset: LevelPreset): Promise<EquipmentSpec> => {
	const equipment = EquipmentSpec.fromJson({ items: preset.equipment });
	await Database.loadLeftoversIfNecessary(equipment);
	return equipment;
};

// A level's talents, gear, EP and target armor from the spec page's file, as the result the sim gives. A level the file
// holds is its entry. Any other takes the EP of the nearest level below it in the file (the lowest for one below them
// all), the spec's default talents trimmed to its points, the best gear in slot for it by that EP, which sims nothing,
// and a creature of its level's armor. The gear keeps the hands of the level the EP comes from, a two-hander or a pair,
// which the sim chose there, so the player wears that level's gear and the trimmed talents while it is picked; with
// `gearOnly` the talents are left as they are.
export const presetLevelDefaults = async (
	host: IndividualSimHost<any>,
	file: LevelPresetFile,
	level: number,
	gearOnly: boolean,
): Promise<LevelDefaultsResult> => {
	const { player, sim, individualConfig: config } = host;
	const exact = file.levels[String(level)];
	if (exact) {
		return LevelDefaultsResult.create({
			talentsString: exact.talents,
			equipment: await presetGear(exact),
			dps: StatWeightValues.create({ epValues: UnitStats.fromJson(exact.ep) }),
			targetArmor: exact.targetArmor,
		});
	}
	const levels = Object.keys(file.levels)
		.map(Number)
		.sort((a, b) => a - b);
	const from = levels.filter(l => l < level).at(-1) ?? levels[0];
	const base = file.levels[String(from)];
	const epValues = UnitStats.fromJson(base.ep);
	const talentsString = gearOnly ? player.getTalentsString() : trimTalents(config.defaults.talents.talentsString, maxTalentPoints(level));
	const baseGear = await presetGear(base);
	// A newer level edit has taken over, and this level's result is dropped unseen.
	if (player.getLevel() != level) throw new Error(`level ${level} was left before its gear was picked`);
	batch(() => {
		player.setTalentsString(talentsString);
		player.setGear(sim.db.lookupEquipmentSpec(baseGear));
	});
	return LevelDefaultsResult.create({
		talentsString,
		equipment: await sim.bisGear(player, level, Stats.fromProto(epValues)),
		dps: StatWeightValues.create({ epValues }),
		targetArmor: file.creatureArmor[level - 1],
	});
};

// A level's talents and gear for a spec page that cannot sim (a gear planner): the spec's default talents trimmed to the
// level's points and the best gear in slot for the level by the player's current EP, which sims nothing. There is no
// stat weight run behind it, so the EP stays the player's, and no target to set.
export const plannerLevelDefaults = async (host: IndividualSimHost<any>, level: number, gearOnly: boolean): Promise<LevelDefaultsResult> => {
	const { player, sim, individualConfig: config } = host;
	const talentsString = gearOnly ? player.getTalentsString() : trimTalents(config.defaults.talents.talentsString, maxTalentPoints(level));
	player.setTalentsString(talentsString);
	return LevelDefaultsResult.create({ talentsString, equipment: await sim.bisGear(player, level, player.getEpWeights()) });
};
