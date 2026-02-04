<script lang="ts">
	import type { Role, Permission } from '$lib/domain/role/index.js';
	import type { UserRole } from '$lib/domain/user/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { ChevronDown, ChevronRight, Search } from '@lucide/svelte';
	import { cn } from '$lib/utils.js';

	interface Props {
		role: Role;
		allPermissions: Permission[];
		onSubmit: (data: { permissions: string[] }) => void;
		onCancel: () => void;
		isSubmitting?: boolean;
		error?: string | null;
	}

	let {
		role,
		allPermissions,
		onSubmit,
		onCancel,
		isSubmitting = false,
		error = null
	}: Props = $props();

	// Track selected permissions - use $derived for initial value from role
	const initialPermissions = $derived(new Set(role.permissions.filter((p) => p !== '*')));

	let selectedPermissions = $state<Set<string>>(new Set());
	let searchQuery = $state('');
	let expandedResources = $state<Set<string>>(new Set());

	// Initialize selected permissions when role changes
	$effect(() => {
		selectedPermissions = new Set(initialPermissions);
	});

	// Check if role has all permissions (superadmin)
	const hasAllPermissions = $derived(role.permissions.includes('*'));

	// Group permissions by resource
	const permissionsByResource = $derived(() => {
		const grouped: Record<string, Permission[]> = {};
		for (const perm of allPermissions) {
			if (!grouped[perm.resource]) {
				grouped[perm.resource] = [];
			}
			grouped[perm.resource].push(perm);
		}
		return grouped;
	});

	// Filter permissions based on search
	const filteredPermissionsByResource = $derived(() => {
		const grouped = permissionsByResource();
		if (!searchQuery.trim()) return grouped;

		const query = searchQuery.toLowerCase();
		const filtered: Record<string, Permission[]> = {};

		for (const [resource, perms] of Object.entries(grouped)) {
			const matchingPerms = perms.filter(
				(p) =>
					p.name.toLowerCase().includes(query) ||
					p.description.toLowerCase().includes(query) ||
					p.resource.toLowerCase().includes(query) ||
					p.action.toLowerCase().includes(query)
			);
			if (matchingPerms.length > 0) {
				filtered[resource] = matchingPerms;
			}
		}
		return filtered;
	});

	const resources = $derived(Object.keys(filteredPermissionsByResource()).sort());

	function togglePermission(permissionName: string) {
		if (hasAllPermissions) return; // Can't modify superadmin permissions

		const newSet = new Set(selectedPermissions);
		if (newSet.has(permissionName)) {
			newSet.delete(permissionName);
		} else {
			newSet.add(permissionName);
		}
		selectedPermissions = newSet;
	}

	function toggleResource(resource: string) {
		const newExpanded = new Set(expandedResources);
		if (newExpanded.has(resource)) {
			newExpanded.delete(resource);
		} else {
			newExpanded.add(resource);
		}
		expandedResources = newExpanded;
	}

	function toggleAllInResource(resource: string) {
		if (hasAllPermissions) return;

		const resourcePerms = permissionsByResource()[resource] || [];
		const allSelected = resourcePerms.every((p) => selectedPermissions.has(p.name));

		const newSet = new Set(selectedPermissions);
		if (allSelected) {
			// Deselect all
			resourcePerms.forEach((p) => newSet.delete(p.name));
		} else {
			// Select all
			resourcePerms.forEach((p) => newSet.add(p.name));
		}
		selectedPermissions = newSet;
	}

	function isResourceFullySelected(resource: string): boolean {
		const resourcePerms = permissionsByResource()[resource] || [];
		return resourcePerms.length > 0 && resourcePerms.every((p) => selectedPermissions.has(p.name));
	}

	function isResourcePartiallySelected(resource: string): boolean {
		const resourcePerms = permissionsByResource()[resource] || [];
		const selectedCount = resourcePerms.filter((p) => selectedPermissions.has(p.name)).length;
		return selectedCount > 0 && selectedCount < resourcePerms.length;
	}

	function getResourceSelectedCount(resource: string): number {
		const resourcePerms = permissionsByResource()[resource] || [];
		return resourcePerms.filter((p) => selectedPermissions.has(p.name)).length;
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		onSubmit({
			permissions: Array.from(selectedPermissions)
		});
	}

	function selectAll() {
		if (hasAllPermissions) return;
		selectedPermissions = new Set(allPermissions.map((p) => p.name));
	}

	function deselectAll() {
		if (hasAllPermissions) return;
		selectedPermissions = new Set();
	}

	// Expand all resources when searching
	$effect(() => {
		if (searchQuery.trim()) {
			expandedResources = new Set(Object.keys(filteredPermissionsByResource()));
		}
	});
</script>

