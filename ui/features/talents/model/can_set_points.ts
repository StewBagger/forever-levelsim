// Whether a proposed point change is accepted.
//
// A point can always be taken back, even when that breaks a talent in a lower row or a child of the
// talent: the broken talents keep their points, the picker marks them and the sim still applies them,
// so a build can be tested with a point taken out. A point is only added where the talent would not
// itself be broken.

import { findBrokenTalents } from '@sim/talents/broken_talents';
import type { TalentsConfig } from '@sim/talents/config';
import type { TalentPoints } from '@sim/talents/talents_string';
import { totalPointsSpent, withTalentPoints } from '@sim/talents/talents_string';

// A character earns one talent point per level from 10 onwards: 51 at level 60, none below 10.
export const maxTalentPoints = (level: number): number => Math.max(level - 9, 0);

export const canSetPoints = <TalentsProto>(
	config: TalentsConfig<TalentsProto>,
	points: TalentPoints,
	treeIdx: number,
	talentIdx: number,
	newPoints: number,
	maxPoints: number,
): boolean => {
	const oldPoints = points[treeIdx][talentIdx];
	if (newPoints <= oldPoints) return true;
	if (totalPointsSpent(points) + (newPoints - oldPoints) > maxPoints) return false;

	return !findBrokenTalents(config, withTalentPoints(points, treeIdx, talentIdx, newPoints))[treeIdx][talentIdx];
};
