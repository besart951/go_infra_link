import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		// adapter-static configuration for SPA mode
		adapter: adapter({
			fallback: 'index.html', // Enables SPA mode
			strict: false
		}),
		alias: {
			'@/*': './path/to/lib/*'
		}
	}
};

export default config;
