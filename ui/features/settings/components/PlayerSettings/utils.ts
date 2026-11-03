import type { ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import { applyLevelEdit } from '@features/settings/model/level_defaults';
import { Profession } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateProfession, translateRace } from '@i18n/localization';
import type { Player } from '@sim/player/player';
import type { IndividualSimHost } from '@sim/sim_host';
import { getEnumValues } from '@sim/utils/collections';

export const raceInput = (player: Player<any>): ChoiceFieldConfig<Player<any>> => ({
	id: 'simui-race',
	label: i18n.t('settings_tab.player.race'),
	values: player.getPlayerClass().races.map(race => ({ name: translateRace(race), value: race })),
	storeField: 'race' as const,
	getValue: modObject => modObject.getRace(),
	setValue: (modObject, newValue) => modObject.setRace(newValue),
});

// The game files have no Inscription skill line.
const ABSENT_PROFESSIONS = [Profession.Inscription];

export const professionInput = (which: 1 | 2): ChoiceFieldConfig<Player<any>> => ({
	id: `simui-profession${which}`,
	label: i18n.t(`settings_tab.player.profession_${which}`),
	values: (getEnumValues(Profession) as Array<Profession>)
		.filter(profession => !ABSENT_PROFESSIONS.includes(profession))
		.map(profession => ({ name: translateProfession(profession), value: profession })),
	storeField: ['profession1', 'profession2'] as const,
	getValue: modObject => (which === 1 ? modObject.getProfession1() : modObject.getProfession2()),
	setValue: (modObject, newValue) => (which === 1 ? modObject.setProfession1(newValue) : modObject.setProfession2(newValue)),
});

// A user edit runs the level's defaults, kept at or above the spec's `minLevel`; loading settings or a link sets the
// level through the player alone.
export const levelInput = (host: IndividualSimHost<any>): NumberFieldConfig<Player<any>> => ({
	id: 'simui-level',
	label: i18n.t('settings_tab.player.level.label'),
	labelTooltip: i18n.t('settings_tab.player.level.tooltip', { min: host.individualConfig.minLevel ?? 1 }),
	positive: true,
	storeField: 'level' as const,
	getValue: modObject => modObject.getLevel(),
	setValue: (_, newValue) => void applyLevelEdit(host, newValue),
});
