<script lang="ts" generics="T">
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { cn } from '$lib/utils.js';
	import Check from 'lucide-svelte/icons/check';
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import X from 'lucide-svelte/icons/x';

	// Props
	export let value: string[] = [];
	export let fetcher: (search: string) => Promise<T[]>;
	export let fetchByIds: ((ids: string[]) => Promise<T[]>) | undefined = undefined;
	export let labelKey: keyof T;
	export let idKey: keyof T = 'id' as keyof T;
	export let id: string | undefined = undefined;
	export let placeholder: string = 'Select items...';
	export let searchPlaceholder: string = 'Search...';
	export let emptyText: string = 'No results found.';
	export let width: string = 'w-full';
	export let disabled: boolean = false;

	let open = false;
	let items: T[] = [];
	let search = '';
	let loading = false;
	let debounceTimer: ReturnType<typeof setTimeout>;
	let initialized = false;
	let selectedItems: T[] = [];
	let selectedLoading = false;

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

	// Initialize on first open
	$: if (open && !initialized) {
		initialized = true;
		loadItems('');
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

	// Trigger search when search term changes
	$: if (initialized) {
		loadItems(search);
	}

	// Load selected items when value changes
	$: if (value && value.length > 0) {
		loadSelected();
	} else {
		selectedItems = [];
	}

	// Filter out selected items from available items
	$: availableItems = items.filter((item) => !value.includes(String(item[idKey])));

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
