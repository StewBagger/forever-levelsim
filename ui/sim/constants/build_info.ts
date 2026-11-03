import type { BuildInfo, SpecRelease } from '../../../tools/vite/build_info.mjs';
import { ISSUES_REPO } from './other';

export type { Change } from '../../../tools/vite/build_info.mjs';

// The release this page was built as. vite.config.mts defines it at build time (tools/vite/build_info.mts,
// which owns the types).
export const BUILD_INFO: BuildInfo = import.meta.env.VITE_BUILD_INFO ?? { date: '', commit: '', gameVersion: '', specs: {} };

/** The release of the spec whose page is at `simLink`, which ends in `/<class>/<spec>/`. */
export const specRelease = (simLink: string): SpecRelease => Object.entries(BUILD_INFO.specs).find(([key]) => simLink.endsWith(`/${key}/`))?.[1] ?? {};

// The repository the issue links name, so a fork's build links its own commit.
export const commitUrl = (commit: string): string => `https://github.com/${ISSUES_REPO}/commit/${commit}`;
