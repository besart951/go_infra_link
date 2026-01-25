<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { listObjectData } from '$lib/infrastructure/api/facility.adapter.js';
	import type { ObjectData } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<ObjectData[]> {
		const res = await listObjectData({ search, limit: 20 });
		return res.items || [];
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	labelKey="description"
	placeholder="Select Object Data..."
	{width}
/>
