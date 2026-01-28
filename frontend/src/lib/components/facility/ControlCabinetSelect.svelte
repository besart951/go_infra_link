<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import {
		getControlCabinet,
		listControlCabinets
	} from '$lib/infrastructure/api/facility.adapter.js';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<ControlCabinet[]> {
		const res = await listControlCabinets({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<ControlCabinet> {
		return getControlCabinet(id);
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
