<script lang="ts" generics="T">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { Search, ChevronLeft, ChevronRight } from '@lucide/svelte';
	import type { Snippet } from 'svelte';
	import type { ListState } from '$lib/application/useCases/listUseCase.js';

	interface Props {
		state: ListState<T>;
		columns: Array<{ key: string; label: string; width?: string }>;
		rowSnippet: Snippet<[T]>;
		emptyMessage?: string;
		searchPlaceholder?: string;
		onSearch: (text: string) => void;
		onPageChange: (page: number) => void;
		onReload?: () => void;
	}

	let {
		state,
		columns,
		rowSnippet,
		emptyMessage = 'No items found',
		searchPlaceholder = 'Search...',
		onSearch,
		onPageChange,
		onReload
	}: Props = $props();

	let searchInput = $derived(state.searchText);

	function handleSearchInput(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		onSearch(value);
	}

	function handlePrevious() {
		if (state.page > 1) {
			onPageChange(state.page - 1);
		}
	}

	function handleNext() {
		if (state.page < state.totalPages) {
			onPageChange(state.page + 1);
		}
	}
</script>

<div class="flex flex-col gap-4">
	<!-- Search Bar -->
	<div class="flex items-center gap-4">
		<div class="relative flex-1 ">
			<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder={searchPlaceholder}
				class="pl-9"
				value={searchInput}
				oninput={handleSearchInput}
			/>
		</div>
		{#if onReload}
			<Button variant="outline" onclick={onReload} disabled={state.loading}>Refresh</Button>
		{/if}
	</div>

	<!-- Error Message -->
	{#if state.error}
		<div class="rounded-md border border-destructive/50 bg-destructive/15 px-4 py-3 text-destructive">
			<p class="font-medium">Error</p>
			<p class="text-sm">{state.error}</p>
		</div>
	{/if}

	<!-- Table -->
	<div class="rounded-lg border bg-background">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					{#each columns as column}
						<Table.Head class={column.width}>{column.label}</Table.Head>
					{/each}
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if state.loading && state.items.length === 0}
					{#each Array(5) as _}
						<Table.Row>
							{#each columns as _}
								<Table.Cell>
									<Skeleton class="h-8 w-full" />
								</Table.Cell>
							{/each}
						</Table.Row>
					{/each}
				{:else if state.items.length === 0}
					<Table.Row>
						<Table.Cell colspan={columns.length} class="h-24 text-center">
							<div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
								<p class="font-medium">{emptyMessage}</p>
								{#if state.searchText}
									<p class="text-sm">Try adjusting your search</p>
								{/if}
							</div>
						</Table.Cell>
					</Table.Row>
				{:else}
					{#each state.items as item}
						<Table.Row class={state.loading ? "opacity-60" : undefined}>
							{@render rowSnippet(item)}
						</Table.Row>
					{/each}
				{/if}
			</Table.Body>
		</Table.Root>
	</div>

	<!-- Pagination -->
	{#if state.totalPages > 1}
		<div class="flex items-center justify-between">
			<div class="text-sm text-muted-foreground">
				Page {state.page} of {state.totalPages} • {state.total}
				{state.total === 1 ? 'item' : 'items'} total
			</div>
			<div class="flex items-center gap-2">
				<Button
					variant="outline"
					size="sm"
					disabled={state.page <= 1 || state.loading}
					onclick={handlePrevious}
				>
					<ChevronLeft class="h-4 w-4 mr-1" />
					Previous
				</Button>
				<Button
					variant="outline"
					size="sm"
					disabled={state.page >= state.totalPages || state.loading}
					onclick={handleNext}
				>
					Next
					<ChevronRight class="h-4 w-4 ml-1" />
				</Button>
			</div>
		</div>
	{/if}
</div>
