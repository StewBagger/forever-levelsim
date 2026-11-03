import { ArmorType, Class, Race, RangedWeaponType, WeaponType } from '@generated/proto/common';

import { isSpecShown, PlayerSpec } from './player_spec';
export type IconSize = 'small' | 'medium' | 'large';

export interface EligibleWeaponType {
	weaponType: WeaponType;
	canUseTwoHand?: boolean;
}

export abstract class PlayerClass<ClassType extends Class> {
	static classID: Class;
	static friendlyName: string;
	static hexColor: string;
	static specs: Record<string, PlayerSpec<any>>;
	static races: Race[];
	static armorTypes: ArmorType[];
	static weaponTypes: EligibleWeaponType[];
	static rangedWeaponTypes: RangedWeaponType[];

	abstract readonly classID: ClassType;
	abstract readonly friendlyName: string;
	abstract readonly hexColor: string;
	abstract readonly specs: Record<string, PlayerSpec<any>>;
	abstract readonly races: Race[];
	abstract readonly armorTypes: ArmorType[];
	abstract readonly weaponTypes: EligibleWeaponType[];
	abstract readonly rangedWeaponTypes: RangedWeaponType[];

	abstract getIcon(size: IconSize): string;
}

// The class's specs that are offered (isSpecShown), in the class's own order. A class with none is not offered.
export const shownSpecs = (playerClass: PlayerClass<Class>): Array<PlayerSpec<any>> => Object.values(playerClass.specs).filter(isSpecShown);
