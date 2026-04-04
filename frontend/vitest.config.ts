import { mergeConfig, defineConfig } from 'vitest/config';
import viteConfig from './vite.config';

const vitestOverrides = defineConfig({
  resolve: {
    conditions: ['browser']
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    globals: true,
    include: ['src/**/*.test.ts']
  }
});

export default defineConfig(async ({ mode }) => {
  const resolvedViteConfig =
    typeof viteConfig === 'function'
      ? await viteConfig({
          command: 'serve',
          mode,
          isSsrBuild: false,
          isPreview: false
        })
      : viteConfig;

  return mergeConfig(resolvedViteConfig, vitestOverrides);
});
