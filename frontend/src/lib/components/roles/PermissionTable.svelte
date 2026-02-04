<script lang="ts">
	import type { Permission } from '$lib/domain/role/index.js';
	import {
		parsePermissionName,
		FACILITY_RESOURCES,
		type FacilityResource
	} from '$lib/domain/role/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { MoreVertical, Pencil, Trash2, Search, Plus } from '@lucide/svelte';

	interface Props {
		permissions: Permission[];
		onEdit?: (permission: Permission) => void;
		onDelete?: (permission: Permission) => void;
		onCreate?: () => void;
		canManage?: boolean;
	}

	let { permissions, onEdit, onDelete, onCreate, canManage = false }: Props = $props();

	let searchQuery = $state('');

	const filteredPermissions = $derived(() => {
		if (!searchQuery.trim()) return permissions;

		const query = searchQuery.toLowerCase();
		return permissions.filter(
			(p) =>
				p.name.toLowerCase().includes(query) ||
				p.description.toLowerCase().includes(query) ||
				p.resource.toLowerCase().includes(query) ||
				p.action.toLowerCase().includes(query)
		);
	});

	// Group permissions by category for display
	const groupedByCategory = $derived(() => {
		const grouped: Record<string, Permission[]> = {
			general: [],
			facility: [],
			project: []
		};
		for (const perm of filteredPermissions()) {
			const parsed = parsePermissionName(perm.name);
			grouped[parsed.category].push(perm);
		}
		return grouped;
	});

	const categories = $derived(
		(['general', 'facility', 'project'] as const).filter(
			(cat) => groupedByCategory()[cat].length > 0
		)
	);

	function getCategoryLabel(cat: string): string {
		switch (cat) {
			case 'general':
				return 'General';
			case 'facility':
				return 'Facility';
			case 'project':
				return 'Project Resources';
			default:
				return cat;
		}
	}

	function getCategoryBadgeClass(cat: string): string {
		switch (cat) {
			case 'general':
				return 'bg-green-500/10 text-green-600';
			case 'facility':
				return 'bg-amber-500/10 text-amber-600';
			case 'project':
				return 'bg-blue-500/10 text-blue-600';
			default:
				return 'bg-muted text-muted-foreground';
		}
	}

	function getActionVariant(action: string): 'default' | 'secondary' | 'outline' | 'destructive' {
		switch (action) {
			case 'create':
				return 'default';
			case 'delete':
				return 'destructive';
			case 'update':
				return 'secondary';
			default:
				return 'outline';
		}
	}

	function formatResource(perm: Permission): string {
		const parsed = parsePermissionName(perm.name);
		if (parsed.subResource) {
			return `${parsed.resource}.${parsed.subResource}`;
		}
		return parsed.resource;
	}
</script>

<div class="space-y-4">
	<!-- Header -->
	<div class="flex items-center justify-between gap-4">
		<div class="relative max-w-sm flex-1">
			<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search permissions..."
				class="pl-9"
				bind:value={searchQuery}
			/>
		</div>
		{#if canManage && onCreate}
			<Button onclick={onCreate}>
				<Plus class="mr-2 h-4 w-4" />
				Create Permission
			</Button>
		{/if}
	</div>

	<!-- Stats -->
	<div class="flex flex-wrap gap-3 text-sm text-muted-foreground">
		<span>{filteredPermissions().length} permissions</span>
		{#if searchQuery}
			<span>matching "{searchQuery}"</span>
		{/if}
		<span class="text-muted-foreground/50">|</span>
		{#each categories as cat}
			<span class={getCategoryBadgeClass(cat) + ' rounded px-2 py-0.5 text-xs'}>
				{getCategoryLabel(cat)}: {groupedByCategory()[cat].length}
			</span>
		{/each}
	</div>

	<!-- Table -->
	<div class="rounded-lg border bg-background">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head class="w-50">Permission</Table.Head>
					<Table.Head class="w-30">Category</Table.Head>
					<Table.Head class="w-36">Resource</Table.Head>
					<Table.Head class="w-25">Action</Table.Head>
					<Table.Head>Description</Table.Head>
					{#if canManage}
						<Table.Head class="w-20">Actions</Table.Head>
					{/if}
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if filteredPermissions().length === 0}
					<Table.Row>
						<Table.Cell colspan={canManage ? 6 : 5} class="h-24 text-center">
							<div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
								<p class="font-medium">
									{#if searchQuery}
										No permissions found matching "{searchQuery}"
									{:else}
										No permissions available
									{/if}
								</p>
							</div>
						</Table.Cell>
					</Table.Row>
				{:else}
					{#each filteredPermissions() as permission}
						{@const parsed = parsePermissionName(permission.name)}
						<Table.Row>
							<Table.Cell class="font-mono text-sm">
								{permission.name}
							</Table.Cell>
							<Table.Cell>
								<span
									class={getCategoryBadgeClass(parsed.category) + ' rounded px-2 py-0.5 text-xs'}
								>
									{getCategoryLabel(parsed.category)}
								</span>
							</Table.Cell>
							<Table.Cell>
								<Badge variant="outline" class="capitalize">
									{formatResource(permission)}
								</Badge>
							</Table.Cell>
							<Table.Cell>
								<Badge variant={getActionVariant(permission.action)}>
									{permission.action}
								</Badge>
							</Table.Cell>
							<Table.Cell class="text-muted-foreground">
								{permission.description}
							</Table.Cell>
							{#if canManage}
								<Table.Cell>
									<DropdownMenu.Root>
										<DropdownMenu.Trigger>
											<Button variant="ghost" size="icon" class="h-8 w-8">
												<MoreVertical class="h-4 w-4" />
												<span class="sr-only">Open menu</span>
											</Button>
										</DropdownMenu.Trigger>
										<DropdownMenu.Content align="end">
											{#if onEdit}
												<DropdownMenu.Item onclick={() => onEdit(permission)}>
													<Pencil class="mr-2 h-4 w-4" />
													Edit
												</DropdownMenu.Item>
											{/if}
											{#if onDelete}
												<DropdownMenu.Separator />
												<DropdownMenu.Item
													class="text-destructive focus:text-destructive"
													onclick={() => onDelete(permission)}
												>
													<Trash2 class="mr-2 h-4 w-4" />
													Delete
												</DropdownMenu.Item>
											{/if}
										</DropdownMenu.Content>
									</DropdownMenu.Root>
								</Table.Cell>
							{/if}
						</Table.Row>
					{/each}
				{/if}
			</Table.Body>
		</Table.Root>
	</div>
</div>
