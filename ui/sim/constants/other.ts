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

// Github pages serves our site under the /forever directory
export const REPO_NAME = 'forever';
export const REPO_URL = `https://github.com/wowsims/${REPO_NAME}`;
export const REPO_RELEASES_URL = `${REPO_URL}/releases`;
// The repository whose CI built this page (see vite.config.mts). Empty for a build made outside
// GitHub Actions, which then offers no issue links at all.
export const ISSUES_REPO: string = import.meta.env.VITE_GITHUB_REPOSITORY;
export const REPO_NEW_ISSUE_URL = `https://github.com/${ISSUES_REPO}/issues/new`;
export const REPO_CHOOSE_NEW_ISSUE_URL = `${REPO_NEW_ISSUE_URL}/choose`;

export const WOWSIMS_URL = 'https://www.wowsims.com/';
// The site list and footer content every WoWSims site shares, published by wowsims/wowsims.github.io
export const VERSIONS_MANIFEST_URL = `${WOWSIMS_URL}versions.json`;
export const ADDON_URL = 'https://www.curseforge.com/wow/addons/wowsimsexporter';
// Null Sims is a fork of WoWSims Forever that adds levelling sims; its own links point at the fork, and
// REPO_URL above stays the upstream project it credits.
export const NULL_SIMS_REPO_URL = 'https://github.com/StewBagger/forever-levelsim';
export const NULL_SIMS_ISSUES_URL = `${NULL_SIMS_REPO_URL}/issues`;
export const SITE_HOME_PATH = `/${REPO_NAME}/`;

export interface Social {
	key: string;
	href: string;
	className: string;
	icon: 'discord' | 'github' | 'patreon';
	tooltip: 'info.discord' | 'info.github' | 'info.patreon';
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
	return `/${REPO_NAME}/${classString}/${specString}/`;
}

export const LOCAL_STORAGE_PREFIX = '__forever';

export enum SortDirection {
	ASC,
	DESC,
}
