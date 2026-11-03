import { maxTalentPoints } from '@features/talents/model/can_set_points';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { professionNames } from '@sim/proto/names';
import { getTalentPoints } from '@sim/proto/utils';
import type { SimWarning } from '@sim/sim_host';
import { subscribeAll, subscribePlayerField, subscribeRaidField } from '@sim/state/subscriptions';

export function defaultSimWarnings<SpecType extends number>(player: Player<SpecType>): Array<SimWarning> {
	return [
		{
			updateOn: subscribeAll([
				subscribePlayerField(player, 'gear'),
				subscribePlayerField(player, 'profession1'),
				subscribePlayerField(player, 'profession2'),
			]),
			getContent: () => {
				const failedProfReqs = player.getGear().getFailedProfessionRequirements(player.getProfessions());
				if (failedProfReqs.length == 0) {
					return '';
				}

				return failedProfReqs.map(fpr =>
					i18n.t('sidebar.warnings.profession_requirement', {
						itemName: fpr.name,
						professionName: professionNames.get(fpr.requiredProfession)!,
					}),
				);
			},
		},
		{
			updateOn: subscribeAll([subscribePlayerField(player, 'talentsString'), subscribePlayerField(player, 'level')]),
			getContent: () => {
				const talentPoints = getTalentPoints(player.getTalentsString());
				const maxPoints = maxTalentPoints(player.getLevel());

				// Skipped during initial load, when the string is still empty.
				if (talentPoints == 0 || talentPoints >= maxPoints) {
					return '';
				}

				return i18n.t('sidebar.warnings.unspent_talent_points', {
					count: maxPoints - talentPoints,
				});
			},
		},
		{
			updateOn: subscribePlayerField(player, 'talentsString'),
			getContent: () => {
				const brokenTalents = player.getBrokenTalentNames();
				if (brokenTalents.length == 0) {
					return '';
				}

				return i18n.t('sidebar.warnings.broken_talents', { talents: brokenTalents.join(', ') });
			},
		},
		{
			updateOn: subscribeRaidField(player.sim.raid, 'buffs'),
			getContent: () => {
				const buffs = player.sim.raid.getBuffs();
				if (!buffs.windfuryTotem || !buffs.flametongueTotem) {
					return '';
				}

				return i18n.t('sidebar.warnings.windfury_over_flametongue');
			},
		},
	];
}
