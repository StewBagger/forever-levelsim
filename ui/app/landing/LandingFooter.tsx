import i18n from '@i18n/config';
import { NOSSUARY_URL, NULL_SIMS_ISSUES_URL, NULL_SIMS_REPO_URL, WOWSIMS_REPO_URL, WOWSIMS_URL } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { Trans } from 'react-i18next';

import { LandingBrand } from './LandingBrand';

// The year the WoWSims licence names (LICENSE), the first year of the simulator's copyright line.
const WOWSIMS_COPYRIGHT_START_YEAR = 2022;

type FooterLink = { href: string; label: string };

const NULL_SIMS_LINKS: ReadonlyArray<FooterLink> = [
	{ href: NOSSUARY_URL, label: 'landing.footer.nossuary' },
	{ href: NULL_SIMS_REPO_URL, label: 'landing.footer.source' },
	{ href: NULL_SIMS_ISSUES_URL, label: 'landing.footer.issues' },
];

const CREDIT_LINKS: ReadonlyArray<FooterLink> = [
	{ href: WOWSIMS_URL, label: 'landing.footer.wowsims' },
	{ href: WOWSIMS_REPO_URL, label: 'landing.footer.wowsims_source' },
];

const FooterGroup = ({ id, title, links }: { id: string; title: string; links: ReadonlyArray<FooterLink> }) => (
	<nav aria-labelledby={id} className="ui-landing-footer-group">
		<h2 id={id} className="ui-landing-footer-heading">
			{title}
		</h2>
		<ul className="ui-landing-footer-links">
			{links.map(({ href, label }) => (
				<li key={href}>
					<Button href={href} target="_blank" variant="unstyled" className="ui-landing-footer-link">
						{i18n.t(label)}
					</Button>
				</li>
			))}
		</ul>
	</nav>
);

/** The landing page's footer: Null Sims' own links, the credit to WoWSims it is built on, and the legal line. */
export const LandingFooter = () => (
	<footer className="ui-landing-footer">
		<div className="ui-landing-container">
			<div className="ui-landing-footer-columns">
				<div className="ui-landing-footer-brand">
					<LandingBrand />
					<p className="ui-landing-footer-about">
						<Trans
							i18n={i18n}
							i18nKey="landing.footer.about"
							components={{ wowsims: <Button href={WOWSIMS_URL} target="_blank" variant="unstyled" className="ui-landing-footer-link" /> }}
						/>
					</p>
				</div>
				<FooterGroup id="landing-footer-sims" title={i18n.t('landing.footer.sims')} links={NULL_SIMS_LINKS} />
				<FooterGroup id="landing-footer-credits" title={i18n.t('landing.footer.credits')} links={CREDIT_LINKS} />
			</div>
			<div className="ui-landing-footer-legal">
				<span>
					{i18n.t('landing.footer.copyright', {
						startYear: WOWSIMS_COPYRIGHT_START_YEAR,
						year: new Date().getFullYear(),
					})}
				</span>
				<span aria-hidden="true">·</span>
				<span>{i18n.t('landing.footer.disclaimer')}</span>
			</div>
		</div>
	</footer>
);
