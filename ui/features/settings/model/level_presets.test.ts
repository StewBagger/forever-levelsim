import { LevelDefaultsResult } from '@generated/proto/api';
import { ConsumesSpec, EquipmentSpec, ItemSlot, ItemSpec, ItemType, Spec, Stat } from '@generated/proto/common';
import { SavedTalents, UIDatabase, UIItem } from '@generated/proto/ui';
import { SITE_BASE } from '@sim/constants/other';
import { Stats } from '@sim/proto/stats';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';

import { applyLevelEdit } from './level_defaults';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

// Common items db.json leaves to the leftovers, the way the worked-out gear holds them.
const LEFTOVER_HELM = UIItem.create({ id: 7915, name: 'Barbaric Iron Helm', type: ItemType.ItemTypeHead, scalingOptions: { 0: { ilvl: 30 } } });
const LEFTOVER_CHEST = UIItem.create({ id: 2870, name: 'Shining Silver Breastplate', type: ItemType.ItemTypeChest, scalingOptions: { 0: { ilvl: 30 } } });

const DEFAULT_TALENTS = '30305213002-15053105102210501';

const PRESET_FILE = {
	spec: 'warrior_dps',
	creatureArmor: Array.from({ length: 60 }, (_, i) => 20 * (i + 1)),
	levels: {
		'10': { talents: '-1', equipment: [{ id: 2870 }], ep: { stats: [1, 0.25], pseudoStats: [2] }, targetArmor: 200 },
		'30': { talents: '-150531051', equipment: [{ id: 7915 }, {}], ep: { stats: [1, 0.5862], pseudoStats: [4.0779] }, targetArmor: 1212 },
	},
};

// fetch for the page: each worked-out file by its url (a status for a missing one, or a network failure), and the
// leftover database for anything else.
const serveSite = (files: Record<string, number | object | Error>) => {
	const fetch = vi.fn((url: string) => {
		const file = files[url];
		if (file instanceof Error) return Promise.reject(file);
		if (typeof file == 'number') return Promise.resolve({ ok: false, status: file, json: () => Promise.reject(new Error('no body')) });
		if (file) return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve(file) });
		const leftovers = UIDatabase.create({ items: [LEFTOVER_HELM, LEFTOVER_CHEST] });
		return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve(UIDatabase.toJson(leftovers)) });
	});
	vi.stubGlobal('fetch', fetch);
	return fetch;
};

const presetUrl = (page: string) => `${SITE_BASE}assets/levelsim/defaults/${page}.json`;
const fetchesOf = (fetch: ReturnType<typeof serveSite>, page: string) => fetch.mock.calls.filter(([url]) => url == presetUrl(page)).length;

// A spec page's file is kept once read, so a case that needs another file runs on a page of its own.
const setup = async (spec: Spec) => {
	const config = {
		autoRotation: () => undefined,
		epStats: [Stat.StatStrength],
		epReferenceStat: Stat.StatStrength,
		presets: { epWeights: [] },
		defaults: {
			talents: SavedTalents.create({ talentsString: DEFAULT_TALENTS }),
			gear: EquipmentSpec.create(),
			epWeights: new Stats(),
			consumables: ConsumesSpec.create(),
			specOptions: undefined as unknown,
		},
	};
	const { sim, host, player } = await makeIndividualSim({ spec, config });
	sim.applyDefaults(false, false);
	config.defaults.specOptions = player.getSpecOptions();
	const levelDefaults = vi.spyOn(sim, 'levelDefaults').mockResolvedValue(LevelDefaultsResult.create({ talentsString: '-15053' }));
	const bisGear = vi.spyOn(sim, 'bisGear').mockResolvedValue(EquipmentSpec.create({ items: [ItemSpec.create({ id: 7915 })] }));
	const head = () => player.getGear().getEquippedItem(ItemSlot.ItemSlotHead)?.item.id;
	const armor = () => sim.encounter.getTarget(0)!.stats[Stat.StatArmor];
	return { host, player, levelDefaults, bisGear, head, armor };
};

describe('applyLevelEdit with worked-out levels', () => {
	afterEach(() => {
		vi.restoreAllMocks();
		serveDatabase();
	});

	it("applies a level the page's file holds without running the sim, leftover gear included", async () => {
		const fetch = serveSite({ [presetUrl('warrior_dps')]: PRESET_FILE });
		const { host, player, levelDefaults, bisGear, head, armor } = await setup(Spec.SpecDpsWarrior);
		await applyLevelEdit(host, 30);
		expect(levelDefaults).not.toHaveBeenCalled();
		expect(bisGear).not.toHaveBeenCalled();
		expect(player.getTalentsString()).toBe('-150531051');
		expect(head()).toBe(7915);
		expect(player.getEpWeights().getStat(Stat.StatAgility)).toBe(0.5862);
		expect(armor()).toBe(1212);

		// Any other level takes the EP of the level below it, with the default talents trimmed to its points, the best
		// gear in slot by that EP, picked in the gear of that level, and a creature of its level's armor.
		await applyLevelEdit(host, 25);
		expect(levelDefaults).not.toHaveBeenCalled();
		expect(bisGear).toHaveBeenCalledOnce();
		const [, level, ep] = bisGear.mock.calls[0];
		expect(level).toBe(25);
		expect(ep.getStat(Stat.StatAgility)).toBe(0.25);
		expect(player.getTalentsString()).toBe('-15053101');
		expect(player.getEpWeights().getStat(Stat.StatAgility)).toBe(0.25);
		expect(head()).toBe(7915);
		expect(armor()).toBe(500);
		expect(fetchesOf(fetch, 'warrior_dps')).toBe(1);
	});

	it('gives a level below every level of the file the EP of the lowest', async () => {
		serveSite({ [presetUrl('paladin_retribution')]: { ...PRESET_FILE, spec: 'paladin_retribution' } });
		const { host, player, bisGear, armor } = await setup(Spec.SpecRetributionPaladin);
		await applyLevelEdit(host, 5);
		expect(bisGear.mock.calls[0][2].getStat(Stat.StatAgility)).toBe(0.25);
		expect(player.getTalentsString()).toBe('');
		expect(armor()).toBe(100);
	});

	it('runs the sim for a page the site has no file for', async () => {
		serveSite({ [presetUrl('rogue_dps')]: 404 });
		const { host, levelDefaults, bisGear } = await setup(Spec.SpecRogue);
		await applyLevelEdit(host, 30);
		expect(levelDefaults).toHaveBeenCalledOnce();
		expect(bisGear).not.toHaveBeenCalled();
	});

	it('runs the sim when the file could not be fetched, and asks again on the next edit', async () => {
		const fetch = serveSite({ [presetUrl('mage_dps')]: new Error('offline') });
		const { host, levelDefaults } = await setup(Spec.SpecMage);
		await applyLevelEdit(host, 30);
		expect(levelDefaults).toHaveBeenCalledOnce();
		await applyLevelEdit(host, 31);
		expect(levelDefaults).toHaveBeenCalledTimes(2);
		expect(fetchesOf(fetch, 'mage_dps')).toBe(2);
	});
});
