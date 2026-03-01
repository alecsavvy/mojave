// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: 'Mojave',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/alecsavvy/mojave' }],
			sidebar: [
				{
					label: 'Overview',
					items: [
						{ label: 'Introduction', slug: 'overview/introduction' },
					],
				},
			],
		}),
	],
});
