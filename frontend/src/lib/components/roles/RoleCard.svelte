<script lang="ts">
	import type { Role } from '$lib/domain/role/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import RoleBadge from '$lib/components/role-badge.svelte';
	import { getRoleLabel } from '$lib/utils/permissions.js';
	import { Settings2, Shield, Users, ChevronRight } from '@lucide/svelte';

	interface Props {
		role: Role;
		onEdit?: (role: Role) => void;
		onViewPermissions?: (role: Role) => void;
		canEdit?: boolean;
	}

	let { role, onEdit, onViewPermissions, canEdit = false }: Props = $props();

	const canManageCount = $derived(role.can_manage.length);
</script>

<Card.Root class="group transition-shadow hover:shadow-md">
	<Card.Header class="pb-3">
		<div class="flex items-start justify-between">
			<div class="space-y-1">
				<div class="flex items-center gap-2">
					<RoleBadge role={role.name} />
					<Badge variant="outline" class="text-xs">Level {role.level}</Badge>
				</div>
				<p class="text-sm text-muted-foreground">{role.description}</p>
			</div>
			{#if canEdit && onEdit}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<Button
							variant="ghost"
							size="icon"
							class="h-8 w-8 opacity-0 transition-opacity group-hover:opacity-100"
							onclick={() => onEdit(role)}
						>
							<Settings2 class="h-4 w-4" />
						</Button>
					</Tooltip.Trigger>
					<Tooltip.Content>Edit permissions</Tooltip.Content>
				</Tooltip.Root>
			{/if}
		</div>
	</Card.Header>

	<Card.Content class="space-y-4">
		<!-- Permissions Summary -->
		<div class="space-y-2">
			<div class="flex items-center gap-2 text-sm font-medium text-muted-foreground">
				<Shield class="h-4 w-4" />
				<span>Permissions</span>
			</div>
			<div class="flex flex-wrap gap-1">
				{#if role.permissions.length === 0}
					<span class="text-xs text-muted-foreground">No permissions assigned</span>
				{:else if role.permissions.length <= 4}
					{#each role.permissions as perm}
						<Badge variant="outline" class="text-xs">{perm}</Badge>
					{/each}
				{:else}
					{#each role.permissions.slice(0, 3) as perm}
						<Badge variant="outline" class="text-xs">{perm}</Badge>
					{/each}
					<Badge variant="secondary" class="text-xs">
						+{role.permissions.length - 3} more
					</Badge>
				{/if}
			</div>
		</div>

		<!-- Can Manage Summary -->
		<div class="space-y-2">
			<div class="flex items-center gap-2 text-sm font-medium text-muted-foreground">
				<Users class="h-4 w-4" />
				<span>Can Manage</span>
			</div>
			<div class="flex flex-wrap gap-1">
				{#if canManageCount === 0}
					<span class="text-xs text-muted-foreground">No roles</span>
				{:else if canManageCount <= 3}
					{#each role.can_manage as managedRole}
						<Badge variant="secondary" class="text-xs">{getRoleLabel(managedRole)}</Badge>
					{/each}
				{:else}
					{#each role.can_manage.slice(0, 2) as managedRole}
						<Badge variant="secondary" class="text-xs">{getRoleLabel(managedRole)}</Badge>
					{/each}
					<Badge variant="secondary" class="text-xs">
						+{canManageCount - 2} more
					</Badge>
				{/if}
			</div>
		</div>
	</Card.Content>

	{#if onViewPermissions}
		<Card.Footer class="pt-0">
			<Button
				variant="ghost"
				size="sm"
				class="w-full justify-between"
				onclick={() => onViewPermissions(role)}
			>
				View all permissions
				<ChevronRight class="h-4 w-4" />
			</Button>
		</Card.Footer>
	{/if}
</Card.Root>
