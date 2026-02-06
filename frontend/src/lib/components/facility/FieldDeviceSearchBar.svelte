<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Search, X, Trash2, Settings2, RefreshCw } from '@lucide/svelte';

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
			placeholder="Search field devices..."
			class="pl-9"
			value={searchInput}
			oninput={handleSearchInput}
		/>
	</div>

	{#if selectedCount > 0}
		<div class="flex items-center gap-2">
			<span class="text-sm text-muted-foreground">{selectedCount} selected</span>
			<Button variant="outline" size="sm" onclick={onClearSelection}>
				<X class="mr-1 h-4 w-4" />
				Clear
			</Button>
			<Button variant="destructive" size="sm" onclick={onBulkDelete}>
				<Trash2 class="mr-1 h-4 w-4" />
				Delete
			</Button>
			<Button
				variant={showBulkEditPanel ? 'secondary' : 'outline'}
				size="sm"
				onclick={onToggleBulkEdit}
			>
				<Settings2 class="mr-1 h-4 w-4" />
				Bulk Edit
			</Button>
		</div>
	{/if}

	<Button
		variant="outline"
		onclick={onRefresh}
		disabled={loading}
	>
		Refresh
	</Button>
</div>
