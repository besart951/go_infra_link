<script lang="ts">
	/**
	 * PermissionGuard Component
	 *
	 * Conditionally renders children based on user permissions.
	 *
	 * Usage:
	 *   <PermissionGuard action="create" resource="user">
	 *     <button>Create User</button>
	 *   </PermissionGuard>
	 *
	 *   <PermissionGuard canManageRole="entrepreneur">
	 *     <button>Assign Entrepreneur Role</button>
	 *   </PermissionGuard>
	 */

	import type { Snippet } from 'svelte';
	import type { UserRole } from '$lib/api/users.js';
	import { canPerform } from '$lib/utils/permissions.js';
	import { canManageRole } from '$lib/stores/auth.svelte';

	interface Props {
		children: Snippet;
		action?: string;
		resource?: string;
		canManageRole?: UserRole;
		fallback?: Snippet;
	}

	let { children, action, resource, canManageRole: targetRole, fallback }: Props = $props();

	// Compute permission check
	const hasPermission = $derived(() => {
		if (targetRole) {
			return canManageRole(targetRole);
		}
		if (action && resource) {
			return canPerform(action, resource);
		}
		return false;
	});
</script>

{#if hasPermission()}
	{@render children()}
{:else if fallback}
	{@render fallback()}
{/if}
