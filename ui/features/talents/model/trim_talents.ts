// A full talent string's points spent up to a budget, the way a character levelling into that build spends them: its
// deepest tree first, then the next, each tree in the string's own order, which is row by row from the top. A tree's row
// is reached only once the rows above it are spent in full, so the trimmed string is as legal as the one it came from.
// It is TrimTalents in sim/levelsim/talents.go, which the sim trims a level's talents with; the two give the same string.
export const trimTalents = (talents: string, points: number): string => {
	const trees = talents.split('-');
	const spent = trees.map(tree => [...tree].reduce((sum, c) => sum + Number(c), 0));
	// Array.prototype.sort is stable, so trees that tie keep the string's order.
	const order = trees.map((_, i) => i).sort((a, b) => spent[b] - spent[a]);
	const out: string[] = trees.map(() => '');
	let left = points;
	for (const i of order) {
		for (const c of trees[i]) {
			const take = Math.min(Number(c), left);
			left -= take;
			out[i] += String(take);
		}
	}
	return out
		.map(tree => tree.replace(/0+$/, ''))
		.join('-')
		.replace(/-+$/, '');
};
