import type { BooleanPickerConfig, ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import { DamageProfile, SpellSchool } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateSpellSchool } from '@i18n/localization';
import type { Encounter } from '@sim/raid/encounter';

import { trackEvent } from '../../../../../tracking/utils';

type NumberKey = 'tankDtps' | 'tankHitSeconds' | 'groupDamage' | 'groupDamageVariation' | 'groupCadenceSeconds' | 'groupCadenceVariation' | 'groupTargets';

const profileOf = (encounter: Encounter) => encounter.getDamageProfile() ?? DamageProfile.create();

const write = (encounter: Encounter, label: string, value: number | boolean, change: (profile: DamageProfile) => void) => {
	trackEvent({ action: 'settings', category: 'healing_group', label, value });
	const profile = profileOf(encounter);
	change(profile);
	encounter.setDamageProfile(profile);
};

interface NumberSpec {
	key: NumberKey;
	locale: string;
	float: boolean;
	// The field shows the value times this: a share as a percentage.
	scale?: number;
}

// The damage profile's numbers in the order the editor lists them. Each 0 turns its part off or takes the sim's default
// (2 s between tank hits, 8 s between pulses), as DamageProfile says.
const NUMBERS: Array<NumberSpec> = [
	{ key: 'tankDtps', locale: 'tank_dtps', float: true },
	{ key: 'tankHitSeconds', locale: 'tank_hit_seconds', float: true },
	{ key: 'groupDamage', locale: 'group_damage', float: true },
	{ key: 'groupDamageVariation', locale: 'group_damage_variation', float: true, scale: 100 },
	{ key: 'groupCadenceSeconds', locale: 'group_cadence_seconds', float: true },
	{ key: 'groupCadenceVariation', locale: 'group_cadence_variation', float: true },
	{ key: 'groupTargets', locale: 'group_targets', float: false },
];

const label = (key: string) => i18n.t(`settings_tab.encounter.healing_group.${key}.label`);
const tooltip = (key: string) => i18n.t(`settings_tab.encounter.healing_group.${key}.tooltip`);

export const damageProfileNumberConfigs = (): Array<NumberFieldConfig<Encounter>> =>
	NUMBERS.map(({ key, locale, float, scale = 1 }) => ({
		id: `encounter-healing-group-${locale.replaceAll('_', '-')}`,
		label: label(locale),
		labelTooltip: tooltip(locale),
		storeField: 'encounter:damageProfile',
		float,
		positive: true,
		getValue: (encounter: Encounter) => profileOf(encounter)[key] * scale,
		setValue: (encounter: Encounter, newValue: number) =>
			write(encounter, locale, newValue, profile => {
				profile[key] = newValue / scale;
			}),
	}));

const SCHOOLS = [
	SpellSchool.SpellSchoolPhysical,
	SpellSchool.SpellSchoolArcane,
	SpellSchool.SpellSchoolFire,
	SpellSchool.SpellSchoolFrost,
	SpellSchool.SpellSchoolHoly,
	SpellSchool.SpellSchoolNature,
	SpellSchool.SpellSchoolShadow,
];

export const groupSchoolConfig = (): ChoiceFieldConfig<Encounter> => ({
	id: 'encounter-healing-group-school',
	label: label('group_school'),
	labelTooltip: tooltip('group_school'),
	values: SCHOOLS.map(value => ({ name: translateSpellSchool(value), value })),
	storeField: 'encounter:damageProfile',
	getValue: (encounter: Encounter) => profileOf(encounter).groupSchool,
	setValue: (encounter: Encounter, newValue: number) =>
		write(encounter, 'group_school', newValue, profile => {
			profile.groupSchool = newValue;
		}),
});

export const groupHitsTankConfig = (): BooleanPickerConfig<Encounter> => ({
	id: 'encounter-healing-group-hits-tank',
	label: label('group_hits_tank'),
	labelTooltip: tooltip('group_hits_tank'),
	storeField: 'encounter:damageProfile',
	getValue: (encounter: Encounter) => profileOf(encounter).groupHitsTank,
	setValue: (encounter: Encounter, newValue: boolean) =>
		write(encounter, 'group_hits_tank', newValue, profile => {
			profile.groupHitsTank = newValue;
		}),
});
