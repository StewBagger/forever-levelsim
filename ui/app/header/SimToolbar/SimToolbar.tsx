import i18n from '@i18n/config';
import { ISSUES_REPO, REPO_CHOOSE_NEW_ISSUE_URL, SOCIALS } from '@sim/constants/other';
import { SocialLink } from '@ui-kit/SocialLink';
import type { ReactNode } from 'react';

import { ToolbarItem } from './ToolbarItem';

export interface SimToolbarViewProps {
	knownIssues: ReadonlyArray<ReactNode>;
	onOpenSettings: () => void;
}

export const SimToolbarView = ({ knownIssues, onOpenSettings }: SimToolbarViewProps) => (
	<>
		<ToolbarItem
			className="text-link-danger"
			testId="known-issues"
			hidden={knownIssues.length === 0}
			tooltip={
				<ul className="mb-0 pl-4 text-left">
					{knownIssues.map((issue, index) => (
						<li key={index}>{issue}</li>
					))}
				</ul>
			}>
			{i18n.t('info.known_issues')}
		</ToolbarItem>

		{ISSUES_REPO && <ToolbarItem href={REPO_CHOOSE_NEW_ISSUE_URL} icon="bug" tooltip={i18n.t('info.bug_report')} />}

		<ToolbarItem testId="sim-options" icon="gear" tooltip={i18n.t('info.sim_options')} onClick={onOpenSettings} />

		<div className="mt-2 mb-2 ml-4 flex border-l border-l-border [&_a]:my-2 [&_button]:my-2" data-testid="sim-toolbar-socials">
			{SOCIALS.map(social => (
				<div key={social.key} className="ml-4 flex transition-colors duration-150 ease-in-out" data-testid="sim-toolbar-item">
					<SocialLink social={social} />
				</div>
			))}
		</div>
	</>
);

export interface SimToolbarProps {
	knownIssues: ReadonlyArray<ReactNode>;
	onOpenSettings: () => void;
}

export const SimToolbar = ({ knownIssues, onOpenSettings }: SimToolbarProps) => (
	<div className="ml-auto flex flex-nowrap items-end text-(length:--text-ui) font-bold" data-testid="sim-toolbar">
		<SimToolbarView knownIssues={knownIssues} onOpenSettings={onOpenSettings} />
	</div>
);
