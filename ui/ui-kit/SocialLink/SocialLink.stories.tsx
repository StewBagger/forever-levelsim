import { type Social, SOCIALS } from '@sim/constants/other';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { SocialLink } from './SocialLink';

const [github] = SOCIALS;
const labelled: Social = { ...github, key: 'labelled', label: ' GitHub' };

const meta = {
	title: 'Kit/SocialLink',
	component: SocialLink,
	tags: ['autodocs'],
	args: { social: github },
	argTypes: { social: { control: false } },
} satisfies Meta<typeof SocialLink>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Github: Story = {};
export const WithLabel: Story = { args: { social: labelled } };
export const Row: Story = {
	render: () => (
		<div className="flex items-center gap-4">
			{SOCIALS.map(social => (
				<SocialLink key={social.key} social={social} />
			))}
		</div>
	),
};
