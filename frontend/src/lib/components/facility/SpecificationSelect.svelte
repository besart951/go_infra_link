<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { listSpecifications } from '$lib/infrastructure/api/facility.adapter.js';
	import type { Specification } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<Specification[]> {
		const res = await listSpecifications({ search, limit: 20 });
		return res.items || [];
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	labelKey="specification_type"
	placeholder="Select Specification..."
	{width}
/>
