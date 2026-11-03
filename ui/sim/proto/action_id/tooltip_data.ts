import type { EquippedItem } from '../equipped_item';
import type { Gear } from '../gear';
import type { WowheadTooltipItemParams, WowheadTooltipSpellParams } from '../wowhead';
import { ActionId } from './index';

export function actionIdWowheadTooltipData(
	actionId: ActionId,
	params?: Omit<WowheadTooltipItemParams, 'itemId'> | Omit<WowheadTooltipSpellParams, 'spellId'>,
): Promise<string> {
	return actionId.itemId
		? ActionId.makeItemTooltipData(actionId.itemId, params)
		: ActionId.makeSpellTooltipData(actionId.spellIdTooltipOverride || actionId.spellId, params);
}

export function equippedItemWowheadTooltipData(gear: Gear, equippedItem: EquippedItem, level?: number): Promise<string> {
	const enchantIds = [equippedItem.enchant?.effectId].filter((id): id is number => id !== undefined);
	return actionIdWowheadTooltipData(equippedItem.asActionId(), {
		level,
		itemLevel: Number(equippedItem.ilvl),
		enchantIds: enchantIds,
		randomEnchantmentId: equippedItem.randomSuffix?.id,
		setPieceIds: gear
			.asArray()
			.filter(ei => ei != null)
			.map(ei => ei!.item.id),
	});
}
