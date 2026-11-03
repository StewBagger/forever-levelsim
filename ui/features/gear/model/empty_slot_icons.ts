import { ItemSlot } from '@generated/proto/common';
import { SITE_BASE } from '@sim/constants/other';

const emptySlotIcons: Record<ItemSlot, string> = {
	[ItemSlot.ItemSlotHead]: `${SITE_BASE}assets/item_slots/head.jpg`,
	[ItemSlot.ItemSlotNeck]: `${SITE_BASE}assets/item_slots/neck.jpg`,
	[ItemSlot.ItemSlotShoulder]: `${SITE_BASE}assets/item_slots/shoulders.jpg`,
	[ItemSlot.ItemSlotBack]: `${SITE_BASE}assets/item_slots/shirt.jpg`,
	[ItemSlot.ItemSlotChest]: `${SITE_BASE}assets/item_slots/chest.jpg`,
	[ItemSlot.ItemSlotWrist]: `${SITE_BASE}assets/item_slots/wrists.jpg`,
	[ItemSlot.ItemSlotHands]: `${SITE_BASE}assets/item_slots/hands.jpg`,
	[ItemSlot.ItemSlotWaist]: `${SITE_BASE}assets/item_slots/waist.jpg`,
	[ItemSlot.ItemSlotLegs]: `${SITE_BASE}assets/item_slots/legs.jpg`,
	[ItemSlot.ItemSlotFeet]: `${SITE_BASE}assets/item_slots/feet.jpg`,
	[ItemSlot.ItemSlotFinger1]: `${SITE_BASE}assets/item_slots/finger.jpg`,
	[ItemSlot.ItemSlotFinger2]: `${SITE_BASE}assets/item_slots/finger.jpg`,
	[ItemSlot.ItemSlotTrinket1]: `${SITE_BASE}assets/item_slots/trinket.jpg`,
	[ItemSlot.ItemSlotTrinket2]: `${SITE_BASE}assets/item_slots/trinket.jpg`,
	[ItemSlot.ItemSlotMainHand]: `${SITE_BASE}assets/item_slots/mainhand.jpg`,
	[ItemSlot.ItemSlotOffHand]: `${SITE_BASE}assets/item_slots/offhand.jpg`,
	[ItemSlot.ItemSlotRanged]: `${SITE_BASE}assets/item_slots/ranged.jpg`,
};

export const getEmptySlotIconUrl = (slot: ItemSlot): string => emptySlotIcons[slot];
