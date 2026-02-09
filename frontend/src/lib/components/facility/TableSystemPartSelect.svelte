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
</script>

<StaticCombobox
	{items}
	bind:value
	labelKey="name"
	placeholder="Select System Part..."
	{width}
	{onValueChange}
	{disabled}
	{error}
/>
