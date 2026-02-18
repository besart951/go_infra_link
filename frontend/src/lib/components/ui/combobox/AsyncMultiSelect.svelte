<script lang="ts" generics="T">
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { cn } from '$lib/utils.js';
	import { Check, ChevronsUpDown, X } from '@lucide/svelte';

	interface AsyncMultiSelectProps<T> {
		value?: string[];
		fetcher: (search: string) => Promise<T[]>;
		fetchByIds?: (ids: string[]) => Promise<T[]>;
		labelKey: keyof T;
		idKey?: keyof T;
		id?: string;
		placeholder?: string;
		searchPlaceholder?: string;
		emptyText?: string;
		width?: string;
		disabled?: boolean;
	}

	let {
		value = $bindable([]),
		fetcher,
		fetchByIds,
		labelKey,
		idKey = 'id' as keyof T,
		id,
		placeholder = 'Select items...',
		searchPlaceholder = 'Search...',
		emptyText = 'No results found.',
		width = 'w-full',
		disabled = false
	}: AsyncMultiSelectProps<T> = $props();

	let open = $state(false);
	let items = $state<T[]>([]);
	let search = $state('');
	let loading = $state(false);
	let debounceTimer: ReturnType<typeof setTimeout>;
	let initialized = $state(false);
	let selectedItems = $state<T[]>([]);
	let selectedLoading = $state(false);

	// Derived state
	const availableItems = $derived(items.filter((item) => !value.includes(String(item[idKey]))));

	// Load selected items by IDs
	async function loadSelected() {
		if (!fetchByIds || value.length === 0) {
			selectedItems = [];
			return;
		}
		selectedLoading = true;
		try {
			selectedItems = await fetchByIds(value);
		} catch (error) {
			console.error('Failed to fetch selected items:', error);
			selectedItems = [];
		} finally {
			selectedLoading = false;
		}
	}

	// Load items from fetcher
	function loadItems(query: string) {
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(async () => {
			loading = true;
			try {
				const res = await fetcher(query);
				items = res;
			} catch (error) {
				console.error('Failed to fetch items:', error);
				items = [];
			} finally {
				loading = false;
			}
		}, 500);
	}

	// Effects
	$effect(() => {
		if (open && !initialized) {
			initialized = true;
			loadItems('');
		}
	});

	$effect(() => {
		if (initialized) {
			loadItems(search);
		}
	});

	$effect(() => {
		if (value && value.length > 0) {
			loadSelected();
		} else {
			selectedItems = [];
		}
	});

	function handleSelect(itemId: string) {
		if (value.includes(itemId)) {
			// Remove from selection
			value = value.filter((id) => id !== itemId);
		} else {
			// Add to selection
			value = [...value, itemId];
		}
	}

	function handleRemove(itemId: string) {
		value = value.filter((id) => id !== itemId);
	}
</script>

<div class={cn('space-y-2', width)}>
	<!-- Selected items display -->
	{#if selectedItems.length > 0}
		<div class="flex flex-wrap gap-2">
			{#each selectedItems as item (String(item[idKey]))}
				<Badge variant="secondary" class="pr-1 pl-2">
					{String(item[labelKey] ?? '')}
					<button
						type="button"
						class="ml-1 rounded-full p-0.5 hover:bg-secondary-foreground/20"
						onclick={() => handleRemove(String(item[idKey]))}
						{disabled}
					>
						<X class="h-3 w-3" />
					</button>
				</Badge>
			{/each}
		</div>
	{/if}

	<!-- Dropdown selector -->
	<Popover.Root bind:open>
		<Popover.Trigger>
			{#snippet child({ props })}
				<Button
					{...props}
					{id}
					variant="outline"
					role="combobox"
					aria-expanded={open}
					class={cn('justify-between', width)}
					{disabled}
				>
					{selectedItems.length > 0 ? `${selectedItems.length} selected` : placeholder}
					<ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
				</Button>
			{/snippet}
		</Popover.Trigger>
		<Popover.Content class={cn('p-0', width)}>
			<Command.Root shouldFilter={false}>
				<Command.Input placeholder={searchPlaceholder} bind:value={search} />
				<Command.List>
					<Command.Empty>{loading ? 'Loading...' : emptyText}</Command.Empty>
					<Command.Group>
						{#each availableItems as item (String(item[idKey]))}
							<Command.Item
								value={String(item[idKey])}
								onSelect={() => handleSelect(String(item[idKey]))}
							>
								<Check
									class={cn(
										'mr-2 h-4 w-4',
										value.includes(String(item[idKey])) ? 'opacity-100' : 'opacity-0'
									)}
								/>
								{String(item[labelKey] ?? '')}
							</Command.Item>
						{/each}
					</Command.Group>
				</Command.List>
			</Command.Root>
		</Popover.Content>
	</Popover.Root>
</div>
