<script lang="ts">
	import { onMount } from 'svelte';
	import './layout.css';
	import { initTheme } from '$lib/stores/theme.js';
	import { i18n } from '$lib/i18n/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	const { children } = $props();
	const t = createTranslator();

	let translationsReady = $state(false);
	let translationError = $state<string | null>(null);

	onMount(() => {
		initTheme();

		// Subscribe to i18n store to know when translations are loaded
		const unsubscribe = i18n.subscribe((state) => {
			translationsReady = !state.isLoading;
			translationError = state.error;
		});

		return unsubscribe;
	});
</script>

{#if !translationsReady && !translationError}
	<div class="flex h-screen items-center justify-center">
		<div class="text-center">
			<div
				class="mb-4 inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]"
			></div>
			<p class="text-muted-foreground">{$t('app.loading')}</p>
		</div>
	</div>
{:else if translationError}
	<div class="flex h-screen items-center justify-center">
		<div class="text-center text-red-600">
			<p class="mb-2 text-lg font-semibold">{$t('app.translation_error_title')}</p>
			<p class="text-sm">{translationError}</p>
			<button
				onclick={() => i18n.reload()}
				class="mt-4 rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
			>
				{$t('app.retry')}
			</button>
		</div>
	</div>
{:else}
	{@render children()}
{/if}
