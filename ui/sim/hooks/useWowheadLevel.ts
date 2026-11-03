import { getWowheadLevel, subscribeWowheadLevel } from '@sim/proto/wowhead';
import { useSyncExternalStore } from 'react';

/** The level Wowhead links and tooltips are built for (`getWowheadLevel`), re-rendering when it changes. */
export const useWowheadLevel = (): number => useSyncExternalStore(subscribeWowheadLevel, getWowheadLevel);
