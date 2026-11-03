import i18n from '@i18n/config';
import { Icon } from '@ui-kit/Icon';

export const UnlaunchedNotice = () => (
	<div
		className="mt-auto mr-auto mb-auto ml-auto flex max-w-100 flex-col items-center text-center [&_i]:text-danger"
		data-testid="sim-ui-unlaunched-container">
		<Icon name="ban" size="3x" className="mb-2" />
		<h6>{i18n.t('sim.unlaunched.title')}</h6>
	</div>
);
