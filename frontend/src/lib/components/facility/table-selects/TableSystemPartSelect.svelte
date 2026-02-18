<script lang="ts">
	import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
	import type { SystemPart } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

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

	const t = createTranslator();

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
	placeholder={$t('field_device.table_select.system_part')}
	{width}
	{onValueChange}
	{disabled}
	{error}
/>
