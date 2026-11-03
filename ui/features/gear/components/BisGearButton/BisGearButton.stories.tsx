import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { BisGearButtonView } from './BisGearButton';

const meta = {
	title: 'Features/Gear/BisGearButton',
	component: BisGearButtonView,
	tags: ['autodocs'],
	args: { level: 40, busy: false, onClick: fn() },
} satisfies Meta<typeof BisGearButtonView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Busy: Story = { args: { busy: true } };