<form onsubmit={handleSubmit} class="flex flex-col gap-4">
	<!-- Header -->
	<div>
		<h3 class="text-lg font-semibold">Edit Role Permissions</h3>
		<p class="text-sm text-muted-foreground">
			Configure permissions for <span class="font-medium">{role.display_name}</span>
		</p>
	</div>

	{#if error}
		<div
			class="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive"
		>
			{error}
		</div>
	{/if}

	{#if hasAllPermissions}
		<div
			class="rounded-md border border-blue-200 bg-blue-50 p-3 text-sm text-blue-800 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-200"
		>
			<p class="font-medium">Full Access Role</p>
			<p>This role has all permissions and cannot be modified.</p>
		</div>
	{/if}

	<!-- Search and bulk actions -->
	<div class="flex items-center gap-4">
		<div class="relative flex-1">
			<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search permissions..."
				class="pl-9"
				bind:value={searchQuery}
			/>
		</div>
		{#if !hasAllPermissions}
			<div class="flex gap-2">
				<Button type="button" variant="outline" size="sm" onclick={selectAll}>Select All</Button>
				<Button type="button" variant="outline" size="sm" onclick={deselectAll}>
					Deselect All
				</Button>
			</div>
		{/if}
	</div>

	<!-- Selected count -->
	<div class="text-sm text-muted-foreground">
		{#if hasAllPermissions}
			All permissions granted
		{:else}
			{selectedPermissions.size} of {allPermissions.length} permissions selected
		{/if}
	</div>

	<!-- Permission Groups -->
	<div class="max-h-100 space-y-2 overflow-y-auto rounded-md border p-2">
		{#each resources as resource}
			{@const resourcePerms = filteredPermissionsByResource()[resource] || []}
			{@const isExpanded = expandedResources.has(resource)}
			{@const isFullySelected = isResourceFullySelected(resource)}
			{@const isPartiallySelected = isResourcePartiallySelected(resource)}
			{@const selectedCount = getResourceSelectedCount(resource)}
			{@const totalCount = (permissionsByResource()[resource] || []).length}

			<div class="rounded-md border bg-card">
				<!-- Resource Header -->
				<div class="flex items-center gap-2 p-3">
					{#if !hasAllPermissions}
						<Checkbox
							checked={isFullySelected}
							indeterminate={isPartiallySelected}
							onCheckedChange={() => toggleAllInResource(resource)}
							aria-label={`Select all ${resource} permissions`}
						/>
					{/if}

					<button
						type="button"
						class="-ml-2 flex flex-1 items-center gap-2 rounded px-2 py-1 text-left hover:bg-accent/50"
						onclick={() => toggleResource(resource)}
					>
						{#if isExpanded}
							<ChevronDown class="h-4 w-4 text-muted-foreground" />
						{:else}
							<ChevronRight class="h-4 w-4 text-muted-foreground" />
						{/if}
						<span class="font-medium capitalize">{resource}</span>
						<Badge variant="secondary" class="ml-auto">
							{selectedCount}/{totalCount}
						</Badge>
					</button>
				</div>

				<!-- Permission List -->
				{#if isExpanded}
					<div class="border-t px-3 pb-3">
						<div class="mt-2 space-y-1">
							{#each resourcePerms as perm}
								{@const isSelected = hasAllPermissions || selectedPermissions.has(perm.name)}
								<label
									class={cn(
										'flex items-start gap-3 rounded-md p-2 transition-colors',
										hasAllPermissions
											? 'cursor-not-allowed opacity-60'
											: 'cursor-pointer hover:bg-accent/50'
									)}
								>
									<Checkbox
										checked={isSelected}
										disabled={hasAllPermissions}
										onCheckedChange={() => togglePermission(perm.name)}
										class="mt-0.5"
									/>
									<div class="flex-1 space-y-1">
										<div class="flex items-center gap-2">
											<span class="font-mono text-sm">{perm.name}</span>
										</div>
										<p class="text-xs text-muted-foreground">{perm.description}</p>
									</div>
								</label>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/each}

		{#if resources.length === 0}
			<div class="py-8 text-center text-muted-foreground">
				{#if searchQuery}
					No permissions found matching "{searchQuery}"
				{:else}
					No permissions available
				{/if}
			</div>
		{/if}
	</div>

	<!-- Actions -->
	<div class="flex justify-end gap-3 pt-2">
		<Button type="button" variant="outline" onclick={onCancel} disabled={isSubmitting}>
			Cancel
		</Button>
		<Button type="submit" disabled={isSubmitting || hasAllPermissions}>
			{#if isSubmitting}
				<span class="mr-2 h-4 w-4 animate-spin">⟳</span>
			{/if}
			Save Changes
		</Button>
	</div>
</form>
