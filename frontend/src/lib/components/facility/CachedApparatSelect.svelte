<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { lookupCache } from '$lib/stores/facility/lookupCache.js';
	import type { Apparat } from '$lib/domain/facility/index.js';

	interface Props {
		value?: string;
		width?: string;
		onValueChange?: (value: string) => void;
		disabled?: boolean;
	}

	let {
		value = $bindable(''),
		width = 'w-full',
		onValueChange,
		disabled = false
	}: Props = $props();

	async function fetcher(search: string): Promise<Apparat[]> {
		return lookupCache.fetchApparats(search);
	}

	async function fetchById(id: string): Promise<Apparat | null> {
		return lookupCache.fetchApparatById(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder="Select Apparat..."
	{width}
	{onValueChange}
	{disabled}
/>
