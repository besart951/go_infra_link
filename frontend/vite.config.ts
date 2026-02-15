import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	optimizeDeps: {
		include: ['lucide-svelte', 'bits-ui', 'clsx', 'tailwind-merge', 'tailwind-variants']
	}
});
