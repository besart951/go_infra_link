<script lang="ts">
	import type { Permission } from '$lib/domain/role/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { ChevronDown, ChevronRight } from '@lucide/svelte';
	import PermissionItem from './PermissionItem.svelte';

	interface Props {
		resource: string;
		displayName: string;
		permissions: Permission[];
		selectedPermissions: Set<string>;
		isExpanded: boolean;
		disabled?: boolean;
		onToggleExpand: () => void;
		onToggleAll: () => void;
		onTogglePermission: (permissionName: string) => void;
	}

	let {
		resource,
		displayName,
		permissions,
		selectedPermissions,
		isExpanded,
		disabled = false,
		onToggleExpand,
		onToggleAll,
		onTogglePermission
	}: Props = $props();

	const selectedCount = $derived(permissions.filter((p) => selectedPermissions.has(p.name)).length);
	const totalCount = $derived(permissions.length);
	const isFullySelected = $derived(totalCount > 0 && selectedCount === totalCount);
	const isPartiallySelected = $derived(selectedCount > 0 && selectedCount < totalCount);
</script>

<div class="rounded-md border bg-card">
	<!-- Resource Header -->
	<div class="flex items-center gap-2 px-3 py-2">
		{#if !disabled}
			<Checkbox
				checked={isFullySelected}
				indeterminate={isPartiallySelected}
				onCheckedChange={onToggleAll}
				aria-label={`Select all ${resource} permissions`}
			/>
		{/if}

		<button
			type="button"
			class="flex flex-1 items-center gap-2 rounded px-1 py-0.5 text-left hover:bg-accent/50"
			onclick={onToggleExpand}
		>
			{#if isExpanded}
				<ChevronDown class="h-3.5 w-3.5 text-muted-foreground" />
			{:else}
				<ChevronRight class="h-3.5 w-3.5 text-muted-foreground" />
			{/if}
			<span class="text-sm font-medium capitalize">{displayName}</span>
			<Badge variant="secondary" class="ml-auto text-xs">
				{selectedCount}/{totalCount}
			</Badge>
		</button>
	</div>

	<!-- Permission List -->
	{#if isExpanded}
		<div class="border-t px-3 pb-2">
			<div class="mt-1 space-y-0.5">
				{#each permissions as perm (perm.id)}
					<PermissionItem
						permission={perm}
						isSelected={disabled || selectedPermissions.has(perm.name)}
						{disabled}
						onToggle={() => onTogglePermission(perm.name)}
					/>
				{/each}
			</div>
		</div>
	{/if}
</div>
