<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';
	export let onValueChange: ((value: string) => void) | undefined = undefined;

	async function fetcher(search: string): Promise<SystemPart[]> {
		const res = await systemPartRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<SystemPart> {
		return systemPartRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder="Select System Part..."
	{width}
	{onValueChange}
/>
