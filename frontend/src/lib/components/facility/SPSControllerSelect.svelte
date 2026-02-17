<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
	import type { SPSController } from '$lib/domain/facility/index.js';

	interface Props {
		value?: string;
		width?: string;
	}

	let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

	async function fetcher(search: string): Promise<SPSController[]> {
		const res = await spsControllerRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<SPSController> {
		return spsControllerRepository.get(id);
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
