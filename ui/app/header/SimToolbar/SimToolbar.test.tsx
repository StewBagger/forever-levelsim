import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { SimToolbarView } from './SimToolbar';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

describe('SimToolbarView', () => {
	it('lists the known issues and hides the entry when there are none', () => {
		const { rerender } = render(<SimToolbarView knownIssues={[]} onOpenSettings={() => {}} />);
		expect(screen.queryByTestId('known-issues')).toBeNull();

		rerender(<SimToolbarView knownIssues={['Rage decay is off']} onOpenSettings={() => {}} />);
		expect(screen.getByTestId('known-issues').textContent).toBe('info.known_issues');
	});

	it('opens the settings from the gear', () => {
		const onOpenSettings = vi.fn();
		render(<SimToolbarView knownIssues={[]} onOpenSettings={onOpenSettings} />);

		fireEvent.click(screen.getByTestId('sim-options'));

		expect(onOpenSettings).toHaveBeenCalledTimes(1);
	});
});
