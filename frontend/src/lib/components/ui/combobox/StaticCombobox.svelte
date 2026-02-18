<script lang="ts" generics="T">
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import { cn } from '$lib/utils.js';
	import { Check, ChevronsUpDown } from '@lucide/svelte';

	interface StaticComboboxProps<T> {
		items: T[];
		value?: string;
		labelKey: keyof T;
		idKey?: keyof T;
		id?: string;
		disabled?: boolean;
		clearable?: boolean;
		clearText?: string;
		placeholder?: string;
		searchPlaceholder?: string;
		emptyText?: string;
		width?: string;
		onValueChange?: (value: string) => void;
		error?: string;
	}

	let {
		items,
		value = $bindable(''),
		labelKey,
		idKey = 'id' as keyof T,
		id,
		disabled = false,
		clearable = false,
		clearText = 'Clear selection',
		placeholder = 'Select item...',
		searchPlaceholder = 'Search...',
		emptyText = 'No results found.',
		width = 'w-[200px]',
		onValueChange,
		error
	}: StaticComboboxProps<T> = $props();

	let open = $state(false);
	let search = $state('');

	const selectedItem = $derived(items.find((i) => String(i[idKey]) === value));
	const selectedLabel = $derived(selectedItem ? String(selectedItem[labelKey] ?? '') : undefined);
	const hasError = $derived(!!error);

	const filteredItems = $derived(
		search
			? items.filter((i) =>
					String(i[labelKey] ?? '')
						.toLowerCase()
						.includes(search.toLowerCase())
				)
			: items
	);

	function clearSelection() {
		value = '';
		onValueChange?.('');
		open = false;
	}
</script>

<Popover.Root bind:open>
	<Popover.Trigger>
		{#snippet child({ props })}
				{#if hasError}
					<Tooltip.Provider>
						<Tooltip.Root>
							<Tooltip.Trigger>
								{#snippet child({ props: tooltipProps })}
									<Button
										{...props}
										{...tooltipProps}
										{id}
										variant="outline"
										role="combobox"
										aria-expanded={open}
										{disabled}
										aria-disabled={disabled}
										class={cn(
											'justify-between border-destructive text-destructive',
											width
										)}
									>
										{selectedLabel || (value ? value : placeholder)}
										<ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
									</Button>
								{/snippet}
							</Tooltip.Trigger>
							<Tooltip.Content side="top">
								<p>{error}</p>
							</Tooltip.Content>
						</Tooltip.Root>
					</Tooltip.Provider>
				{:else}
					<Button
						{...props}
						{id}
						variant="outline"
						role="combobox"
						aria-expanded={open}
						{disabled}
						aria-disabled={disabled}
						class={cn('justify-between', width)}
					>
						{selectedLabel || (value ? value : placeholder)}
						<ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
					</Button>
				{/if}
		{/snippet}
	</Popover.Trigger>
	<Popover.Content class={cn('p-0', width)}>
		<Command.Root shouldFilter={false}>
			<Command.Input placeholder={searchPlaceholder} bind:value={search} />
			<Command.List>
				<Command.Empty>{emptyText}</Command.Empty>
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
					{#each filteredItems as item (String(item[idKey]))}
						<Command.Item
							value={String(item[idKey])}
							onSelect={() => {
								const next = String(item[idKey] ?? '');
								if (!next || next === 'undefined' || next === 'null') return;
								value = next;
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
							{String(item[labelKey] ?? '')}
						</Command.Item>
					{/each}
				</Command.Group>
			</Command.List>
		</Command.Root>
	</Popover.Content>
</Popover.Root>
