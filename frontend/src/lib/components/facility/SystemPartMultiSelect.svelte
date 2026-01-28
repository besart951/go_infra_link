<script lang="ts">
	import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
	import { getSystemPart, listSystemParts } from '$lib/infrastructure/api/facility.adapter.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';

	export let value: string[] = [];
	export let width: string = 'w-full';
	export let disabled: boolean = false;
	export let id: string | undefined = undefined;

	async function fetcher(search: string): Promise<SystemPart[]> {
		const res = await listSystemParts({ search, limit: 50 });
		return res.items || [];
	}

	async function fetchByIds(ids: string[]): Promise<SystemPart[]> {
		// Fetch each system part by ID
		const promises = ids.map((id) => getSystemPart(id));
		const results = await Promise.allSettled(promises);
		return results
			.filter((r): r is PromiseFulfilledResult<SystemPart> => r.status === 'fulfilled')
			.map((r) => r.value);
	}
</script>

<AsyncMultiSelect
	bind:value
	{fetcher}
	{fetchByIds}
	labelKey="name"
	placeholder="Select System Parts..."
	searchPlaceholder="Search system parts..."
	emptyText="No system parts found."
	{width}
	{disabled}
	{id}
/>
