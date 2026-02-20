<script lang="ts" generics="T">
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import {Check, ChevronsUpDown} from '@lucide/svelte';

	interface AsyncComboboxProps<T> {
		value?: string;
		fetcher: (search: string) => Promise<T[]>;
		fetchById?: (id: string) => Promise<T | null | undefined>;
		labelKey: keyof T;
		idKey?: keyof T;
		labelFormatter?: (item: T) => string;
		id?: string;
		disabled?: boolean;
		clearable?: boolean;
		clearText?: string;
		placeholder?: string;
		searchPlaceholder?: string;
		emptyText?: string;
		width?: string;
		onValueChange?: (value: string) => void;
	}

	let {
		value = $bindable(),
		fetcher,
		fetchById,
		labelKey,
		idKey = 'id' as keyof T,
		labelFormatter,
		id,
		disabled = false,
		clearable = false,
		clearText = 'Clear selection',
		placeholder = 'Select item...',
		searchPlaceholder = 'Search...',
		emptyText = 'No results found.',
		width = 'w-[200px]',
		onValueChange
	}: AsyncComboboxProps<T> = $props();

	let open = $state(false);
	let items = $state<T[]>([]);
	let search = $state('');
	let loading = $state(false);
	let debounceTimer: ReturnType<typeof setTimeout>;
	let initialized = $state(false);
	let selectedLoading = $state(false);
	let selectedRequestId = $state(0);
	let selectedValue = $state<string | undefined>(undefined);
	let selectedLabel = $state<string | undefined>(undefined);

	// Derived state
	const selectedItem = $derived(items.find((i) => String(i[idKey]) === value));

	function getItemLabel(item: T): string {
		return labelFormatter ? labelFormatter(item) : String(item[labelKey] ?? '');
	}

	function clearSelection() {
		value = '';
		selectedLabel = undefined;
		selectedValue = undefined;
		onValueChange?.('');
		open = false;
	}

	// Load selected item by ID
	async function loadSelected(id: string) {
		if (!fetchById) return;
		selectedLoading = true;
		const requestId = ++selectedRequestId;
		try {
			const item = await fetchById(id);
			if (requestId !== selectedRequestId) return;
			if (item) {
				selectedLabel = getItemLabel(item);
				selectedValue = id;
			}
		} catch (error) {
			console.error('Failed to fetch selected item:', error);
		} finally {
			if (requestId === selectedRequestId) {
				selectedLoading = false;
			}
		}
	}

	// Load items with debounce
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
		if (selectedItem) {
			selectedLabel = getItemLabel(selectedItem);
			selectedValue = value;
		}
	});

	$effect(() => {
		if (value && selectedValue && value !== selectedValue) {
			selectedLabel = undefined;
		}
	});

	$effect(() => {
		if (value && !selectedLabel && !selectedItem && fetchById && !selectedLoading) {
			loadSelected(value);
		}
	});

	$effect(() => {
		if (!value) {
			selectedLabel = undefined;
			selectedValue = undefined;
		}
	});
</script>

<Popover.Root bind:open>
	<Popover.Trigger>
		{#snippet child({ props })}
			<Button
				{...props}
				{id}
				variant="outline"
				role="combobox"
				aria-expanded={open}
				{disabled}
				aria-disabled={disabled}
				class={cn('justify-between gap-2 min-w-0', width)}
			>
				<span class="min-w-0 flex-1 truncate text-left">
					{selectedLabel || (value ? value : placeholder)}
				</span>
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
					{#if clearable && value}
						<Command.Item
							value=""
							onSelect={() => {
								clearSelection();
							}}
						>
							{clearText}
						</Command.Item>
					{/if}
					{#each items as item (String(item[idKey]))}
						<Command.Item
							value={String(item[idKey])}
							onSelect={() => {
								const next = String(item[idKey] ?? '');
								if (!next || next === 'undefined' || next === 'null') {
									console.warn('AsyncCombobox: selected item has no valid id', item);
									return;
								}
								value = next;
								selectedLabel = getItemLabel(item);
								selectedValue = value;
								onValueChange?.(value);
								open = false;
							}}
						>
							<Check
								class={cn(
									'mr-2 h-4 w-4',
									value === String(item[idKey]) ? 'opacity-100' : 'opacity-0'
								)}
							/>
							{getItemLabel(item)}
						</Command.Item>
					{/each}
				</Command.Group>
			</Command.List>
		</Command.Root>
	</Popover.Content>
</Popover.Root>
