import type { ActionId } from '@sim/proto/action_id';
import { actionIdWowheadTooltipData } from '@sim/proto/action_id/tooltip_data';
import { useMemo } from 'react';

import { useWowheadDataset } from './useWowheadDataset';
import { useWowheadLevel } from './useWowheadLevel';

export const useActionIdWowheadDataset = (actionId: ActionId | null | undefined, useBuffAura?: boolean) => {
	const level = useWowheadLevel();
	const resolve = useMemo(() => (actionId ? () => actionIdWowheadTooltipData(actionId, { useBuffAura, level }) : null), [actionId, useBuffAura, level]);
	return useWowheadDataset(resolve);
};
