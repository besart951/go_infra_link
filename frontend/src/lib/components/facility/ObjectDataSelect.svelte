<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
	import type { ObjectData } from '$lib/domain/facility/index.js';

	type Props = {
		value?: string;
		width?: string;
	};

	let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

	async function fetcher(search: string): Promise<ObjectData[]> {
		const res = await objectDataRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<ObjectData> {
		return objectDataRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="description"
	placeholder="Select Object Data..."
	{width}
/>
