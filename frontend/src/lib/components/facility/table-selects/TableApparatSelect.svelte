<script lang="ts">
	import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
	import type { Apparat } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface Props {
		items: Apparat[];
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

	const t = createTranslator();

	function formatLabel(item: Apparat): string {
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
	placeholder={$t('field_device.table_select.apparat')}
	{width}
	{onValueChange}
	{disabled}
	{error}
/>
