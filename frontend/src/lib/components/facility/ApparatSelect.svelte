<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { getApparat, listApparats } from '$lib/infrastructure/api/facility.adapter.js';
	import type { Apparat } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<Apparat[]> {
		const res = await listApparats({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<Apparat> {
		return getApparat(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder="Select Apparat..."
	{width}
/>
