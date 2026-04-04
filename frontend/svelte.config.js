import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  compilerOptions: {
    runes: true
  },
  kit: {
    adapter: adapter({
      // Production is a static SPA served by Caddy. Deep links fall back to index.html,
      // while the edge reverse proxy keeps /api/* on the same origin.
      fallback: 'index.html',
      strict: false
    })
  }
};

export default config;
