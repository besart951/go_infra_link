<script lang="ts">
	import { onMount } from 'svelte';
	import './layout.css';
	import { initTheme } from '$lib/stores/theme.js';

	const { children } = $props();

	onMount(() => {
		initTheme();
	});
</script>

{@render children()}
