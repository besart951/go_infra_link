<script lang="ts">
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Search, Trash2, Settings2, TableIcon, Filter, X, RefreshCcw } from '@lucide/svelte';
	import { createTranslator } from '$lib/i18n/translator.js';
	import * as ButtonGroup from '$lib/components/ui/button-group/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import { canPerform } from '$lib/utils/permissions.js';

	interface Props {
		searchInput: string;
		selectedCount: number;
		loading: boolean;
		showBulkEditPanel: boolean;
		showExportPanel: boolean;
		showFilterPanel: boolean;
		hasActiveFilters: boolean;
		onSearch: (value: string) => void;
		onClearSelection: () => void;
		onBulkDelete: () => void;
		onToggleBulkEdit: () => void;
		onToggleExport: () => void;
		onToggleFilterPanel: () => void;
		onRefresh: () => void;
	}

	let {
		searchInput,
		selectedCount,
		loading,
		showBulkEditPanel,
		showExportPanel,
		showFilterPanel,
		hasActiveFilters,
		onSearch,
		onClearSelection,
		onBulkDelete,
		onToggleBulkEdit,
		onToggleExport,
		onToggleFilterPanel,
		onRefresh
	}: Props = $props();

	const t = createTranslator();

	function handleSearchInput(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		onSearch(value);
	}
</script>

<div class="flex items-center gap-3">
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

	<div class="ml-auto flex items-center gap-2">
		{#if selectedCount > 0}
			<span class="text-sm text-muted-foreground">
				{$t('field_device.search.selected', { count: selectedCount })}
			</span>
		{/if}

		<Tooltip.Provider>
			<ButtonGroup.Root>
				{#if selectedCount > 0}
					<Tooltip.Root>
						<Tooltip.Trigger
							class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
							onclick={onClearSelection}
						>
							<X />
						</Tooltip.Trigger>
						<Tooltip.Content>{$t('field_device.search.clear')}</Tooltip.Content>
					</Tooltip.Root>

					{#if canPerform('delete', 'fielddevice')}
					<Tooltip.Root>
						<Tooltip.Trigger
							class={buttonVariants({ variant: 'destructive', size: 'icon-sm' })}
							onclick={onBulkDelete}
						>
							<Trash2 />
						</Tooltip.Trigger>
						<Tooltip.Content>{$t('field_device.search.delete')}</Tooltip.Content>
					</Tooltip.Root>
					{/if}

					{#if canPerform('update', 'fielddevice')}
					<Tooltip.Root>
						<Tooltip.Trigger
							class={buttonVariants({
								variant: showBulkEditPanel ? 'secondary' : 'outline',
								size: 'icon-sm'
							})}
							onclick={onToggleBulkEdit}
						>
							<Settings2 />
						</Tooltip.Trigger>
						<Tooltip.Content>{$t('field_device.search.bulk_edit')}</Tooltip.Content>
					</Tooltip.Root>
					{/if}
				{/if}

				<Tooltip.Root>
					<Tooltip.Trigger
						class={buttonVariants({
							variant: showExportPanel ? 'secondary' : 'outline',
							size: 'icon-sm'
						})}
						onclick={onToggleExport}
					>
						<TableIcon />
					</Tooltip.Trigger>
					<Tooltip.Content>{$t('field_device.search.table')}</Tooltip.Content>
				</Tooltip.Root>

				<Tooltip.Root>
					<Tooltip.Trigger
						class={`${buttonVariants({
							variant: showFilterPanel ? 'secondary' : 'outline',
							size: 'icon-sm'
						})} relative`}
						onclick={onToggleFilterPanel}
					>
						<Filter />
						{#if hasActiveFilters}
							<span class="pointer-events-none absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full bg-green-500 ring-2 ring-background"></span>
						{/if}
					</Tooltip.Trigger>
					<Tooltip.Content>{$t('common.filter')}</Tooltip.Content>
				</Tooltip.Root>

				<Tooltip.Root>
					<Tooltip.Trigger
						class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
						onclick={onRefresh}
						disabled={loading}
					>
						<RefreshCcw />
					</Tooltip.Trigger>
					<Tooltip.Content>{$t('field_device.search.refresh')}</Tooltip.Content>
				</Tooltip.Root>
			</ButtonGroup.Root>
		</Tooltip.Provider>
	</div>
</div>
