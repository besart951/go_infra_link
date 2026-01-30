<script lang="ts">
	import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
	import { getApparat, listApparats } from '$lib/infrastructure/api/facility.adapter.js';
	import type { Apparat } from '$lib/domain/facility/index.js';

	export let value: string[] = [];
	export let width: string = 'w-full';
	export let disabled: boolean = false;
	export let id: string | undefined = undefined;

	type ApparatOption = Apparat & { label: string };

	function toOption(apparat: Apparat): ApparatOption {
		return {
			...apparat,
			label: `${apparat.short_name} — ${apparat.name}`
		};
	}

	async function fetcher(search: string): Promise<ApparatOption[]> {
		const res = await listApparats({ search, limit: 50 });
		return (res.items || []).map(toOption);
	}

	async function fetchByIds(ids: string[]): Promise<ApparatOption[]> {
		const promises = ids.map((id) => getApparat(id));
		const results = await Promise.allSettled(promises);
		return results
			.filter((r): r is PromiseFulfilledResult<Apparat> => r.status === 'fulfilled')
			.map((r) => toOption(r.value));
	}
</script>

<AsyncMultiSelect
	bind:value
	{fetcher}
	{fetchByIds}
	labelKey="label"
	placeholder="Select Apparats..."
	searchPlaceholder="Search apparats..."
	emptyText="No apparats found."
	{width}
	{disabled}
	{id}
/>
