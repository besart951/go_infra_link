<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
	import type { Apparat } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';
	export let onValueChange: ((value: string) => void) | undefined = undefined;

	async function fetcher(search: string): Promise<Apparat[]> {
		const res = await apparatRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<Apparat> {
		return apparatRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder="Select Apparat..."
	{width}
	{onValueChange}
/>
