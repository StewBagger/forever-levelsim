import i18n from '@i18n/config';
import { NOSSUARY_URL, NULL_SIMS_ISSUES_URL, NULL_SIMS_REPO_URL, WOWSIMS_REPO_URL, WOWSIMS_URL } from '@sim/constants/other';
import { render, screen, within } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import english from '../../../assets/locales/en/translation.json';
import { LandingFooter } from './LandingFooter';

beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

const year = new Date().getFullYear();

const hrefs = (group: HTMLElement) =>
	within(group)
		.getAllByRole('link')
		.map(link => link.getAttribute('href'));

describe('LandingFooter', () => {
	it('renders without fetching anything', () => {
		render(<LandingFooter />);
		expect(screen.getByRole('contentinfo')).toBeTruthy();
	});

	it('links Null Ossuary, the Null Sims source and its issues', () => {
		render(<LandingFooter />);
		const group = screen.getByRole('navigation', { name: i18n.t('landing.footer.sims') });
		expect(hrefs(group)).toEqual([NOSSUARY_URL, NULL_SIMS_REPO_URL, NULL_SIMS_ISSUES_URL]);
	});

	it('credits WoWSims, its source and its exporter addon', () => {
		render(<LandingFooter />);
		const group = screen.getByRole('navigation', { name: i18n.t('landing.footer.credits') });
		expect(hrefs(group)).toEqual([WOWSIMS_URL, WOWSIMS_REPO_URL]);
	});

	it('opens every link in a new tab', () => {
		render(<LandingFooter />);
		const external = screen.getAllByRole('link').filter(link => link.getAttribute('href')?.startsWith('https://'));
		expect(external.length).toBeGreaterThan(0);
		external.forEach(link => expect(link.getAttribute('target')).toBe('_blank'));
	});

	it('gives the simulator copyright to the WoWSims contributors', () => {
		render(<LandingFooter />);
		expect(screen.getByRole('contentinfo').textContent).toContain(`© 2022-${year} WoWSims contributors`);
	});
});
