import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import inspect from 'vite-plugin-inspect';
import { defineConfig } from 'vite';

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

export default defineConfig({
	plugins: [tailwindcss(), sveltekit(), inspect()],
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true,
				secure: false,
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
});
