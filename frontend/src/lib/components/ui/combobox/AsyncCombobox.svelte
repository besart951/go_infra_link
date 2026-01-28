<script lang="ts" generics="T">
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import Check from 'lucide-svelte/icons/check';
	import ChevronsUpDown from 'lucide-svelte/icons/chevrons-up-down';
	import { tick } from 'svelte';

	// Props
	export let value: string = '';
	export let fetcher: (search: string) => Promise<T[]>;
	export let fetchById: ((id: string) => Promise<T | null | undefined>) | undefined = undefined;
	export let labelKey: keyof T;
	export let idKey: keyof T = 'id' as keyof T;
	export let id: string | undefined = undefined;
	export let placeholder: string = 'Select item...';
	export let searchPlaceholder: string = 'Search...';
	export let emptyText: string = 'No results found.';
	export let width: string = 'w-[200px]';

	let open = false;
	let items: T[] = [];
	let search = '';
	let loading = false;
	let debounceTimer: ReturnType<typeof setTimeout>;
	let initialized = false;
	let selectedLoading = false;
	let selectedRequestId = 0;
	let currentValue = value;

	// We keep track of the label for the selected value to display it even if it's not in the current search results
	let selectedLabel: string | undefined = undefined;

	async function loadSelected(id: string) {
		if (!fetchById) return;
		selectedLoading = true;
		const requestId = ++selectedRequestId;
		try {
			const item = await fetchById(id);
			if (requestId !== selectedRequestId) return;
			if (item) {
				selectedLabel = String(item[labelKey] ?? '');
			}
		} catch (error) {
			console.error('Failed to fetch selected item:', error);
		} finally {
			if (requestId === selectedRequestId) {
				selectedLoading = false;
			}
		}
	}

	$: if (open && !initialized) {
		initialized = true;
		loadItems('');
	}

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

	// Trigger search when search term changes, skip initial empty string if not open/initialized
	$: if (initialized) {
		loadItems(search);
	}

	$: selectedItem = items.find((i) => String(i[idKey]) === value);
	$: if (selectedItem) {
		selectedLabel = String(selectedItem[labelKey] ?? '');
	}

	$: if (value !== currentValue) {
		currentValue = value;
		selectedLabel = undefined;
	}

	$: if (value && !selectedLabel && !selectedItem && fetchById && !selectedLoading) {
		loadSelected(value);
	}

	$: if (!value) {
		selectedLabel = undefined;
	}
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
				class={cn('justify-between', width)}
			>
				{selectedLabel || (value ? value : placeholder)}
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
					{#each items as item (String(item[idKey]))}
						<Command.Item
							value={String(item[idKey])}
							onSelect={() => {
								value = String(item[idKey]);
								selectedLabel = String(item[labelKey] ?? '');
								open = false;
							}}
						>
							<Check
								class={cn(
									'mr-2 h-4 w-4',
									value === String(item[idKey]) ? 'opacity-100' : 'opacity-0'
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
