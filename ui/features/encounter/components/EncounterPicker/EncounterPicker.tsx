import { applyHealingGroup } from '@features/encounter/model/healing_group';
import { ConnectedBooleanPicker } from '@features/inputs/components/ConnectedBooleanPicker';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { Button } from '@ui-kit/Button';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { type ReactNode, useMemo, useState } from 'react';

import { AdvancedEncounterModal } from '../AdvancedEncounterModal';
import { TargetInputsPicker } from '../TargetsPicker';
import { AreaTypesPicker } from './AreaTypesPicker';
import { durationConfigs, executeConfigs, minBaseDamageConfig, numAlliesConfig, presetEncounterConfig, zoneConfig } from './utils/configs';
import { damageProfileNumberConfigs, groupHitsTankConfig, groupSchoolConfig } from './utils/healing_group_configs';

export interface EncounterPickerViewProps {
	duration: ReactNode;
	execute?: ReactNode;
	zone: ReactNode;
	areaTypes: ReactNode;
	preset?: ReactNode;
	allies?: ReactNode;
	minBaseDamage?: ReactNode;
	// A healer's group and the damage it takes, with the button that resets them to the level's.
	healingGroup?: ReactNode;
	onHealingGroupReset?: () => void;
	targetInputs: ReactNode;
	onAdvancedClick: () => void;
}

export const EncounterPickerView = ({
	duration,
	execute,
	zone,
	areaTypes,
	preset,
	allies,
	minBaseDamage,
	healingGroup,
	onHealingGroupReset,
	targetInputs,
	onAdvancedClick,
}: EncounterPickerViewProps) => (
	<div className="flex flex-wrap gap-2" data-testid="encounter-picker-root">
		<FieldGrid>{duration}</FieldGrid>
		{execute && <FieldGrid className="w-full">{execute}</FieldGrid>}
		{zone}
		{areaTypes}
		{preset}
		{allies}
		{minBaseDamage}
		{healingGroup && (
			<fieldset className="flex w-full flex-col gap-2" data-testid="healing-group">
				<legend className="mb-1 font-bold">{i18n.t('settings_tab.encounter.healing_group.title')}</legend>
				<FieldGrid>{healingGroup}</FieldGrid>
				{onHealingGroupReset && (
					<Button className="w-full" data-testid="healing-group-reset" onClick={onHealingGroupReset}>
						{i18n.t('settings_tab.encounter.healing_group.level_default')}
					</Button>
				)}
			</fieldset>
		)}
		{targetInputs}
		<Button className="w-full" data-testid="advanced-button" onClick={onAdvancedClick}>
			{i18n.t('settings_tab.encounter.advanced')}
		</Button>
	</div>
);

export interface EncounterPickerProps {
	showExecuteProportion: boolean;
}

export const EncounterPicker = ({ showExecuteProportion }: EncounterPickerProps) => {
	const host = useSimHost();
	const encounter = host.sim.encounter;
	const player = host.player;
	const { damage: showDamage } = useDisplayMetrics(host.sim);

	const duration = useMemo(() => durationConfigs(encounter), [encounter]);
	const execute = useMemo(() => (showExecuteProportion ? executeConfigs(encounter) : []), [encounter, showExecuteProportion]);
	const preset = useMemo(() => presetEncounterConfig(encounter), [encounter]);
	const allies = useMemo(() => numAlliesConfig(player), [player]);
	const minBaseDamage = useMemo(() => minBaseDamageConfig(), []);
	const zone = useMemo(() => zoneConfig(encounter), [encounter]);
	const healer = player.getPlayerSpec().isHealingSpec;
	const damageNumbers = useMemo(() => damageProfileNumberConfigs(), []);
	const groupSchool = useMemo(() => groupSchoolConfig(), []);
	const groupHitsTank = useMemo(() => groupHitsTankConfig(), []);

	const [advancedOpen, setAdvancedOpen] = useState(false);

	return (
		<>
			<EncounterPickerView
				duration={duration.map(config => (
					<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
				))}
				execute={showExecuteProportion && execute.map(config => <ConnectedNumberField key={config.id} modObject={encounter} config={config} />)}
				zone={<ConnectedChoiceField modObject={encounter} config={zone} />}
				areaTypes={<AreaTypesPicker encounter={encounter} />}
				preset={showDamage && <ConnectedChoiceField modObject={encounter} config={preset} />}
				allies={player.canEnableTargetDummies() && <ConnectedNumberField modObject={host.sim.raid} config={allies} />}
				minBaseDamage={player.getPlayerSpec().isTankSpec && <ConnectedNumberField modObject={encounter} config={minBaseDamage} />}
				healingGroup={
					healer && (
						<>
							{damageNumbers.map(config => (
								<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
							))}
							<ConnectedChoiceField modObject={encounter} config={groupSchool} />
							<ConnectedBooleanPicker modObject={encounter} config={groupHitsTank} />
						</>
					)
				}
				onHealingGroupReset={healer ? () => applyHealingGroup(host.sim, player.getLevel()) : undefined}
				targetInputs={<TargetInputsPicker encounter={encounter} targetIndex={0} />}
				onAdvancedClick={() => setAdvancedOpen(true)}
			/>
			<AdvancedEncounterModal open={advancedOpen} onOpenChange={setAdvancedOpen} />
		</>
	);
};
