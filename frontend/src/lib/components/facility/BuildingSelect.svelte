<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
	import type { Building } from '$lib/domain/facility/index.js';

	type BuildingOption = Building & { display_name: string };

	type Props = {
		value?: string;
		width?: string;
	};

	let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

	function toOption(item: Building): BuildingOption {
		return {
			...item,
			display_name: `${item.iws_code}-${item.building_group}`
		};
	}

	async function fetcher(search: string): Promise<BuildingOption[]> {
		const res = await buildingRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items.map(toOption);
	}

	async function fetchById(id: string): Promise<BuildingOption> {
		const building = await buildingRepository.get(id);
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
