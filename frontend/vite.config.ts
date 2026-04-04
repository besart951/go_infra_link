import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import inspect from 'vite-plugin-inspect';
import { defineConfig, loadEnv } from 'vite';

type ProxyListenerHost = {
  on: (event: string, listener: (...args: any[]) => void) => void;
};

type ProxyReqLike = {
  getHeader: (name: string) => unknown;
};

type ReqLike = {
  method?: string;
  url?: string;
};

type ProxyResLike = {
  statusCode?: number;
};

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '');
  const backendPort = env.BACKEND_PORT ?? '8080';
  const backendUrl = env.BACKEND_URL ?? `http://localhost:${backendPort}`;

  return {
    plugins: [
      tailwindcss(),
      sveltekit(),
      inspect({
        build: false,
        outputDir: '.vite-inspect'
      })
    ],
    test: {
      environment: 'jsdom',
      setupFiles: ['./vitest.setup.ts'],
      globals: true,
      include: ['src/**/*.test.ts']
    },
    server: {
      proxy: {
        '/api': {
          target: backendUrl,
          changeOrigin: true,
          secure: false,
          // Dev-only parity with the production edge proxy. The production build is a static SPA.
          configure: (proxy) => {
            const proxyWithEvents = proxy as unknown as ProxyListenerHost;

            proxyWithEvents.on('error', (err: Error) => {
              console.error('[proxy] error', err);
            });
            proxyWithEvents.on('proxyReq', (proxyReq: ProxyReqLike, req: ReqLike) => {
              console.log(`[proxy] ${req.method} ${req.url}`);
              if (proxyReq.getHeader('origin')) {
                console.log(`[proxy] origin ${proxyReq.getHeader('origin')}`);
              }
            });
            proxyWithEvents.on('proxyRes', (proxyRes: ProxyResLike, req: ReqLike) => {
              console.log(`[proxy] ${req.method} ${req.url} -> ${proxyRes.statusCode}`);
            });
          }
        }
      }
    },
    build: {
      sourcemap: true
    }
  };
});
