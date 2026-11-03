import { LevelDefaultsResult } from '@generated/proto/api';
import { ConsumesSpec, EquipmentSpec, Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { Stats } from '@sim/proto/stats';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { toastManager } from '@ui-kit/Toast';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';

import { applyLevelEdit, changeLevel } from './level_defaults';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

describe('changeLevel', () => {
	const setup = async (targetLevel: number) => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		sim.encounter.modifyTarget(0, target => {
			target.level = targetLevel;
		});
		const set = (level: number) => changeLevel(player, level);
		const target = () => sim.encounter.getTarget(0)!.level;
		return { player, set, target };
	};

	it('keeps the level between 1 and 60', async () => {
		const { player, set } = await setup(63);
		set(75);
		expect(player.getLevel()).toBe(60);
		set(0);
		expect(player.getLevel()).toBe(1);
	});

	it('moves a boss level target to the new level, and back to the boss level at 60', async () => {
		const { set, target } = await setup(63);
		set(40);
		expect(target()).toBe(40);
		set(45);
		expect(target()).toBe(45);
		set(60);
		expect(target()).toBe(63);
	});

	it('moves a target 3 above the old level', async () => {
		const { set, target } = await setup(43);
		set(40);
		expect(target()).toBe(43);
		set(30);
		expect(target()).toBe(30);
	});

	it('leaves a target level picked on purpose', async () => {
		const { set, target } = await setup(61);
		set(40);
		expect(target()).toBe(61);
	});

	it("keeps the level at or above the spec's minimum", async () => {
		const { player } = await setup(63);
		changeLevel(player, 10, 20);
		expect(player.getLevel()).toBe(20);
	});

	it('clears talents that spend more points than the new level has', async () => {
		const { player, set } = await setup(63);
		player.setTalentsString('11-11-11');
		const talents = player.getTalentsString();
		expect(talents).not.toBe('');
		set(20);
		expect(player.getTalentsString()).toBe(talents);
		set(12);
		expect(player.getTalentsString()).toBe('');
	});
});

describe('applyLevelEdit', () => {
	const FULL_TALENTS = '5555555555-5555555555-5555555555';

	afterEach(() => {
		vi.restoreAllMocks();
	});

	const setup = async () => {
		const config = {
			autoRotation: () => undefined,
			epStats: [Stat.StatStrength],
			epReferenceStat: Stat.StatStrength,
			presets: { epWeights: [] },
			defaults: {
				talents: SavedTalents.create({ talentsString: FULL_TALENTS }),
				gear: EquipmentSpec.create(),
				epWeights: new Stats(),
				consumables: ConsumesSpec.create({ flaskId: 13510, mhImbueId: 20749 }),
				specOptions: undefined as unknown,
			},
		};
		const made = await makeIndividualSim({ spec: Spec.SpecDpsWarrior, config });
		const { sim, host } = made;
		const player = made.player as Player<Spec.SpecDpsWarrior>;
		sim.applyDefaults(false, false);
		const options = player.getSpecOptions();
		options.classOptions!.useBattleShout = true;
		config.defaults.specOptions = options;
		// What a lower level forced off.
		const forceOff = () => {
			const off = player.getSpecOptions();
			off.classOptions!.useBattleShout = false;
			player.setSpecOptions(off);
			player.setConsumes(ConsumesSpec.create());
		};
		const loads: Array<(result: LevelDefaultsResult) => void> = [];
		vi.spyOn(sim, 'levelDefaults').mockImplementation(() => new Promise(resolve => loads.push(resolve)));
		const shown = vi.spyOn(toastManager, 'add');
		const closed = vi.spyOn(toastManager, 'close');
		return { player, host, forceOff, loads, shown, closed };
	};

	it("restores the spec's default options and consumables once a level's talents and gear land", async () => {
		const { player, host, forceOff, loads } = await setup();
		const edit = applyLevelEdit(host, 40);
		forceOff();
		await vi.waitFor(() => expect(loads).toHaveLength(1));
		loads[0](LevelDefaultsResult.create({ talentsString: '5555555' }));
		await edit;
		expect(player.getSpecOptions().classOptions!.useBattleShout).toBe(true);
		expect(player.getConsumes()).toMatchObject({ flaskId: 13510, mhImbueId: 20749 });
	});

	it('restores them exactly to the defaults at 60', async () => {
		const { player, host, forceOff } = await setup();
		player.setLevel(40);
		forceOff();
		await applyLevelEdit(host, 60);
		expect(player.getSpecOptions().classOptions!.useBattleShout).toBe(true);
		expect(player.getConsumes()).toMatchObject({ flaskId: 13510, mhImbueId: 20749 });
	});

	it('closes the working-out toast of a load a newer edit has replaced', async () => {
		const { host, loads, shown, closed } = await setup();
		void applyLevelEdit(host, 40);
		await vi.waitFor(() => expect(loads).toHaveLength(1));
		const first = shown.mock.results[0].value as string;
		expect(closed).not.toHaveBeenCalledWith(first);
		void applyLevelEdit(host, 30);
		expect(closed).toHaveBeenCalledWith(first);
	});

	it('shows only the first line of a failed load', async () => {
		const { host, shown } = await setup();
		vi.mocked(host.sim.levelDefaults).mockRejectedValue(new Error('Feral Cat needs Cat Form.\nStack Trace:\ngoroutine 1'));
		await applyLevelEdit(host, 40);
		expect(shown).toHaveBeenLastCalledWith(expect.objectContaining({ variant: 'error', body: 'Feral Cat needs Cat Form.' }));
	});
});
