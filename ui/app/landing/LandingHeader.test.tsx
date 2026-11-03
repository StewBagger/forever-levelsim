import i18n from '@i18n/config';
import { SITE_HOME_PATH } from '@sim/constants/other';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { LandingHeader } from './LandingHeader';

describe('LandingHeader', () => {
	it('opens and closes the collapsed navigation from its togglers', () => {
		render(<LandingHeader />);
		const [open, close] = screen.getAllByTestId('navbar-toggler');
		expect(open.getAttribute('aria-expanded')).toBe('false');
		expect(close.getAttribute('aria-expanded')).toBe('false');

		fireEvent.click(open);
		expect(open.getAttribute('aria-expanded')).toBe('true');
		expect(close.getAttribute('aria-expanded')).toBe('true');

		fireEvent.click(close);
		expect(open.getAttribute('aria-expanded')).toBe('false');
		expect(close.getAttribute('aria-expanded')).toBe('false');
	});

	it('opens every external nav link in a new tab with no opener', () => {
		render(<LandingHeader />);
		const names = [i18n.t('landing.header.github')];

		names.forEach(name => {
			const link = screen.getByRole('link', { name });
			expect(link.getAttribute('target')).toBe('_blank');
			expect(link.getAttribute('rel')).toBe('noopener noreferrer');
		});
	});

	it('links the brand to the site home in the same tab', () => {
		render(<LandingHeader />);
		const brand = screen.getByTestId('landing-brand');
		expect(brand.getAttribute('href')).toBe(SITE_HOME_PATH);
		expect(brand.getAttribute('target')).toBeNull();
		expect(screen.getByTestId('wowsims-title').closest('a')).toBe(brand);
	});
});
