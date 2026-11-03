import { ProtoVersion } from '@generated/proto/common';
import { readMessageOption } from '@protobuf-ts/runtime';

export enum Phase {
	Phase1 = 1,
	Phase2,
	Phase3,
	Phase4,
	Phase5,
}

export const CURRENT_PHASE = Phase.Phase1;

export enum LaunchStatus {
	Unlaunched,
	// Gear, enchants, talents and the gear optimizer work. No simulation, and none planned.
	GearPlanner,
	Alpha,
	Beta,
	Launched,
}

export const CURRENT_API_VERSION: number = readMessageOption(ProtoVersion, 'proto.current_version_number')! as number;

// The path the site is served under, with a leading and a trailing slash: '/forever/' by default, which the Go server
// serves its copy at, and whatever SIM_BASE names for a static build (tools/vite/site_base.mts). Every root-absolute
// url the UI builds starts with it.
export const SITE_BASE: string = import.meta.env.VITE_SITE_BASE;
// The public home of Null Sims, and the site of the Null Ossuary project it belongs to.
export const NULL_SIMS_URL = 'https://nossuary.com/sims/';
export const NOSSUARY_URL = 'https://nossuary.com/';

// Null Sims is a fork of WoWSims Forever that adds levelling sims; its own links point at the fork, and the
// WOWSIMS_* links credit the upstream project.
const NULL_SIMS_REPO = 'StewBagger/forever-levelsim';
export const NULL_SIMS_REPO_URL = `https://github.com/${NULL_SIMS_REPO}`;
export const NULL_SIMS_ISSUES_URL = `${NULL_SIMS_REPO_URL}/issues`;
// The repository whose CI built this page (see vite.config.mts), so a fork's build files its issues on
// the fork. A build made outside GitHub Actions, the server's included, files them on Null Sims.
export const ISSUES_REPO: string = import.meta.env.VITE_GITHUB_REPOSITORY || NULL_SIMS_REPO;
export const REPO_NEW_ISSUE_URL = `https://github.com/${ISSUES_REPO}/issues/new`;
export const REPO_CHOOSE_NEW_ISSUE_URL = `${REPO_NEW_ISSUE_URL}/choose`;

// WoWSims, the MIT-licensed project Null Sims is built on.
export const WOWSIMS_URL = 'https://www.wowsims.com/';
export const WOWSIMS_REPO_URL = 'https://github.com/wowsims/forever';
// The WoWSims Exporter, the WoWSims in-game addon the addon importer reads.
export const ADDON_URL = 'https://www.curseforge.com/wow/addons/wowsimsexporter';
export const SITE_HOME_PATH = SITE_BASE;

export interface Social {
	key: string;
	href: string;
	className: string;
	icon: 'github';
	tooltip: 'info.github';
	label?: string;
}

export const SOCIALS: ReadonlyArray<Social> = [
	{ key: 'github', href: NULL_SIMS_REPO_URL, className: 'ui-social-link', icon: 'github', tooltip: 'info.github' },
];

// Root-relative path of the individual sim page for the given spec. Resolve it
// against the page origin at the point of use (see SimTitleDropdown) — this
// layer has no `window`. Lives here rather than in proto/utils so that
// player/specs/<class>.ts (which calls it at module scope) does not import the
// proto <-> player/specs cycle; the evaluation order of that cycle decides
// whether PlayerSpecs' lookup table is populated.
export function getSpecSitePath(classString: string, specString: string): string {
	return `${SITE_BASE}${classString}/${specString}/`;
}

export const LOCAL_STORAGE_PREFIX = '__forever';

export enum SortDirection {
	ASC,
	DESC,
}
