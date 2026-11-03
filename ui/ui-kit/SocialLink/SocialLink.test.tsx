import i18n from '@i18n/config';
import { type Social, SOCIALS } from '@sim/constants/other';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { SocialLink } from './SocialLink';

const [github] = SOCIALS;
const labelled: Social = { ...github, key: 'labelled', label: ' GitHub' };

describe('SocialLink', () => {
	it('renders an external link named by its tooltip, opening a new tab safely', () => {
		render(<SocialLink social={github} testId="social-github" />);
		const link = screen.getByRole('link', { name: i18n.t(github.tooltip) });
		expect(link.getAttribute('href')).toBe(github.href);
		expect(link.getAttribute('target')).toBe('_blank');
		expect(link.getAttribute('rel')).toBe('noopener noreferrer');
		expect(link.getAttribute('data-testid')).toBe('social-github');
		const icon = link.querySelector('svg')!;
		expect(icon.getAttribute('data-icon')).toBe('github');
		expect(icon.getAttribute('data-prefix')).toBe('fab');
	});

	it('renders the label beside the icon only for a social that has one', () => {
		const { unmount } = render(<SocialLink social={labelled} />);
		expect(screen.getByRole('link').textContent).toBe(labelled.label);
		unmount();

		render(<SocialLink social={github} />);
		expect(screen.getByRole('link').textContent).toBe('');
	});

	it('shows its tooltip on hover and nothing before', () => {
		render(<SocialLink social={github} />);
		const tooltip = i18n.t(github.tooltip);
		expect(screen.queryByText(tooltip)).toBeNull();

		fireEvent.mouseEnter(screen.getByRole('link'));
		expect(screen.getByText(tooltip)).toBeTruthy();
	});
});
