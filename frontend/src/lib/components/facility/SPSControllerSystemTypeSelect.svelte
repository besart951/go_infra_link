<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import {
		getSPSControllerSystemType,
		listSPSControllerSystemTypes
	} from '$lib/infrastructure/api/facility.adapter.js';
	import type { SPSControllerSystemType } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<SPSControllerSystemType[]> {
		const res = await listSPSControllerSystemTypes({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<SPSControllerSystemType> {
		return getSPSControllerSystemType(id);
	}

	function formatLabel(item: SPSControllerSystemType): string {
		const number = item.number ?? '';
		const documentName = item.document_name ?? '';
		return `${number} - ${documentName}`;
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="document_name"
	labelFormatter={formatLabel}
	placeholder="Select SPS Controller System Type..."
	{width}
/>
