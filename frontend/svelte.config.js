import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	compilerOptions: {
		runes: true
	},
	kit: {
		// adapter-static configuration for SPA mode
		adapter: adapter({
			fallback: 'index.html', // Enables SPA mode
			strict: false
		}),
		csrf: {
			trustedOrigins: []
		},
		alias: {
			'@/*': './path/to/lib/*'
		}
	}
};

export default config;
