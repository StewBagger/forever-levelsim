// Which deployment this page is running as.

export type Environments = 'local' | 'external';

// Pure classification of a hostname. The browser probe that reads the actual
// page hostname lives in ui/ui-kit/utils/dom.ts; anything with an `Env` in hand
// (e.g. Sim) passes `env.location.hostname` here instead.
export const environmentOf = (hostname: string): Environments => (hostname.includes('localhost') ? 'local' : 'external');
export const isDevMode = () => {
	return import.meta.env.DEV;
};
// Everywhere but upstream's own site. Null Sims shows upstream's unfinished level 60 gear sets on every host it runs
// on (the dev server, the local Go server, its Cloudflare Pages previews and nossuary.com) rather than starting those
// specs with none; only wowsims.com, keyed on the hostname, keeps them hidden until upstream releases them.
const UPSTREAM_HOSTS = ['wowsims.com', 'www.wowsims.com'];

export const isPreviewOrDevMode = (hostname: string) => isDevMode() || !UPSTREAM_HOSTS.includes(hostname);
