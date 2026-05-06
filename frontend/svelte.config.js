import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  compilerOptions: {
    runes: true
  },
  kit: {
    output: {
      // Production favors the lowest possible JS request count. This emits one
      // cacheable app bundle instead of route/chunk waterfalls while keeping
      // assets outside HTML for normal browser caching.
      bundleStrategy: 'single'
    },
    adapter: adapter({
      // Production is a static SPA served by Caddy. Deep links fall back to index.html,
      // while the edge reverse proxy keeps /api/* on the same origin.
      fallback: 'index.html',
      strict: false
    })
  }
};

export default config;
