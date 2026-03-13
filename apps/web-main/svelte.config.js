import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const fromProjectRoot = (relativePath) =>
  resolve(fileURLToPath(new URL('.', import.meta.url)), relativePath);

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
      '@ui-svelte': fromProjectRoot('../../packages/ui-svelte/src/lib'),
      '@theme': fromProjectRoot('../../packages/theme/src/lib'),
      '@i18n': fromProjectRoot('../../packages/i18n/src/lib')
    }
  }
};

export default config;
