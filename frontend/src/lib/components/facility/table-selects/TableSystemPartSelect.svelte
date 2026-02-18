<script lang="ts">
	import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
	import type { SystemPart } from '$lib/domain/facility/index.js';

	interface Props {
		items: SystemPart[];
		value?: string;
		width?: string;
		onValueChange?: (value: string) => void;
		disabled?: boolean;
		error?: string;
	}

	let {
		items,
		value = $bindable(''),
		width = 'w-full',
		onValueChange,
		disabled = false,
		error
	}: Props = $props();

	function formatLabel(item: SystemPart): string {
		return `${item.short_name ?? ''} - ${item.name ?? ''}`.trim();
	}

	const formattedItems = $derived(
		items.map((item) => ({
			...item,
			display_name: formatLabel(item)
		}))
	);
</script>

<StaticCombobox
	items={formattedItems}
	bind:value
	labelKey="display_name"
	placeholder="Select System Part..."
	{width}
	{onValueChange}
	{disabled}
	{error}
/>
