import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import inspect from 'vite-plugin-inspect';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit(), inspect()],
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true,
				secure: false,
				configure: (proxy) => {
					proxy.on('error', (err) => {
						console.error('[proxy] error', err);
					});
					proxy.on('proxyReq', (proxyReq, req) => {
						console.log(`[proxy] ${req.method} ${req.url}`);
						if (proxyReq.getHeader('origin')) {
							console.log(`[proxy] origin ${proxyReq.getHeader('origin')}`);
						}
					});
					proxy.on('proxyRes', (proxyRes, req) => {
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
