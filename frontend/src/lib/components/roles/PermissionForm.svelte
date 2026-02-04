<script lang="ts">
	import type { Permission, PermissionCategory } from '$lib/domain/role/index.js';
	import {
		PERMISSION_ACTIONS,
		GENERAL_RESOURCES,
		FACILITY_RESOURCES,
		PROJECT_SUB_RESOURCES,
		createPermissionName,
		parsePermissionName
	} from '$lib/domain/role/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { cn } from '$lib/utils.js';
	import Building2 from '@lucide/svelte/icons/building-2';
	import FolderKanban from '@lucide/svelte/icons/folder-kanban';
	import Settings from '@lucide/svelte/icons/settings';

	interface Props {
		permission?: Permission | null;
		onSubmit: (data: {
			name: string;
			description: string;
			resource: string;
			action: string;
		}) => void;
		onCancel: () => void;
		isSubmitting?: boolean;
		error?: string | null;
	}

	let {
		permission = null,
		onSubmit,
		onCancel,
		isSubmitting = false,
		error = null
	}: Props = $props();

	// Parse existing permission to determine initial state
	const parsed = $derived(permission ? parsePermissionName(permission.name) : null);

	// Category selection
	let category = $state<PermissionCategory>('general');

	// Resource/action selection
	let resource = $state('');
	let subResource = $state('');
	let action = $state('');
	let description = $state('');

	// Custom inputs
	let customResource = $state('');
	let customSubResource = $state('');
	let customAction = $state('');

	// Initialize state when permission changes
	$effect(() => {
		if (parsed) {
			category = parsed.category;
			resource = parsed.resource;
			subResource = parsed.subResource || '';
			action = parsed.action;
			description = permission?.description || '';
		} else {
			category = 'general';
			resource = '';
			subResource = '';
			action = '';
			description = '';
		}
	});

	const isEditMode = $derived(!!permission);
	const title = $derived(isEditMode ? 'Edit Permission' : 'Create Permission');
	const submitText = $derived(isEditMode ? 'Update' : 'Create');

	// Get resources based on category
	const availableResources = $derived.by(() => {
		switch (category) {
			case 'facility':
				return [...FACILITY_RESOURCES];
			case 'project':
				return ['project']; // Only project for this category
			case 'general':
			default:
				return [...GENERAL_RESOURCES];
		}
	});

	// Build permission name based on category
	const permissionName = $derived.by(() => {
		if (category === 'project' && subResource && action) {
			return createPermissionName('project', action, subResource);
		}
		if (resource && action) {
			return createPermissionName(resource, action);
		}
		return '';
	});

	// Check if form is valid
	const isValid = $derived.by(() => {
		if (category === 'project') {
			return subResource && action;
		}
		return resource && action;
	});

	function handleSubmit(e: Event) {
		e.preventDefault();
		if (!isValid) return;

		const finalResource = category === 'project' ? `project.${subResource}` : resource;

		onSubmit({
			name: permissionName,
			description,
			resource: finalResource,
			action
		});
	}

	function selectCategory(cat: PermissionCategory) {
		category = cat;
		// Reset selections when changing category
		resource = cat === 'project' ? 'project' : '';
		subResource = '';
		action = '';
		customResource = '';
		customSubResource = '';
	}

	function selectResource(res: string) {
		resource = res;
		customResource = '';
	}

	function selectSubResource(sub: string) {
		subResource = sub;
		customSubResource = '';
	}

	function selectAction(act: string) {
		action = act;
		customAction = '';
	}

	function handleCustomResource(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		customResource = value;
		resource = value;
	}

	function handleCustomSubResource(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		customSubResource = value;
		subResource = value;
	}

	function handleCustomAction(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		customAction = value;
		action = value;
	}

	const categoryConfig = [
		{
			id: 'general' as const,
			label: 'General',
			icon: Settings,
			description: 'User, Team, Project, Phase...',
			example: 'user.create'
		},
		{
			id: 'facility' as const,
			label: 'Facility',
			icon: Building2,
			description: 'Building, Systempart, Apparat...',
			example: 'building.read'
		},
		{
			id: 'project' as const,
			label: 'Project Resources',
			icon: FolderKanban,
			description: 'Control Cabinet, SPS Controller...',
			example: 'project.controlcabinet.create'
		}
	];
</script>

