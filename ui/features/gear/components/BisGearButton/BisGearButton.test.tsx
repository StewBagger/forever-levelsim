import { EquipmentSpec, ItemSpec } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { Stats } from '@sim/proto/stats';
import { makeStubHost } from '@testing/storybook/fixtures/models';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { BisGearButton } from './BisGearButton';

describe('BisGearButton', () => {
	it("equips the sim's pick for the level and the current EP", async () => {
		const equipment = EquipmentSpec.create({ items: [ItemSpec.create({ id: 1 })] });
		const gear = { id: 'gear' };
		const ep = new Stats();
		const bisGear = vi.fn(async () => equipment);
		const lookupEquipmentSpec = vi.fn(() => gear);
		const setGear = vi.fn();
		const host = makeStubHost({
			player: { getEpWeights: () => ep, setGear },
			sim: { bisGear, db: { lookupEquipmentSpec } },
			playerSlice: { level: 30 },
		});
		render(
			<SimHostProvider host={host}>
				<BisGearButton />
			</SimHostProvider>,
		);

		fireEvent.click(screen.getByTestId('bis-gear-button'));
		await waitFor(() => expect(setGear).toHaveBeenCalledWith(gear));
		expect(bisGear).toHaveBeenCalledWith(host.player, 30, ep);
		expect(lookupEquipmentSpec).toHaveBeenCalledWith(equipment);
	});
});
