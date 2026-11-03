import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { levelInput } from './utils';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

describe('levelInput', () => {
	const setup = async (targetLevel: number) => {
		const { sim, player } = await makeIndividualSim();
		sim.applyDefaults(false, false);
		sim.encounter.modifyTarget(0, target => {
			target.level = targetLevel;
		});
		const set = (level: number) => levelInput.setValue(player, level);
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
});
