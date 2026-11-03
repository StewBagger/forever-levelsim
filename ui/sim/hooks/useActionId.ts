import { ActionId } from '@sim/proto/action_id';
import { useEffect, useMemo, useRef, useState } from 'react';

import { useWowheadLevel } from './useWowheadLevel';

export interface ActionIdState {
	iconUrl: string;
	name: string;
	/** The wowhead URL — known without filling; '' for an id carrying neither an item nor a spell. */
	href: string;
	/** False while the icon and name are still being fetched. */
	ready: boolean;
}

// Rank and trait definition steer the wowhead request but are not part of ActionId identity,
// so they are keyed here rather than in equalityKey(), which equals() must keep matching. An id
// carrying no item, spell or other id is a placeholder whose identity is its name and icon - a
// pet family's, say - and every such id has the same equalityKey.
export const actionIdKey = (actionId: ActionId | undefined) =>
	actionId ? `${actionId.equalityKey()}|${actionId.rank}|${actionId.definitionId}` + (actionId.anyId() ? '' : `|${actionId.name}|${actionId.iconUrl}`) : '';

export const actionIdHref = (actionId: ActionId, level?: number) => {
	if (actionId.itemId) return ActionId.makeItemUrl(actionId.itemId, actionId.randomSuffixId, level);
	if (actionId.spellId) return ActionId.makeSpellUrl(actionId.spellIdTooltipOverride || actionId.spellId, actionId.rank, actionId.definitionId);
	return '';
};

/** Already carries its name or icon, or has no id to fill from. */
export const isFilled = (actionId: ActionId) => !!(actionId.name || actionId.iconUrl) || !actionId.anyId();

// What the hook holds: the fields of the id as last filled. The href is built from it at render, so a change of the
// Wowhead level (`useWowheadLevel`) moves the link without fetching the icon and name again.
interface HeldActionId {
	iconUrl: string;
	name: string;
	ready: boolean;
	source?: ActionId;
}

const EMPTY: HeldActionId = { iconUrl: '', name: '', ready: true };

const heldOf = (actionId: ActionId | undefined): HeldActionId =>
	actionId ? { iconUrl: actionId.iconUrl, name: actionId.name, ready: isFilled(actionId), source: actionId } : EMPTY;

/** Resolves an `ActionId` to the fields a component renders: icon, name and wowhead href. */
export const useActionId = (actionId: ActionId | undefined): ActionIdState => {
	const key = actionIdKey(actionId);
	const idRef = useRef(actionId);
	idRef.current = actionId;
	const level = useWowheadLevel();

	const [held, setHeld] = useState(() => heldOf(actionId));
	// Re-seeding during the render that changed the id, rather than in an effect, keeps the previous id's icon from being painted under the new one's.
	const [seenKey, setSeenKey] = useState(key);
	if (seenKey !== key) {
		setSeenKey(key);
		setHeld(heldOf(actionId));
	}

	useEffect(() => {
		const id = idRef.current;
		if (!id || heldOf(id).ready) return;

		const controller = new AbortController();
		id.fill()
			.then(filled => {
				if (!controller.signal.aborted) setHeld({ iconUrl: filled.iconUrl, name: filled.name, ready: true, source: filled });
			})
			.catch((error: unknown) => {
				if (!controller.signal.aborted) throw error;
			});
		return () => controller.abort();
	}, [key]);

	return useMemo(
		() => ({ iconUrl: held.iconUrl, name: held.name, href: held.source ? actionIdHref(held.source, level) : '', ready: held.ready }),
		[held, level],
	);
};
