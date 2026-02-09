<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { getBuilding, listBuildings } from '$lib/infrastructure/api/facility.adapter.js';
	import type { Building } from '$lib/domain/facility/index.js';

	type BuildingOption = Building & { display_name: string };

	export let value: string = '';
	export let width: string = 'w-[250px]';

	function toOption(item: Building): BuildingOption {
		return {
			...item,
			display_name: `${item.iws_code}-${item.building_group}`
		};
	}

	async function fetcher(search: string): Promise<BuildingOption[]> {
		const res = await listBuildings({ search, limit: 20 });
		return (res.items || []).map(toOption);
	}

	async function fetchById(id: string): Promise<BuildingOption> {
		const building = await getBuilding(id);
		return toOption(building);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="display_name"
	placeholder="Select Building..."
	{width}
/>
