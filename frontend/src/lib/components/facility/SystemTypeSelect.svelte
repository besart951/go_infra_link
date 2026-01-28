<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { getSystemType, listSystemTypes } from '$lib/infrastructure/api/facility.adapter.js';
	import type { SystemType } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<SystemType[]> {
		const res = await listSystemTypes({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<SystemType> {
		return getSystemType(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder="Select System Type..."
	{width}
/>
