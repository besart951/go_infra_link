<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { listSystemParts } from '$lib/infrastructure/api/facility.adapter.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<SystemPart[]> {
		const res = await listSystemParts({ search, limit: 20 });
		return res.items || [];
	}
</script>

<AsyncCombobox bind:value {fetcher} labelKey="name" placeholder="Select System Part..." {width} />
