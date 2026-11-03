import type { TalentsConfig } from '@sim/talents/config';
import { newTalentsConfig } from '@sim/talents/config';
import { parseTalentsString } from '@sim/talents/talents_string';
import { describe, expect, it } from 'vitest';

import { canSetPoints, maxTalentPoints } from './can_set_points';

type Fake = Record<string, never>;

const ROWS = 9;
const COLS = 4;

const tree = (name: string, prereqs: Record<string, { rowIdx: number; colIdx: number }> = {}) => ({
	name,
	backgroundUrl: 'background.jpg',
	talents: Array.from({ length: ROWS * COLS }, (_, index) => {
		const location = { rowIdx: Math.floor(index / COLS), colIdx: index % COLS };
		return {
			fieldName: `talent${index}`,
			fancyName: `Talent ${index}`,
			location,
			spellId: 1000 + index,
			maxPoints: 5,
			...(prereqs[String(index)] ? { prereqLocation: prereqs[String(index)] } : {}),
		};
	}),
});

const ROOT_TALENT = { idx: 0, location: { rowIdx: 0, colIdx: 0 } };
const GATED_TALENT = { idx: 8, location: { rowIdx: 2, colIdx: 0 } };

const config: TalentsConfig<Fake> = newTalentsConfig<Fake>([tree('A', { [GATED_TALENT.idx]: ROOT_TALENT.location }), tree('B'), tree('C')]);

const at = (talentsString: string) => parseTalentsString(config, talentsString);
const can = (talentsString: string, treeIdx: number, talentIdx: number, newPoints: number, level = 60) =>
	canSetPoints(config, at(talentsString), treeIdx, talentIdx, newPoints, maxTalentPoints(level));

describe('canSetPoints — adding', () => {
	it('allows a first point in row 0', () => {
		expect(can('', 0, 0, 1)).toBe(true);
	});

	it('refuses row 1 until the rows above hold 5 points', () => {
		expect(can('4', 0, 4, 1)).toBe(false);
		expect(can('5', 0, 4, 1)).toBe(true);
	});

	it('counts only the rows above, not the row being unlocked', () => {
		expect(can('5000' + '4', 0, 9, 1)).toBe(false);
		expect(can('5000' + '5', 0, 9, 1)).toBe(true);
		expect(can('5000' + '45', 0, 9, 1)).toBe(true);
	});

	it('does not count the points of a broken talent towards the next tier', () => {
		expect(can('4000' + '55', 0, 9, 1)).toBe(false);
	});

	it('refuses a talent whose prerequisite is not filled', () => {
		expect(can('45505', 0, 8, 1)).toBe(false);
		expect(can('55505', 0, 8, 1)).toBe(true);
	});

	it('refuses a 52nd point across all three trees', () => {
		expect(can('5555555555-1', 0, 10, 1)).toBe(false);
		expect(can('5555555555', 0, 10, 1)).toBe(true);
	});

	it('caps the points at the level less 9', () => {
		expect(can('5', 0, 1, 1, 15)).toBe(true);
		expect(can('51', 0, 1, 2, 15)).toBe(false);
		expect(can('', 0, 0, 1, 9)).toBe(false);
	});
});

describe('canSetPoints — removing', () => {
	it('allows a decrement while more points are spent than the level gives', () => {
		expect(can('5555', 0, 0, 4, 20)).toBe(true);
	});

	it('allows a plain decrement', () => {
		expect(can('5', 0, 0, 4)).toBe(true);
	});

	it('allows a decrement that breaks a talent in a lower row', () => {
		expect(can('50005', 0, 0, 4)).toBe(true);
	});

	it('allows a decrement that breaks a filled child', () => {
		expect(can('555550005', 0, 0, 4)).toBe(true);
	});

	it('lets a broken talent give its points back', () => {
		expect(can('4000' + '0000' + '1', 0, 8, 0)).toBe(true);
	});
});
