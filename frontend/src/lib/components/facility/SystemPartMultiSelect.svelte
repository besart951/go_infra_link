<script lang="ts">
	import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
	import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';

	export let value: string[] = [];
	export let width: string = 'w-full';
	export let disabled: boolean = false;
	export let id: string | undefined = undefined;

	async function fetcher(search: string): Promise<SystemPart[]> {
		const res = await systemPartRepository.list({
			pagination: { page: 1, pageSize: 50 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchByIds(ids: string[]): Promise<SystemPart[]> {
		const promises = ids.map((id) => systemPartRepository.get(id));
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
