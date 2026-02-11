<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { getSystemType, listSystemTypes } from '$lib/infrastructure/api/facility.adapter.js';
	import type { SystemType } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	function formatNumber(value: number): string {
		return String(value).padStart(4, '0');
	}

	function buildLabel(item: SystemType): string {
		return `${item.name} (${formatNumber(item.number_min)}-${formatNumber(item.number_max)})`;
	}

	async function fetcher(search: string): Promise<(SystemType & { display_label: string })[]> {
		const res = await listSystemTypes({ search, limit: 20 });
		return (res.items || []).map((item) => ({
			...item,
			display_label: buildLabel(item)
		}));
	}

	async function fetchById(id: string): Promise<SystemType & { display_label: string }> {
		const item = await getSystemType(id);
		return {
			...item,
			display_label: buildLabel(item)
		};
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="display_label"
	placeholder="Select System Type..."
	{width}
/>
