<script lang="ts">
	import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
	import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
	import type { Apparat } from '$lib/domain/facility/index.js';

	type Props = {
		value?: string[];
		width?: string;
		disabled?: boolean;
		id?: string;
	};

	let { value = $bindable([]), width = 'w-full', disabled = false, id }: Props = $props();

	type ApparatOption = Apparat & { label: string };

	function toOption(apparat: Apparat): ApparatOption {
		return {
			...apparat,
			label: `${apparat.short_name} — ${apparat.name}`
		};
	}

	async function fetcher(search: string): Promise<ApparatOption[]> {
		const res = await apparatRepository.list({
			pagination: { page: 1, pageSize: 50 },
			search: { text: search }
		});
		return res.items.map(toOption);
	}

	async function fetchByIds(ids: string[]): Promise<ApparatOption[]> {
		if (ids.length === 0) return [];
		const items = await apparatRepository.getBulk(ids);
		return items.map(toOption);
	}
</script>

<AsyncMultiSelect
	bind:value
	{fetcher}
	{fetchByIds}
	labelKey="label"
	placeholder="Select Apparats..."
	searchPlaceholder="Search apparats..."
	emptyText="No apparats found."
	{width}
	{disabled}
	{id}
/>