<form onsubmit={handleSubmit} class="space-y-5">
	<div>
		<h3 class="text-lg font-semibold">{title}</h3>
		<p class="text-sm text-muted-foreground">
			{isEditMode ? 'Update permission details' : 'Create a new permission for role assignment'}
		</p>
	</div>

	{#if error}
		<div
			class="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive"
		>
			{error}
		</div>
	{/if}

	<!-- Category Selection -->
	{#if !isEditMode}
		<div class="space-y-2">
			<Label>Permission Category</Label>
			<div class="grid grid-cols-3 gap-2">
				{#each categoryConfig as cat}
					<button
						type="button"
						class={cn(
							'flex flex-col items-center gap-1.5 rounded-lg border p-3 text-center transition-all',
							category === cat.id
								? 'border-primary bg-primary/5 ring-1 ring-primary'
								: 'border-border hover:border-primary/50 hover:bg-muted/50'
						)}
						onclick={() => selectCategory(cat.id)}
					>
						<cat.icon
							class={cn('h-5 w-5', category === cat.id ? 'text-primary' : 'text-muted-foreground')}
						/>
						<span class="text-sm font-medium">{cat.label}</span>
						<span class="text-xs text-muted-foreground">{cat.description}</span>
						<code class="mt-1 rounded bg-muted px-1.5 py-0.5 text-xs">{cat.example}</code>
					</button>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Resource Selection (for General and Facility) -->
	{#if category !== 'project'}
		<div class="space-y-2">
			<Label for="resource">Resource</Label>
			<Input
				type="text"
				id="resource"
				placeholder="Enter or select resource..."
				value={resource}
				oninput={handleCustomResource}
				disabled={isEditMode}
			/>
			{#if !isEditMode}
				<div class="flex flex-wrap gap-1.5">
					{#each availableResources as res}
						<button
							type="button"
							class={cn(
								'rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
								resource === res
									? 'bg-primary text-primary-foreground'
									: 'bg-muted hover:bg-muted/80'
							)}
							onclick={() => selectResource(res)}
						>
							{res}
						</button>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	<!-- Sub-Resource Selection (for Project category) -->
	{#if category === 'project'}
		<div class="space-y-2">
			<Label for="subResource">Project Resource</Label>
			<p class="text-xs text-muted-foreground">
				Select which project resource this permission applies to
			</p>
			<Input
				type="text"
				id="subResource"
				placeholder="Enter or select project resource..."
				value={subResource}
				oninput={handleCustomSubResource}
				disabled={isEditMode}
			/>
			{#if !isEditMode}
				<div class="flex flex-wrap gap-1.5">
					{#each PROJECT_SUB_RESOURCES as sub}
						<button
							type="button"
							class={cn(
								'rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
								subResource === sub
									? 'bg-primary text-primary-foreground'
									: 'bg-muted hover:bg-muted/80'
							)}
							onclick={() => selectSubResource(sub)}
						>
							{sub}
						</button>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	<!-- Action Selection -->
	<div class="space-y-2">
		<Label for="action">Action</Label>
		<Input
			type="text"
			id="action"
			placeholder="Enter or select action..."
			value={action}
			oninput={handleCustomAction}
			disabled={isEditMode}
		/>
		{#if !isEditMode}
			<div class="flex flex-wrap gap-1.5">
				{#each PERMISSION_ACTIONS as act}
					<button
						type="button"
						class={cn(
							'rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
							action === act ? 'bg-primary text-primary-foreground' : 'bg-muted hover:bg-muted/80'
						)}
						onclick={() => selectAction(act)}
					>
						{act}
					</button>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Permission Name Preview -->
	{#if permissionName}
		<div class="space-y-2">
			<Label>Permission Name</Label>
			<div class="flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2">
				<code class="flex-1 font-mono text-sm">{permissionName}</code>
				{#if category === 'project'}
					<span class="rounded bg-blue-500/10 px-2 py-0.5 text-xs text-blue-600"
						>Project Resource</span
					>
				{:else if category === 'facility'}
					<span class="rounded bg-amber-500/10 px-2 py-0.5 text-xs text-amber-600">Facility</span>
				{:else}
					<span class="rounded bg-green-500/10 px-2 py-0.5 text-xs text-green-600">General</span>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Description -->
	<div class="space-y-2">
		<Label for="description">Description</Label>
		<Textarea
			id="description"
			bind:value={description}
			placeholder="Describe what this permission allows..."
			rows={3}
		/>
	</div>

	<!-- Actions -->
	<div class="flex justify-end gap-3 pt-4">
		<Button type="button" variant="outline" onclick={onCancel} disabled={isSubmitting}>
			Cancel
		</Button>
		<Button type="submit" disabled={isSubmitting || !isValid}>
			{#if isSubmitting}
				<span class="mr-2 h-4 w-4 animate-spin">⟳</span>
			{/if}
			{submitText}
		</Button>
	</div>
</form>
