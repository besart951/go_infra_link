<script lang="ts">
	import type { Permission } from '$lib/domain/role/index.js';
	import type { Component } from 'svelte';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { ChevronDown, ChevronRight } from '@lucide/svelte';
	import ResourceSection from './ResourceSection.svelte';

	interface Props {
		id: string;
		label: string;
		icon: Component<{ class?: string }>;
		resources: string[];
		permissionsByResource: Record<string, Permission[]>;
		selectedPermissions: Set<string>;
		isExpanded: boolean;
		expandedResources: Set<string>;
		disabled?: boolean;
		onToggleExpand: () => void;
		onToggleAll: () => void;
		onToggleResource: (resource: string) => void;
		onToggleResourceAll: (resource: string) => void;
		onTogglePermission: (permissionName: string) => void;
		getResourceDisplayName: (resource: string) => string;
	}

	let {
		id,
		label,
		icon: Icon,
		resources,
		permissionsByResource,
		selectedPermissions,
		isExpanded,
		expandedResources,
		disabled = false,
		onToggleExpand,
		onToggleAll,
		onToggleResource,
		onToggleResourceAll,
		onTogglePermission,
		getResourceDisplayName
	}: Props = $props();

	// Calculate category totals
	const categoryPermissions = $derived(Object.values(permissionsByResource).flat());
	const selectedCount = $derived(
		categoryPermissions.filter((p) => selectedPermissions.has(p.name)).length
	);
	const totalCount = $derived(categoryPermissions.length);
	const isFullySelected = $derived(totalCount > 0 && selectedCount === totalCount);
	const isPartiallySelected = $derived(selectedCount > 0 && selectedCount < totalCount);
</script>

{#if resources.length > 0}
	<div class="rounded-lg border bg-background">
		<!-- Category Header -->
		<div class="flex items-center gap-2 border-b bg-muted/50 px-4 py-3">
			{#if !disabled}
				<Checkbox
					checked={isFullySelected}
					indeterminate={isPartiallySelected}
					onCheckedChange={onToggleAll}
					aria-label={`Select all ${label} permissions`}
				/>
			{/if}

			<button
				type="button"
				class="flex flex-1 items-center gap-2 text-left"
				onclick={onToggleExpand}
			>
				{#if isExpanded}
					<ChevronDown class="h-4 w-4 text-muted-foreground" />
				{:else}
					<ChevronRight class="h-4 w-4 text-muted-foreground" />
				{/if}
				<Icon class="h-5 w-5 text-primary" />
				<span class="font-semibold">{label}</span>
				<Badge variant="outline" class="ml-auto">
					{selectedCount}/{totalCount}
				</Badge>
			</button>
		</div>

		<!-- Resources within Category -->
		{#if isExpanded}
			<div class="space-y-1 p-2">
				{#each resources as resource (resource)}
					{@const resourcePerms = permissionsByResource[resource] || []}
					<ResourceSection
						{resource}
						displayName={getResourceDisplayName(resource)}
						permissions={resourcePerms}
						{selectedPermissions}
						isExpanded={expandedResources.has(resource)}
						{disabled}
						onToggleExpand={() => onToggleResource(resource)}
						onToggleAll={() => onToggleResourceAll(resource)}
						{onTogglePermission}
					/>
				{/each}
			</div>
		{/if}
	</div>
{/if}
