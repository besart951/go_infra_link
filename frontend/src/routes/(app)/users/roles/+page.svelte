<script lang="ts">
	import type { UserRole } from '$lib/api/users';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import RoleBadge from '$lib/components/role-badge.svelte';
	import { ROLE_LABELS, ROLE_PERMISSIONS, getRoleLabel } from '$lib/utils/permissions.js';
	import { ROLE_LEVELS } from '$lib/stores/auth.svelte.js';
	import { Check, Minus } from '@lucide/svelte';

	const ALL_ROLES: UserRole[] = [
		'superadmin',
		'admin_fzag',
		'fzag',
		'admin_planer',
		'planer',
		'admin_entrepreneur',
		'entrepreneur'
	];

	// Collect all unique permission strings (excluding wildcard)
	const allPermissions = (() => {
		const set = new Set<string>();
		for (const role of ALL_ROLES) {
			for (const perm of ROLE_PERMISSIONS[role]) {
				if (perm !== '*') set.add(perm);
			}
		}
		return [...set].sort();
	})();

	function hasPermission(role: UserRole, permission: string): boolean {
		const perms = ROLE_PERMISSIONS[role];
		return perms.includes('*') || perms.includes(permission);
	}

	function canManageRoles(role: UserRole): UserRole[] {
		const level = ROLE_LEVELS[role];
		return ALL_ROLES.filter((r) => ROLE_LEVELS[r] < level);
	}
</script>

<div class="flex flex-col gap-8">
	<div>
		<h1 class="text-3xl font-bold tracking-tight">Roles & Permissions</h1>
		<p class="mt-1 text-muted-foreground">
			Overview of the role hierarchy and permission assignments.
		</p>
	</div>

	<!-- Role Hierarchy Cards -->
	<div>
		<h2 class="mb-4 text-xl font-semibold">Role Hierarchy</h2>
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
			{#each ALL_ROLES as role (role)}
				<Card.Root>
					<Card.Header class="pb-3">
						<div class="flex items-center justify-between">
							<RoleBadge {role} />
							<Badge variant="outline" class="text-xs">Level {ROLE_LEVELS[role]}</Badge>
						</div>
					</Card.Header>
					<Card.Content class="space-y-3">
						<div>
							<p class="mb-1 text-xs font-medium text-muted-foreground">Permissions</p>
							<div class="flex flex-wrap gap-1">
								{#if ROLE_PERMISSIONS[role].includes('*')}
									<Badge variant="default" class="text-xs">All permissions</Badge>
								{:else}
									{#each ROLE_PERMISSIONS[role] as perm}
										<Badge variant="outline" class="text-xs">{perm}</Badge>
									{/each}
								{/if}
							</div>
						</div>
						<div>
							<p class="mb-1 text-xs font-medium text-muted-foreground">Can manage</p>
							{#if canManageRoles(role).length === 0}
								<span class="text-xs text-muted-foreground">No roles</span>
							{:else}
								<div class="flex flex-wrap gap-1">
									{#each canManageRoles(role) as r}
										<Badge variant="secondary" class="text-xs">{getRoleLabel(r)}</Badge>
									{/each}
								</div>
							{/if}
						</div>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	</div>

	<!-- Permissions Matrix -->
	<div>
		<h2 class="mb-4 text-xl font-semibold">Permissions Matrix</h2>
		<div class="overflow-x-auto rounded-lg border bg-background">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head class="sticky left-0 bg-background">Role</Table.Head>
						{#each allPermissions as perm}
							<Table.Head class="text-center text-xs whitespace-nowrap">{perm}</Table.Head>
						{/each}
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each ALL_ROLES as role (role)}
						<Table.Row>
							<Table.Cell class="sticky left-0 bg-background font-medium">
								<RoleBadge {role} showIcon={false} />
							</Table.Cell>
							{#each allPermissions as perm}
								<Table.Cell class="text-center">
									{#if hasPermission(role, perm)}
										<Check class="mx-auto h-4 w-4 text-green-600" />
									{:else}
										<Minus class="mx-auto h-4 w-4 text-muted-foreground/40" />
									{/if}
								</Table.Cell>
							{/each}
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	</div>
</div>
