import { LaunchStatus, Phase } from '@sim/constants/other';
import { describe, expect, it } from 'vitest';

import { PlayerClasses } from './classes';
import { shownSpecs } from './player_class';
import { isSpecShown } from './player_spec';

describe('isSpecShown', () => {
	it('hides a spec that cannot sim and shows every status from alpha up', () => {
		const at = (status: LaunchStatus) => isSpecShown({ launch: { phase: Phase.Phase1, status } });
		expect([LaunchStatus.Unlaunched, LaunchStatus.GearPlanner].map(at)).toEqual([false, false]);
		expect([LaunchStatus.Alpha, LaunchStatus.Beta, LaunchStatus.Launched].map(at)).toEqual([true, true, true]);
	});
});

describe('shownSpecs', () => {
	it("keeps the class's own order and drops its hidden specs", () => {
		PlayerClasses.naturalOrder.forEach(playerClass => {
			expect(shownSpecs(playerClass)).toEqual(Object.values(playerClass.specs).filter(isSpecShown));
		});
	});
});
