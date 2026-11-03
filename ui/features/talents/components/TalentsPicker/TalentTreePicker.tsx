import i18n from '@i18n/config';
import { findBrokenTalents } from '@sim/talents/broken_talents';
import type { TalentsConfig } from '@sim/talents/config';
import type { TalentPoints } from '@sim/talents/talents_string';
import { totalPointsSpent, treePointTotal, withTalentPoints, withTreeCleared } from '@sim/talents/talents_string';
import { Box } from '@ui-kit/Box';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId, useMemo } from 'react';

import { canSetPoints } from '../../model/can_set_points';
import type { TalentGraph } from '../../model/tree_graph';
import { TalentPicker } from './TalentPicker';
import { TalentReqArrow } from './TalentReqArrow';

export interface TalentTreePickerProps<TalentsProto> {
	config: TalentsConfig<TalentsProto>;
	graph: TalentGraph;
	points: TalentPoints;
	treeIdx: number;
	// The points the character's level gives.
	maxPoints: number;
	onChange: (next: TalentPoints) => void;
}

const CELL = 'var(--talent-cell-size, 3.5rem)';

export const TalentTreePicker = <TalentsProto,>({ config, graph, points, treeIdx, maxPoints, onChange }: TalentTreePickerProps<TalentsProto>) => {
	const resetTooltipId = useId();
	const treeConfig = config[treeIdx];
	const treeGraph = graph.trees[treeIdx];
	const treePoints = points[treeIdx];
	const spentInTree = treePointTotal(points, treeIdx);
	const allPointsSpent = totalPointsSpent(points) >= maxPoints;

	const canAdd = useMemo(
		() => treeConfig.talents.map((_, idx) => canSetPoints(config, points, treeIdx, idx, treePoints[idx] + 1, maxPoints)),
		[config, points, treeIdx, treeConfig, treePoints, maxPoints],
	);
	const broken = useMemo(() => findBrokenTalents(config, points)[treeIdx], [config, points, treeIdx]);

	const setPoints = (talentIdx: number, newPoints: number) => {
		const clamped = Math.min(treeConfig.talents[talentIdx].maxPoints, Math.max(0, newPoints));
		// The change event fires even when the write was rejected; the facade's equality guard
		// turns the unchanged points into a no-op write.
		const accepted = canSetPoints(config, points, treeIdx, talentIdx, clamped, maxPoints);
		onChange(accepted ? withTalentPoints(points, treeIdx, talentIdx, clamped) : points);
	};

	return (
		<Box
			padding="none"
			className="ui-talents-picker-tree relative flex flex-1 flex-col not-first:-ml-px"
			testId="talent-tree"
			style={{ maxWidth: `calc(${CELL} * ${graph.numCols + 2})` }}>
			<div className="z-1 flex items-center bg-black p-2 text-base text-white">
				<span className="mr-2 flex-1 font-bold whitespace-nowrap" data-testid="talent-tree-title">
					{treeConfig.name}
				</span>
				<span className="mr-2" data-testid="talent-tree-points">
					{spentInTree}
				</span>
				<Button
					variant={null}
					className="-mr-2 leading-none text-link-danger"
					data-testid="talent-tree-reset"
					{...tooltipAnchorProps(resetTooltipId)}
					onClick={() => onChange(withTreeCleared(points, treeIdx))}>
					<Icon name="xmark" style="base" />
				</Button>
				<Tooltip id={resetTooltipId} content={i18n.t('talents_tab.reset_button.tooltip')} />
			</div>
			{/* The art fills this wrapper, so it starts where the header ends whatever height the header has. */}
			<div className="relative flex flex-1 flex-col">
				<div
					className="absolute inset-0 z-0 bg-size-[100%_100%] bg-no-repeat shadow-talent-tree"
					style={{ backgroundImage: `url('${treeConfig.backgroundUrl}')` }}
				/>
				<div
					className="relative z-1 mx-[2vw] my-2 grid max-xxxl:mx-auto max-lg:mx-10"
					data-testid="talent-tree-main"
					style={{
						gridTemplateRows: `repeat(${graph.numRows}, 1fr)`,
						gridTemplateColumns: `repeat(${graph.numCols}, 1fr)`,
						height: `calc(${CELL} * ${graph.numRows})`,
						maxWidth: `calc(${CELL} * ${graph.numCols})`,
					}}>
					{treeGraph.arrows.map(arrow => (
						<TalentReqArrow
							key={`${arrow.parentIdx}-${arrow.childIdx}`}
							arrow={arrow}
							// A child that holds points and is not broken proves its prerequisite is met.
							fulfilled={canAdd[arrow.childIdx] || (!broken[arrow.childIdx] && treePoints[arrow.childIdx] > 0)}
						/>
					))}
					{treeConfig.talents.map((talent, talentIdx) => (
						<TalentPicker
							key={`${talent.location.rowIdx}-${talent.location.colIdx}`}
							config={talent}
							points={treePoints[talentIdx]}
							canAdd={canAdd[talentIdx]}
							broken={broken[talentIdx]}
							allPointsSpent={allPointsSpent}
							zIndex={treeGraph.zIndex[talentIdx]}
							onSetPoints={newPoints => setPoints(talentIdx, newPoints)}
						/>
					))}
				</div>
			</div>
		</Box>
	);
};
