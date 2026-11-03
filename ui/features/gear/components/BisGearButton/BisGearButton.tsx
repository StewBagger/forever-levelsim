import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { Button } from '@ui-kit/Button';
import { toastManager } from '@ui-kit/Toast';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId, useState } from 'react';

export interface BisGearButtonViewProps {
	level: number;
	busy: boolean;
	onClick: () => void;
}

export const BisGearButtonView = ({ level, busy, onClick }: BisGearButtonViewProps) => {
	const tooltipId = useId();
	return (
		<>
			<Button
				variant="outline-primary"
				className="mb-4 w-full"
				data-testid="bis-gear-button"
				loading={busy}
				disabled={busy}
				onClick={onClick}
				{...tooltipAnchorProps(tooltipId)}>
				{i18n.t('gear_tab.bis_gear.label', { level })}
			</Button>
			<Tooltip id={tooltipId} content={i18n.t('gear_tab.bis_gear.tooltip', { level })} />
		</>
	);
};

// Equips the sim's best gear in slot for the character's level without raid gear, valued by the player's EP weights.
export const BisGearButton = () => {
	const { player, sim } = useSimHost();
	const level = usePlayerStore('level');
	const [busy, setBusy] = useState(false);

	const equip = async () => {
		setBusy(true);
		try {
			const equipment = await sim.bisGear(player, level, player.getEpWeights());
			player.setGear(sim.db.lookupEquipmentSpec(equipment));
		} catch (error) {
			toastManager.add({ variant: 'error', body: error instanceof Error ? error.message : String(error) });
		} finally {
			setBusy(false);
		}
	};

	return <BisGearButtonView level={level} busy={busy} onClick={() => void equip()} />;
};
