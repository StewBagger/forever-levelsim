import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useWowheadDataset } from '@sim/hooks/useWowheadDataset';
import { useWowheadLevel } from '@sim/hooks/useWowheadLevel';
import { equippedItemWowheadTooltipData } from '@sim/proto/action_id/tooltip_data';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { useMemo } from 'react';

// The tooltip marks the set pieces worn in every slot, so it follows the whole gear, not only this item, and is
// shown at the character's level.
export const useEquippedItemWowheadDataset = (item: EquippedItem | null | undefined) => {
	const gear = usePlayerStore('gear');
	const level = useWowheadLevel();
	const resolve = useMemo(() => (item ? () => equippedItemWowheadTooltipData(gear, item, level) : null), [gear, item, level]);
	return useWowheadDataset(resolve);
};
