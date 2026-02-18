<script lang="ts">
	import type { Role, Permission } from '$lib/domain/role/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Search } from '@lucide/svelte';
	import Users from '@lucide/svelte/icons/users';
	import Building2 from '@lucide/svelte/icons/building-2';
	import FolderKanban from '@lucide/svelte/icons/folder-kanban';
	import CategorySection from './CategorySection.svelte';
	import { createTranslator } from '$lib/i18n/translator.js';
	import { t as translate } from '$lib/i18n/index.js';

	// ============================================================================
	// Props
	// ============================================================================

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

	const t = createTranslator();

	// ============================================================================
	// Category Configuration
	// ============================================================================

	type CategoryId = 'users' | 'facility' | 'projects';

	const categories: { id: CategoryId; label: string; icon: typeof Users; resources: string[] }[] = [
		{
			id: 'users',
			label: 'roles.categories.users_access',
			icon: Users,
			resources: ['user', 'team', 'role', 'permission']
		},
		{
			id: 'facility',
			label: 'roles.categories.facility',
			icon: Building2,
			resources: [
				'building',
				'controlcabinet',
				'spscontroller',
				'spscontrollersystemtype',
				'fielddevice',
				'bacnetobject',
				'systemtype',
				'systempart',
				'apparat',
				'objectdata',
				'specification',
				'statetext',
				'alarmdefinition',
				'notificationclass'
			]
		},
		{
			id: 'projects',
			label: 'roles.categories.projects',
			icon: FolderKanban,
			resources: [] // Will include all project.* permissions
		}
	];

	// ============================================================================
	// State
	// ============================================================================

	const initialPermissions = $derived(new Set(role.permissions));
	let selectedPermissions = $state<Set<string>>(new Set());
	let searchQuery = $state('');
	let expandedCategories = $state<Set<CategoryId>>(new Set(['users', 'facility', 'projects']));
	let expandedResources = $state<Set<string>>(new Set());

	// Initialize when role changes
	$effect(() => {
		selectedPermissions = new Set(initialPermissions);
	});

	// ============================================================================
	// Derived State
	// ============================================================================

	// Categorize a permission
	function categorizePermission(perm: Permission): CategoryId {
		if (perm.name.startsWith('project.') || perm.resource.startsWith('project.')) {
			return 'projects';
		}
		for (const cat of categories) {
			if (cat.resources.includes(perm.resource)) {
				return cat.id;
			}
		}
		return 'facility';
	}

	// Group permissions by category, then by resource
	const permissionsByCategory = $derived(() => {
		const result: Record<CategoryId, Record<string, Permission[]>> = {
			users: {},
			facility: {},
			projects: {}
		};

		for (const perm of allPermissions) {
			const category = categorizePermission(perm);
			const resource = perm.resource;

			if (!result[category][resource]) {
				result[category][resource] = [];
			}
			result[category][resource].push(perm);
		}

		// Sort permissions within each resource
		for (const cat of Object.keys(result) as CategoryId[]) {
			for (const resource of Object.keys(result[cat])) {
				result[cat][resource].sort((a, b) => a.action.localeCompare(b.action));
			}
		}

		return result;
	});

	// Filter permissions based on search
	const filteredPermissionsByCategory = $derived(() => {
		const grouped = permissionsByCategory();
		if (!searchQuery.trim()) return grouped;

		const query = searchQuery.toLowerCase();
		const filtered: Record<CategoryId, Record<string, Permission[]>> = {
			users: {},
			facility: {},
			projects: {}
		};

		for (const cat of Object.keys(grouped) as CategoryId[]) {
			for (const [resource, perms] of Object.entries(grouped[cat])) {
				const matchingPerms = perms.filter(
					(p) =>
						p.name.toLowerCase().includes(query) ||
						p.description.toLowerCase().includes(query) ||
						p.resource.toLowerCase().includes(query) ||
						p.action.toLowerCase().includes(query)
				);
				if (matchingPerms.length > 0) {
					filtered[cat][resource] = matchingPerms;
				}
			}
		}
		return filtered;
	});

	const hasAnyPermissions = $derived(
		Object.values(filteredPermissionsByCategory()).some((cat) => Object.keys(cat).length > 0)
	);

	// ============================================================================
	// Actions
	// ============================================================================

	function togglePermission(permissionName: string) {
		const newSet = new Set(selectedPermissions);
		if (newSet.has(permissionName)) {
			newSet.delete(permissionName);
		} else {
			newSet.add(permissionName);
		}
		selectedPermissions = newSet;
	}

	function toggleCategory(categoryId: CategoryId) {
		const newExpanded = new Set(expandedCategories);
		if (newExpanded.has(categoryId)) {
			newExpanded.delete(categoryId);
		} else {
			newExpanded.add(categoryId);
		}
		expandedCategories = newExpanded;
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

	function toggleAllInResource(resource: string, categoryId: CategoryId) {
		const resourcePerms = permissionsByCategory()[categoryId][resource] || [];
		const allSelected = resourcePerms.every((p) => selectedPermissions.has(p.name));

		const newSet = new Set(selectedPermissions);
		if (allSelected) {
			resourcePerms.forEach((p) => newSet.delete(p.name));
		} else {
			resourcePerms.forEach((p) => newSet.add(p.name));
		}
		selectedPermissions = newSet;
	}

	function toggleAllInCategory(categoryId: CategoryId) {
		const categoryPerms = Object.values(permissionsByCategory()[categoryId]).flat();
		const allSelected = categoryPerms.every((p) => selectedPermissions.has(p.name));

		const newSet = new Set(selectedPermissions);
		if (allSelected) {
			categoryPerms.forEach((p) => newSet.delete(p.name));
		} else {
			categoryPerms.forEach((p) => newSet.add(p.name));
		}
		selectedPermissions = newSet;
	}

	function selectAll() {
		selectedPermissions = new Set(allPermissions.map((p) => p.name));
	}

	function deselectAll() {
		selectedPermissions = new Set();
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		onSubmit({ permissions: Array.from(selectedPermissions) });
	}

	const RESOURCE_DISPLAY_NAMES: Record<string, string> = {
		// Users & Access
		user: 'roles.resources.user',
		team: 'roles.resources.team',
		role: 'roles.resources.role',
		permission: 'roles.resources.permission',
		// Facility
		building: 'roles.resources.building',
		controlcabinet: 'roles.resources.controlcabinet',
		spscontroller: 'roles.resources.spscontroller',
		spscontrollersystemtype: 'roles.resources.spscontrollersystemtype',
		fielddevice: 'roles.resources.fielddevice',
		bacnetobject: 'roles.resources.bacnetobject',
		systemtype: 'roles.resources.systemtype',
		systempart: 'roles.resources.systempart',
		apparat: 'roles.resources.apparat',
		objectdata: 'roles.resources.objectdata',
		specification: 'roles.resources.specification',
		statetext: 'roles.resources.statetext',
		alarmdefinition: 'roles.resources.alarmdefinition',
		notificationclass: 'roles.resources.notificationclass',
		// Projects
		'project.controlcabinet': 'roles.resources.project.controlcabinet',
		'project.spscontroller': 'roles.resources.project.spscontroller',
		'project.spscontrollersystemtype': 'roles.resources.project.spscontrollersystemtype',
		'project.fielddevice': 'roles.resources.project.fielddevice',
		'project.bacnetobject': 'roles.resources.project.bacnetobject',
		'project.systemtype': 'roles.resources.project.systemtype'
	};

	function getResourceDisplayName(resource: string): string {
		if (RESOURCE_DISPLAY_NAMES[resource]) {
			return translate(RESOURCE_DISPLAY_NAMES[resource]);
		}
		// Handle project.* resources
		if (resource.startsWith('project.')) {
			const subResource = resource.replace('project.', '');
			return RESOURCE_DISPLAY_NAMES[subResource]
				? translate(RESOURCE_DISPLAY_NAMES[subResource])
				: subResource;
		}
		return resource;
	}

	// Expand all when searching
	$effect(() => {
		if (searchQuery.trim()) {
			const allResources = new Set<string>();
			for (const cat of Object.keys(filteredPermissionsByCategory()) as CategoryId[]) {
				for (const resource of Object.keys(filteredPermissionsByCategory()[cat])) {
					allResources.add(resource);
				}
			}
			expandedResources = allResources;
			expandedCategories = new Set(['users', 'facility', 'projects']);
		}
	});
</script>

<form onsubmit={handleSubmit} class="flex h-full flex-col gap-4 p-6">
	<!-- Header -->
	<div class="shrink-0">
		<h3 class="text-lg font-semibold">{$t('roles.editor.title')}</h3>
		<p class="text-sm text-muted-foreground">
			{$t('roles.editor.description', { role: role.display_name })}
		</p>
	</div>

	<!-- Error Message -->
	{#if error}
		<div
			class="shrink-0 rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive"
		>
			{error}
		</div>
	{/if}

	<!-- Full Access Warning -->
	<!-- Search and Bulk Actions -->
	<div class="flex shrink-0 items-center gap-4">
		<div class="relative flex-1">
			<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder={$t('roles.permissions.search_placeholder')}
				class="pl-9"
				bind:value={searchQuery}
			/>
		</div>
		<div class="flex gap-2">
			<Button type="button" variant="outline" size="sm" onclick={selectAll}>
				{$t('roles.actions.select_all')}
			</Button>
			<Button type="button" variant="outline" size="sm" onclick={deselectAll}>
				{$t('roles.actions.deselect_all')}
			</Button>
		</div>
	</div>

	<!-- Selected Count -->
	<div class="shrink-0 text-sm text-muted-foreground">
		{$t('roles.editor.selected_count', {
			selected: selectedPermissions.size,
			total: allPermissions.length
		})}
	</div>

	<!-- Permission Categories -->
	<div class="min-h-0 flex-1 space-y-3 overflow-y-auto rounded-lg border bg-muted/30 p-3">
		{#each categories as category (category.id)}
			{@const categoryResources = Object.keys(filteredPermissionsByCategory()[category.id]).sort()}
			<CategorySection
				id={category.id}
				label={$t(category.label)}
				icon={category.icon}
				resources={categoryResources}
				permissionsByResource={filteredPermissionsByCategory()[category.id]}
				{selectedPermissions}
				isExpanded={expandedCategories.has(category.id)}
				{expandedResources}
				disabled={false}
				onToggleExpand={() => toggleCategory(category.id)}
				onToggleAll={() => toggleAllInCategory(category.id)}
				onToggleResource={toggleResource}
				onToggleResourceAll={(resource) => toggleAllInResource(resource, category.id)}
				onTogglePermission={togglePermission}
				{getResourceDisplayName}
			/>
		{/each}

		{#if !hasAnyPermissions}
			<div class="py-8 text-center text-muted-foreground">
				{#if searchQuery}
					{$t('roles.permissions.empty_match', { query: searchQuery })}
				{:else}
					{$t('roles.permissions.empty')}
				{/if}
			</div>
		{/if}
	</div>

	<!-- Actions -->
	<div class="flex shrink-0 justify-end gap-3 border-t pt-4">
		<Button type="button" variant="outline" onclick={onCancel} disabled={isSubmitting}>
			{$t('common.cancel')}
		</Button>
		<Button type="submit" disabled={isSubmitting}>
			{#if isSubmitting}
				<span class="mr-2 h-4 w-4 animate-spin">⟳</span>
			{/if}
			{$t('roles.actions.save_changes')}
		</Button>
	</div>
</form>
