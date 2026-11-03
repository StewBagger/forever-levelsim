import i18n from '@i18n/config';
import { SITE_HOME_PATH } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';

/** The Null Sims emblem and name, linking to the site's home in the same tab. */
export const LandingBrand = () => (
	<Button href={SITE_HOME_PATH} variant="unstyled" className="ui-landing-brand" data-testid="landing-brand">
		<img className="ui-landing-brand-logo" src="/forever/assets/img/null-sims-emblem.webp" alt="" />
		<span className="flex flex-col leading-none">
			<span className="font-serif text-xl font-bold tracking-wider text-brand uppercase" data-testid="wowsims-title">
				{i18n.t('landing.header.wowsims')}
			</span>
			<small className="mt-[0.2rem] text-[0.7rem] font-normal text-white/65">{i18n.t('landing.header.subtitle')}</small>
		</span>
	</Button>
);
