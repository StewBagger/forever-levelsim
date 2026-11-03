import type { ChoiceFieldConfig, NumberFieldConfig } from '@features/inputs/configs';
import { Profession } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateProfession, translateRace } from '@i18n/localization';
import { BOSS_LEVEL, CHARACTER_LEVEL } from '@sim/constants/mechanics';
import type { Player } from '@sim/player/player';
import { batch } from '@sim/state/batch';
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

// A first target at the boss level, or at or 3 above the old level, moves with the character: to the same level below
// 60 and back to the boss level at 60. Any other target level was picked on purpose and stays.
export const levelInput: NumberFieldConfig<Player<any>> = {
	id: 'simui-level',
	label: i18n.t('settings_tab.player.level.label'),
	labelTooltip: i18n.t('settings_tab.player.level.tooltip'),
	positive: true,
	storeField: 'level' as const,
	getValue: modObject => modObject.getLevel(),
	setValue: (modObject, newValue) => {
		const oldLevel = modObject.getLevel();
		const level = Math.min(CHARACTER_LEVEL, Math.max(1, Math.round(newValue)));
		if (level == oldLevel) return;
		batch(() => {
			modObject.setLevel(level);
			const encounter = modObject.sim.encounter;
			const targetLevel = encounter.getTarget(0)?.level;
			if (targetLevel === undefined || ![BOSS_LEVEL, oldLevel, oldLevel + 3].includes(targetLevel)) return;
			encounter.modifyTarget(0, target => {
				target.level = level == CHARACTER_LEVEL ? BOSS_LEVEL : level;
			});
		});
	},
};
