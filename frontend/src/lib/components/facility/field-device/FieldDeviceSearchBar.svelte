<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Search, X, Trash2, Settings2, RefreshCw } from '@lucide/svelte';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface Props {
		searchInput: string;
		selectedCount: number;
		loading: boolean;
		showBulkEditPanel: boolean;
		onSearch: (value: string) => void;
		onClearSelection: () => void;
		onBulkDelete: () => void;
		onToggleBulkEdit: () => void;
		onRefresh: () => void;
	}

	let {
		searchInput,
		selectedCount,
		loading,
		showBulkEditPanel,
		onSearch,
		onClearSelection,
		onBulkDelete,
		onToggleBulkEdit,
		onRefresh
	}: Props = $props();

	const t = createTranslator();

	function handleSearchInput(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		onSearch(value);
	}
</script>

<div class="flex items-center gap-4">
	<div class="relative flex-1">
		<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
		<Input
			type="search"
			placeholder={$t('field_device.search.placeholder')}
			class="pl-9"
			value={searchInput}
			oninput={handleSearchInput}
		/>
	</div>

	{#if selectedCount > 0}
		<div class="flex items-center gap-2">
			<span class="text-sm text-muted-foreground">
				{$t('field_device.search.selected', { count: selectedCount })}
			</span>
			<Button variant="outline" size="sm" onclick={onClearSelection}>
				<X class="mr-1 h-4 w-4" />
				{$t('field_device.search.clear')}
			</Button>
			<Button variant="destructive" size="sm" onclick={onBulkDelete}>
				<Trash2 class="mr-1 h-4 w-4" />
				{$t('field_device.search.delete')}
			</Button>
			<Button
				variant={showBulkEditPanel ? 'secondary' : 'outline'}
				size="sm"
				onclick={onToggleBulkEdit}
			>
				<Settings2 class="mr-1 h-4 w-4" />
				{$t('field_device.search.bulk_edit')}
			</Button>
		</div>
	{/if}

	<Button variant="outline" onclick={onRefresh} disabled={loading}>
		{$t('field_device.search.refresh')}
	</Button>
</div>
