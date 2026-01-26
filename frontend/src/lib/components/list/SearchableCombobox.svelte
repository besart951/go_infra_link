<script lang="ts" generics="T extends { id: string }">
	import { Check, ChevronsUpDown } from 'lucide-svelte';
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import type { ListState } from '$lib/application/useCases/listUseCase.js';

	interface Props {
		state: ListState<T>;
		value?: string;
		placeholder?: string;
		searchPlaceholder?: string;
		emptyText?: string;
		getLabel: (item: T) => string;
		onSelect: (id: string) => void;
		onSearch: (text: string) => void;
		onLoadMore?: () => void;
		disabled?: boolean;
	}

	let {
		state,
		value = $bindable(),
		placeholder = 'Select...',
		searchPlaceholder = 'Search...',
		emptyText = 'No items found',
		getLabel,
		onSelect,
		onSearch,
		onLoadMore,
		disabled = false
	}: Props = $props();

	let open = $state(false);
	let searchText = $state('');

	function handleSelect(itemId: string) {
		value = itemId === value ? '' : itemId;
		onSelect(value);
		open = false;
	}

	function handleSearch(newText: string) {
		searchText = newText;
		onSearch(newText);
	}

	const selectedItem = $derived(state.items.find((item) => item.id === value));
	const displayLabel = $derived(selectedItem ? getLabel(selectedItem) : placeholder);
</script>

<Popover.Root bind:open>
<Popover.Trigger asChild let:builder>
<Button
builders={[builder]}
variant="outline"
role="combobox"
aria-expanded={open}
class="w-full justify-between"
{disabled}
>
<span class="truncate">{displayLabel}</span>
<ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
</Button>
</Popover.Trigger>
<Popover.Content class="w-full p-0">
<Command.Root>
<Command.Input placeholder={searchPlaceholder} value={searchText} onValueChange={handleSearch} />
<Command.Empty>{emptyText}</Command.Empty>
<Command.Group>
{#if state.loading}
<Command.Item disabled>Loading...</Command.Item>
{:else}
{#each state.items as item (item.id)}
<Command.Item
value={item.id}
onSelect={() => handleSelect(item.id)}
>
<Check
class={cn(
'mr-2 h-4 w-4',
value === item.id ? 'opacity-100' : 'opacity-0'
)}
/>
{getLabel(item)}
</Command.Item>
{/each}
{/if}
</Command.Group>
{#if onLoadMore && state.page < state.totalPages}
<Command.Item onSelect={onLoadMore}>
Load more...
</Command.Item>
{/if}
</Command.Root>
</Popover.Content>
</Popover.Root>
