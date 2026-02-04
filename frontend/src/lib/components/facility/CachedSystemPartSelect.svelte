<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { lookupCache } from '$lib/stores/facility/lookupCache.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';

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

	async function fetcher(search: string): Promise<SystemPart[]> {
		return lookupCache.fetchSystemParts(search);
	}

	async function fetchById(id: string): Promise<SystemPart | null> {
		return lookupCache.fetchSystemPartById(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder="Select System Part..."
	{width}
	{onValueChange}
	{disabled}
/>
