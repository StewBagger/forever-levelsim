import { isPreviewOrDevMode } from '@sim/utils/env';
import { afterEach, describe, expect, it, vi } from 'vitest';

describe('isPreviewOrDevMode', () => {
	afterEach(() => {
		vi.unstubAllEnvs();
	});

	it('holds on the dev server whatever the host', () => {
		expect(isPreviewOrDevMode('www.wowsims.com')).toBe(true);
	});

	describe('in a production build', () => {
		const onHost = (hostname: string) => {
			vi.stubEnv('DEV', false);
			return isPreviewOrDevMode(hostname);
		};

		it("holds on every host but upstream's", () => {
			expect(onHost('localhost')).toBe(true);
			expect(onHost('192.168.1.30')).toBe(true);
			expect(onHost('nossuary.com')).toBe(true);
			expect(onHost('wowsims-forever.pages.dev')).toBe(true);
			expect(onHost('0123456789abcdef.wowsims-forever.pages.dev')).toBe(true);
		});

		it('does not hold on wowsims.com', () => {
			expect(onHost('www.wowsims.com')).toBe(false);
			expect(onHost('wowsims.com')).toBe(false);
		});
	});
});
