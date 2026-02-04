<script lang="ts">
	import type { Role, Permission, PermissionCategory } from '$lib/domain/role/index.js';
	import { parsePermissionName } from '$lib/domain/role/index.js';
	import type { UserRole } from '$lib/domain/user/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Check, Minus } from '@lucide/svelte';
	import RoleBadge from '$lib/components/role-badge.svelte';

	interface Props {
		roles: Role[];
		permissions: Permission[];
	}

	let { roles, permissions }: Props = $props();

	// Sort roles by level (highest first)
	const sortedRoles = $derived([...roles].sort((a, b) => b.level - a.level));

	// Group permissions by category and resource
	const permissionsByCategory = $derived(() => {
		const grouped: Record<PermissionCategory, Record<string, Permission[]>> = {
			general: {},
			facility: {},
			project: {}
		};
		for (const perm of permissions) {
			const parsed = parsePermissionName(perm.name);
			const category = parsed.category;
			const resource = parsed.subResource ? `project.${parsed.subResource}` : parsed.resource;

			if (!grouped[category][resource]) {
				grouped[category][resource] = [];
			}
			grouped[category][resource].push(perm);
		}
		// Sort permissions within each resource by action
		for (const category of Object.keys(grouped) as PermissionCategory[]) {
			for (const resource of Object.keys(grouped[category])) {
				grouped[category][resource].sort((a, b) => a.action.localeCompare(b.action));
			}
		}
		return grouped;
	});

	const categories = $derived(
		(['general', 'facility', 'project'] as const).filter(
			(cat) => Object.keys(permissionsByCategory()[cat]).length > 0
		)
	);

	function getCategoryLabel(cat: PermissionCategory): string {
		switch (cat) {
			case 'general':
				return '⚙️ General';
			case 'facility':
				return '🏢 Facility';
			case 'project':
				return '📁 Project Resources';
			default:
				return cat;
		}
	}

	function hasPermission(role: Role, permissionName: string): boolean {
		return role.permissions.includes('*') || role.permissions.includes(permissionName);
	}

	function getResourcePermissionCount(
		role: Role,
		resource: string,
		category: PermissionCategory
	): {
		count: number;
		total: number;
	} {
		const resourcePerms = permissionsByCategory()[category][resource] || [];
		const total = resourcePerms.length;
		if (role.permissions.includes('*')) {
			return { count: total, total };
		}
		const count = resourcePerms.filter((p) => role.permissions.includes(p.name)).length;
		return { count, total };
	}
</script>

<div class="space-y-4">
	<div class="text-sm text-muted-foreground">
		Showing {permissions.length} permissions across {roles.length} roles
	</div>

	<div class="overflow-x-auto rounded-lg border bg-background">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head class="sticky left-0 z-10 min-w-38 bg-background">Permission</Table.Head>
					{#each sortedRoles as role}
						<Table.Head class="min-w-30 text-center">
							<div class="flex flex-col items-center gap-1">
								<RoleBadge role={role.name} showIcon={false} />
								<span class="text-xs text-muted-foreground">L{role.level}</span>
							</div>
						</Table.Head>
					{/each}
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each categories as category}
					<!-- Category Header Row -->
					<Table.Row class="bg-muted/50">
						<Table.Cell
							colspan={sortedRoles.length + 1}
							class="sticky left-0 z-10 bg-muted/50 text-sm font-semibold"
						>
							{getCategoryLabel(category)}
						</Table.Cell>
					</Table.Row>

					{#each Object.keys(permissionsByCategory()[category]).sort() as resource}
						<!-- Resource Header Row -->
						<Table.Row class="bg-muted/30">
							<Table.Cell class="sticky left-0 z-10 bg-muted/30 pl-4 font-medium capitalize">
								{resource}
							</Table.Cell>
							{#each sortedRoles as role}
								{@const { count, total } = getResourcePermissionCount(role, resource, category)}
								<Table.Cell class="text-center">
									<Badge
										variant={count === total ? 'default' : count === 0 ? 'outline' : 'secondary'}
										class="text-xs"
									>
										{count}/{total}
									</Badge>
								</Table.Cell>
							{/each}
						</Table.Row>

						<!-- Permission Rows -->
						{#each permissionsByCategory()[category][resource] as perm}
							<Table.Row>
								<Table.Cell class="sticky left-0 z-10 bg-background pl-8">
									<div class="flex flex-col">
										<span class="font-mono text-sm">{perm.name}</span>
										<span class="text-xs text-muted-foreground">{perm.description}</span>
									</div>
								</Table.Cell>
								{#each sortedRoles as role}
									<Table.Cell class="text-center">
										{#if hasPermission(role, perm.name)}
											<Check class="mx-auto h-4 w-4 text-green-600" />
										{:else}
											<Minus class="mx-auto h-4 w-4 text-muted-foreground/40" />
										{/if}
									</Table.Cell>
								{/each}
							</Table.Row>
						{/each}
					{/each}
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
</div>
