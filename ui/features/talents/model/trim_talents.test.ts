import { describe, expect, it } from 'vitest';

import { trimTalents } from './trim_talents';

describe('trimTalents', () => {
	// The cases of TestTrimTalents in sim/levelsim/talents_test.go, which the sim's trim passes.
	it.each([
		// The deepest tree is spent first, row by row; trees that tie go in string order.
		['05032-055-50532310253101', 11, '--5051'],
		['05032-055-50532310253101', 0, ''],
		['05032-055-50532310253101', 51, '05032-055-50532310253101'],
		['05032-055-50532310253101', 31, '--50532310253101'],
		['05032-055-50532310253101', 41, '05032--50532310253101'],
	])('trims %s to %i points as %s', (talents, points, want) => {
		expect(trimTalents(talents, points)).toBe(want);
	});
});
