<script lang="ts">
	import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
	import { getApparats, listApparats } from '$lib/infrastructure/api/facility.adapter.js';
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
		if (ids.length === 0) return [];
		const res = await getApparats(ids);
		return (res.items || []).map(toOption);
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
