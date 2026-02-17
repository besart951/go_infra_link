<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<ControlCabinet[]> {
		const res = await controlCabinetRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<ControlCabinet> {
		return controlCabinetRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="control_cabinet_nr"
	placeholder="Select Control Cabinet..."
	{width}
/>
