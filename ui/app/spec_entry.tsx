// Generic page entry for every spec site. A spec page is `<base>/<class>/<spec>/` (`/forever/` by default, tools/vite/site_base.mts), which mirrors the folder tree (vite's root is `ui/`), so the module to load is derivable from the URL — no per-spec `index.ts`.
import '../shared/page_boot';

import { browserEnv } from '@app/browser_env';
import { SITE_HOME_PATH } from '@sim/constants/other';
import { Player } from '@sim/player/player';
import { shownSpecs } from '@sim/player/player_class';
import { isSpecShown } from '@sim/player/player_spec';
import { PlayerSpecs } from '@sim/player/specs';
import { Sim } from '@sim/sim';
import type { SpecDefinition } from '@sim/spec_config';
import { registerSpecConfig, withoutDevOnlyGear } from '@sim/spec_config';
import { isPreviewOrDevMode } from '@sim/utils/env';
import { StrictMode } from 'react';
import { flushSync } from 'react-dom';
import { createRoot } from 'react-dom/client';

import { SimApp } from './SimApp';

const modules = import.meta.glob<{ default: SpecDefinition<any> }>('../specs/*/*/spec.{ts,tsx}');

// '<base>warrior/dps/' -> '../specs/warrior/dps/spec' (then tried as .ts and .tsx — no spec
// carries real JSX today, but the glob accepts one that starts to).
const specModuleKey = (pathname: string): string => {
	const base = import.meta.env.BASE_URL || '/';
	const rel = (pathname.startsWith(base) ? pathname.slice(base.length) : pathname)
		.replace(/^\/+/, '')
		.replace(/index\.html$/, '')
		.replace(/\/+$/, '');
	return `../specs/${rel}/spec`;
};

// An async IIFE rather than top-level await: the vite build target does not support TLA and downgrades it to a tolerated transform.
void (async () => {
	const key = specModuleKey(location.pathname);
	const loadSpec = modules[`${key}.ts`] || modules[`${key}.tsx`];
	if (!loadSpec) {
		throw new Error(`No spec module for ${location.pathname} (looked for ${key}.ts(x)).`);
	}

	const loaded = (await loadSpec()).default;
	const playerSpec = PlayerSpecs.fromProto(loaded.spec);
	// A hidden spec (isSpecShown) still has its page in the build, since the pages come from ui/specs; it sends the
	// visitor to the class's first shown spec, or to the landing page when the class has none.
	if (!isSpecShown(playerSpec)) {
		location.replace(shownSpecs(PlayerSpecs.getPlayerClass(playerSpec))[0]?.simLink ?? SITE_HOME_PATH);
		return;
	}
	const def = isPreviewOrDevMode(location.hostname) ? loaded : withoutDevOnlyGear(loaded);

	// `new Player()` resolves the spec's config out of the registry in its constructor, so the definition has to be registered before the player is built.
	registerSpecConfig(def.spec, def);

	const sim = new Sim({ env: browserEnv });
	sim.runs.setFlush(flushSync);
	const player = new Player(playerSpec, sim);
	// Opt-in per spec, not inferred from isTankSpec || isHealingSpec: neither restoration
	// spec enables it, so that heuristic would hand them a computed healing model.
	if (def.enableHealing) player.enableHealing();

	sim.raid.setPlayer(0, player);

	const rootElem = document.getElementById('root');
	if (!rootElem) throw new Error('No #root element on the page; ui/index_template.html should provide it.');

	createRoot(rootElem).render(
		<StrictMode>
			<SimApp player={player} def={def} />
		</StrictMode>,
	);
})();
