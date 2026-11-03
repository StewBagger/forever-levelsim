import { describe, expect, it } from 'vitest';

import { errorHeadline, formatDeltaText } from './format';

describe('formatDeltaText', () => {
	it('gives the percentage of the reference either way', () => {
		expect(formatDeltaText(1000, 800, 0, undefined, false, true)).toEqual({ text: '-200 (20%)', tone: 'negative' });
		expect(formatDeltaText(1000, 1200, 0, undefined, false, true)).toEqual({ text: '+200 (20%)', tone: 'positive' });
		expect(formatDeltaText(156.45, 71.17, 2, undefined, true, true)).toEqual({ text: '-85.28 (54.51%)', tone: null });
	});

	it('colours a drop positive when lower is better, and leaves the percentage out unless asked', () => {
		expect(formatDeltaText(1000, 800, 0, true)).toEqual({ text: '-200', tone: 'positive' });
	});
});

describe('errorHeadline', () => {
	it("keeps an error's first line and drops the stack trace after it", () => {
		const message = 'Feral Cat needs Cat Form, which a druid learns at level 20; this druid is level 10.\nStack Trace:\ngoroutine 1 [running]:';
		expect(errorHeadline(new Error(message))).toBe('Feral Cat needs Cat Form, which a druid learns at level 20; this druid is level 10.');
	});

	it('reads a thrown value that is not an Error as text', () => {
		expect(errorHeadline('no gear\nmore')).toBe('no gear');
	});
});
