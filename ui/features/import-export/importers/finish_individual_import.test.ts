// The tail three importers share. Two things here are behaviour rather than plumbing: the class
// guard, which now reaches the caller instead of becoming an unhandled rejection, and which of the
// two closing toasts is shown.
import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { Class, EquipmentSpec, Profession, Race } from '@generated/proto/common';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { fakeHost } from '@testing/sim/fake_host';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { finishIndividualImport } from './finish_individual_import';

const loadLeftovers = vi.hoisted(() => vi.fn().mockResolvedValue(undefined));
vi.mock('@sim/proto/database', () => ({ Database: { loadLeftoversIfNecessary: loadLeftovers } }));

// The real one gates store notifications; here it only has to run the body.
vi.mock('@sim/state/batch', () => ({ batch: (body: () => void) => body() }));

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));

const levelEdits = vi.hoisted(() => [] as Array<number>);
vi.mock('@features/settings/model/level_defaults', () => ({
	applyLevelEdit: async (_host: unknown, level: number) => {
		levelEdits.push(level);
	},
}));

const fromProto = vi.fn();

const player = {
	getClass: () => Class.ClassWarrior,
	getPlayerClass: () => ({ friendlyName: 'Warrior' }),
	setRace: vi.fn(),
	setGear: vi.fn(),
	setTalentsString: vi.fn(),
	setProfessions: vi.fn(),
};
const host = fakeHost({ player, sim: { db: { lookupEquipmentSpec: vi.fn(() => 'the gear') } }, fromProto });

const parsed = (overrides: Partial<Parameters<typeof finishIndividualImport>[1]> = {}) => ({
	charClass: Class.ClassWarrior,
	race: Race.RaceHuman,
	equipmentSpec: EquipmentSpec.create(),
	talentsStr: '05002001-0550000502-05032',
	professions: [],
	...overrides,
});

describe('finishIndividualImport', () => {
	beforeEach(() => {
		toasts.length = 0;
		levelEdits.length = 0;
		for (const setter of [player.setRace, player.setGear, player.setTalentsString, player.setProfessions, fromProto, loadLeftovers]) setter.mockClear();
	});

	it('rejects when the export is for another class, before touching the player', async () => {
		await expect(finishIndividualImport(host, parsed({ charClass: Class.ClassMage }))).rejects.toThrow('import.errors.wrong_class');
		expect(player.setRace).not.toHaveBeenCalled();
		expect(loadLeftovers).not.toHaveBeenCalled();
	});

	// The level's own talents, gear and EP load first, so the export's gear lands over them.
	it("sets the export's level the way a user's edit does, before the gear", async () => {
		player.setGear.mockImplementation(() => expect(levelEdits).toEqual([24]));
		await finishIndividualImport(host, parsed({ level: 24, talentsStr: '' }));

		expect(levelEdits).toEqual([24]);
		expect(player.setGear).toHaveBeenCalledWith('the gear');
		player.setGear.mockReset();
	});

	it('leaves the level alone for an export that names none', async () => {
		await finishIndividualImport(host, parsed({ talentsStr: '' }));
		expect(levelEdits).toEqual([]);
	});

	it('applies race, gear and professions, and reports success', async () => {
		await finishIndividualImport(host, parsed({ professions: [Profession.Engineering], talentsStr: '' }));

		expect(player.setRace).toHaveBeenCalledWith(Race.RaceHuman);
		expect(player.setGear).toHaveBeenCalledWith('the gear');
		expect(player.setProfessions).toHaveBeenCalledWith([Profession.Engineering]);
		expect(toasts).toEqual([{ variant: 'success', body: 'import.success' }]);
	});

	// A talent string indexes the exporter's trees by position, so applying one built for other
	// trees would spend the points on different talents.
	it('does not apply a talent string unless it is marked as built from these trees, and says so', async () => {
		await finishIndividualImport(host, parsed({}));

		expect(player.setTalentsString).not.toHaveBeenCalled();
		expect(toasts[0].variant).toBe('warning');
		expect(toasts[0].body).toBe('import.talents_not_imported');
	});

	it('applies a talent string built from these trees, without a warning', async () => {
		await finishIndividualImport(host, parsed({ talentsMatchTrees: true }));

		expect(player.setTalentsString).toHaveBeenCalledWith('05002001-0550000502-05032');
		expect(toasts).toEqual([{ variant: 'success', body: 'import.success' }]);
	});

	it('sets nothing for a marked export that carried no talents', async () => {
		await finishIndividualImport(host, parsed({ talentsMatchTrees: true, talentsStr: '--' }));

		expect(player.setTalentsString).not.toHaveBeenCalled();
		expect(toasts.every(toast => toast.variant !== 'warning')).toBe(true);
	});

	// `--` is what an empty three-tree talent string looks like: the import carried no talents, so
	// there is nothing to warn about either.
	it('says nothing when the export carried no talents at all', async () => {
		await finishIndividualImport(host, parsed({ talentsStr: '--' }));

		expect(player.setTalentsString).not.toHaveBeenCalled();
		expect(toasts.every(toast => toast.variant !== 'warning')).toBe(true);
	});

	it('lists what the database did not have', async () => {
		await finishIndividualImport(host, parsed({ missingItems: [1, 2], missingEnchants: [3], talentsStr: '' }));
		expect(toasts).toEqual([
			{
				variant: 'info',
				body: 'import.missing_ids.title\n\nimport.missing_ids.items 1, 2\n\nimport.missing_ids.enchants 3',
			},
		]);
	});

	it('applies only the sections it is given', async () => {
		await finishIndividualImport(host, { charClass: Class.ClassWarrior, race: Race.RaceHuman });

		expect(player.setRace).toHaveBeenCalledWith(Race.RaceHuman);
		expect(player.setGear).not.toHaveBeenCalled();
		expect(player.setProfessions).not.toHaveBeenCalled();
		expect(fromProto).not.toHaveBeenCalled();
		expect(loadLeftovers).not.toHaveBeenCalled();
		expect(toasts).toEqual([{ variant: 'success', body: 'import.success' }]);
	});

	it('hands a rotation to the host as the rotation category only', async () => {
		const rotation = APLRotation.create({ type: APLRotation_Type.TypeAPL });
		await finishIndividualImport(host, { charClass: Class.ClassWarrior, rotation });

		expect(fromProto).toHaveBeenCalledTimes(1);
		expect(fromProto.mock.calls[0][0].player.rotation).toEqual(rotation);
		expect(fromProto.mock.calls[0][1]).toEqual([SimSettingCategories.Rotation]);
	});
});
