import i18n from '@i18n/config';
import { NULL_SIMS_REPO_URL } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';

export const LandingActions = () => (
	<div className="flex flex-wrap gap-4" data-testid="landing-actions">
		<Button href={NULL_SIMS_REPO_URL} target="_blank" variant="unstyled" title={i18n.t('landing.actions.github')} className="ui-landing-action">
			<Icon name="github" style="brands" />
			{i18n.t('landing.actions.github')}
		</Button>
	</div>
);
