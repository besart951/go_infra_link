<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { listSPSControllerSystemTypes } from '$lib/infrastructure/api/facility.adapter.js';
	import type { SPSControllerSystemType } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<SPSControllerSystemType[]> {
		const res = await listSPSControllerSystemTypes({ search, limit: 20 });
		return res.items || [];
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	labelKey="document_name"
	placeholder="Select SPS Controller System Type..."
	{width}
/>
