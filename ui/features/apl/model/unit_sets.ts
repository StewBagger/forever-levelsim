import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { Player } from '@sim/player/player';

export type UNIT_SET = 'aura_sources' | 'aura_sources_targets_first' | 'self_and_pets' | 'targets' | 'players';

const ownPetRefs = (player: Player<any>) =>
	player
		.getPetMetadatas()
		.asList()
		.map((_petMetadata, i) => UnitReference.create({ type: UnitType.Pet, index: i, owner: UnitReference.create({ type: UnitType.Self }) }));

// The group members a healer picks by health, read at each use: the lowest in the raid, the lowest in its party, and
// the tank. A healing spec's group of healable target dummies is what gives them a unit to name.
const healTargetRefs = (player: Player<any>) =>
	player.shouldEnableTargetDummies() && player.getPlayerSpec().isHealingSpec
		? [UnitType.LowestHealthAlly, UnitType.LowestHealthPartyMember, UnitType.Tank].map(type => UnitReference.create({ type }))
		: [];

export const unitSets: Record<
	UNIT_SET,
	{
		// Uses target icon by default instead of person icon. This should be set to true for inputs that default to CurrentTarget.
		targetUI?: boolean;
		getUnits: (player: Player<any>) => Array<UnitReference | undefined>;
	}
> = {
	aura_sources: {
		getUnits: player => {
			return [
				undefined,
				ownPetRefs(player),
				UnitReference.create({ type: UnitType.CurrentTarget }),
				UnitReference.create({ type: UnitType.PreviousTarget }),
				UnitReference.create({ type: UnitType.NextTarget }),
				healTargetRefs(player),
				player.sim.raid
					.getActivePlayers()
					.filter(filter => filter != player)
					.map(mapPlayer => UnitReference.create({ type: UnitType.Player, index: mapPlayer.getRaidIndex() })),
				player.sim.encounter.targetsMetadata.asList().map((targetMetadata, i) => UnitReference.create({ type: UnitType.Target, index: i })),
			].flat();
		},
	},
	aura_sources_targets_first: {
		targetUI: true,
		getUnits: player => {
			return [
				undefined,
				player.sim.encounter.targetsMetadata.asList().map((targetMetadata, i) => UnitReference.create({ type: UnitType.Target, index: i })),
				UnitReference.create({ type: UnitType.Self }),
				ownPetRefs(player),
			].flat();
		},
	},
	// Lists every own pet. The sim drops a value read from a pet without that resource, so a bare action condition counts as always true and an And or Or loses that branch.
	self_and_pets: {
		getUnits: player => {
			return [undefined, ownPetRefs(player)].flat();
		},
	},
	targets: {
		targetUI: true,
		getUnits: player => {
			return [
				undefined,
				player.sim.encounter.targetsMetadata.asList().map((_targetMetadata, i) => UnitReference.create({ type: UnitType.Target, index: i })),
				UnitReference.create({ type: UnitType.PreviousTarget }),
				UnitReference.create({ type: UnitType.NextTarget }),
			].flat();
		},
	},
	players: {
		targetUI: true,
		getUnits: player => {
			return [
				undefined,
				healTargetRefs(player),
				player.sim.raid.getActivePlayers().map(other => UnitReference.create({ type: UnitType.Player, index: other.getRaidIndex() })),
			].flat();
		},
	},
};
