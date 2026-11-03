import { Spec, SpellSchool, UnitReference_Type as UnitType } from '@generated/proto/common';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { applyHealingGroup, ensureHealingGroup, healingGroupDamageProfile, healingGroupDummySettings } from './healing_group';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

describe('healingGroupDamageProfile', () => {
	// The numbers TestHealerPriestGroup sims with, from levelsim.HealingGroupDamageProfile.
	it.each([
		[20, 35.9, 50.3],
		[30, 51.9, 87.8],
		[40, 68.4, 135.3],
		[60, 127.0, 260.3],
	])('at level %i deals the tank %f DTPS and each group pulse %f', (level, tankDtps, pulse) => {
		const profile = healingGroupDamageProfile(level);
		expect(profile.tankDtps).toBeCloseTo(tankDtps, 1);
		expect(profile.groupDamage).toBeCloseTo(pulse, 1);
		expect(profile).toMatchObject({
			tankHitSeconds: 2,
			groupCadenceSeconds: 8,
			groupCadenceVariation: 2,
			groupTargets: 2,
			groupSchool: SpellSchool.SpellSchoolFire,
		});
	});

	it('holds a level outside 1-60 at the nearest end', () => {
		expect(healingGroupDamageProfile(0)).toEqual(healingGroupDamageProfile(1));
		expect(healingGroupDamageProfile(70)).toEqual(healingGroupDamageProfile(60));
	});
});

describe('applyHealingGroup', () => {
	it('gives the raid four healable dummies with the first in the tank slot, and the encounter the level and its damage', async () => {
		const { sim } = await makeIndividualSim();

		applyHealingGroup(sim, 30);

		expect(sim.raid.getTargetDummies()).toBe(4);
		expect(sim.raid.getTargetDummySettings()).toEqual(healingGroupDummySettings(30));
		expect(sim.raid.getTargetDummySettings()).toMatchObject({ healable: true, trackDeaths: true });
		expect(sim.raid.getTanks()).toEqual([{ type: UnitType.Player, index: 1, owner: undefined }]);
		expect(sim.encounter.getDamageProfile()).toEqual(healingGroupDamageProfile(30));
		expect(sim.encounter.getTarget(0)).toMatchObject({ level: 32, swingSpeed: 0, tankIndex: 0 });
	});

	it('rides in the sim request and in the saved settings', async () => {
		const { sim, host } = await makeIndividualSim();
		applyHealingGroup(sim, 20);

		const request = sim.makeRaidSimRequest();
		expect(request.raid!.targetDummySettings).toEqual(healingGroupDummySettings(20));
		expect(request.encounter!.damageProfile).toEqual(healingGroupDamageProfile(20));

		const saved = host.toProto();
		applyHealingGroup(sim, 50);
		host.fromProto(saved);
		expect(sim.raid.getTargetDummySettings()).toEqual(healingGroupDummySettings(20));
		expect(sim.encounter.getDamageProfile()).toEqual(healingGroupDamageProfile(20));
	});
});

describe('ensureHealingGroup', () => {
	it("gives a healer's loaded settings the level's group when they hold none, and keeps one they hold", async () => {
		const { sim, player } = await makeIndividualSim({ spec: Spec.SpecHealerPriest });
		player.setLevel(40);
		sim.raid.setTargetDummies(9);

		ensureHealingGroup(player, sim);
		expect(sim.raid.getTargetDummies()).toBe(4);
		expect(sim.encounter.getDamageProfile()).toEqual(healingGroupDamageProfile(40));

		sim.encounter.setDamageProfile({ ...healingGroupDamageProfile(40), tankDtps: 1 });
		ensureHealingGroup(player, sim);
		expect(sim.encounter.getDamageProfile()!.tankDtps).toBe(1);
	});

	it('leaves a damage dealer alone', async () => {
		const { sim, player } = await makeIndividualSim({ spec: Spec.SpecDpsPriest });

		ensureHealingGroup(player, sim);
		expect(sim.raid.getTargetDummies()).toBe(0);
		expect(sim.encounter.getDamageProfile()).toBeUndefined();
	});
});
