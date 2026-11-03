import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { LaunchStatus } from '@sim/constants/other';
import type { SimStatus } from '@sim/player/player_spec';
import type { ReactNode } from 'react';

const globalKnownIssues: ReactNode[] = [];

const statusNotice = (status: LaunchStatus): string => {
	switch (status) {
		case LaunchStatus.Unlaunched:
			return i18n.t('info.status.unlaunched');
		case LaunchStatus.GearPlanner:
			return i18n.t('info.status.gear_planner');
		case LaunchStatus.Alpha:
			return i18n.t('info.status.alpha');
		case LaunchStatus.Beta:
			return i18n.t('info.status.beta');
		default:
			return '';
	}
};

const SPELL_CASTER_SPECS: ReadonlySet<Spec> = new Set([Spec.SpecBalanceDruid, Spec.SpecMage, Spec.SpecDpsPriest, Spec.SpecElementalShaman, Spec.SpecWarlock]);

export const knownIssuesFor = (simStatus: SimStatus, spec: Spec, knownIssues?: ReadonlyArray<ReactNode>): ReactNode[] => {
	const notice = statusNotice(simStatus.status);
	const casterNotices = SPELL_CASTER_SPECS.has(spec) ? [i18n.t('info.spell_resistance')] : [];
	return [i18n.t('info.levelling_unverified'), ...(notice ? [notice] : []), ...casterNotices, ...(knownIssues ?? []), ...globalKnownIssues];
};
