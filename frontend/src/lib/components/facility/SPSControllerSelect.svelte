<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import {
		getSPSController,
		listSPSControllers
	} from '$lib/infrastructure/api/facility.adapter.js';
	import type { SPSController } from '$lib/domain/facility/index.js';

	interface Props {
		value?: string;
		width?: string;
	}

	let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

	async function fetcher(search: string): Promise<SPSController[]> {
		const res = await listSPSControllers({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<SPSController> {
		return getSPSController(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="device_name"
	placeholder="Select SPS Controller..."
	{width}
/>
