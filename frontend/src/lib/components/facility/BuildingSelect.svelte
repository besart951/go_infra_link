<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { getBuilding, listBuildings } from '$lib/infrastructure/api/facility.adapter.js';
	import type { Building } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<Building[]> {
		const res = await listBuildings({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<Building> {
		return getBuilding(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="iws_code"
	placeholder="Select Building..."
	{width}
/>
