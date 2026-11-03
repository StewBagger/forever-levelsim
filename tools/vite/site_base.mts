// Where the site is served from, for every config that builds or renders the UI (the site, its workers, the unit
// tests and Storybook).
//
// `SIM_BASE` sets the path, and nothing else names it: `SIM_BASE=/sims/ make site` builds a site to serve at
// `/sims/`. The default is `/forever/`, the path the Go server (`./wowsimforever`) serves its embedded copy at.
import type { Plugin } from 'vite';

const normaliseBase = (base: string) => `/${base.replace(/^\/+|\/+$/g, '')}/`.replace(/^\/\/$/, '/');

/** The path the site is served under, with a leading and a trailing slash. */
export const SITE_BASE = normaliseBase(process.env.SIM_BASE ?? '/forever/');

/** The public home of Null Sims. Canonical, `og:url` and the spec pages' links name it, wherever a build is served. */
export const SITE_URL = 'https://nossuary.com/sims/';

/** `define` entries that hand the path to the UI as `import.meta.env.VITE_SITE_BASE` (`SITE_BASE` in `ui/sim/constants/other.ts`). */
export const siteBaseDefine = { 'import.meta.env.VITE_SITE_BASE': JSON.stringify(SITE_BASE) };

// The stylesheets name their fonts and images root-absolute under the default base, `url(/forever/assets/...)`:
// vite leaves a url it cannot resolve to a file alone, so the path has to be right as written. This moves that
// prefix to the configured base. It runs after vite's own CSS transform, on the stylesheet tailwind has already
// inlined its imports into.
const SOURCE_BASE = '/forever/';
const SOURCE_URL = /url\((['"]?)\/forever\//g;

export function siteBaseCss(): Plugin {
	return {
		name: 'site-base-css',
		transform(code, id) {
			if (SITE_BASE === SOURCE_BASE || !/\.css(?:$|\?)/.test(id) || !code.includes(SOURCE_BASE)) return null;
			return { code: code.replace(SOURCE_URL, (_, quote) => `url(${quote}${SITE_BASE}`), map: null };
		},
	};
}
