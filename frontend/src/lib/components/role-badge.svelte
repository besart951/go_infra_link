<script lang="ts">
	import type { UserRole } from '$lib/api/users';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { getRoleLabel } from '$lib/utils/permissions';
	import { ShieldCheck, UserCircle } from '@lucide/svelte';

	interface Props {
		role: UserRole;
		showIcon?: boolean;
	}

	let { role, showIcon = true }: Props = $props();

	const ADMIN_ROLES: UserRole[] = ['superadmin', 'admin_fzag', 'admin_planer', 'admin_entrepreneur'];

	function getVariant(r: UserRole): 'default' | 'secondary' | 'outline' {
		if (r === 'superadmin') return 'default';
		if (r.startsWith('admin_')) return 'secondary';
		return 'outline';
	}

	function isAdmin(r: UserRole): boolean {
		return ADMIN_ROLES.includes(r);
	}
</script>

<Badge variant={getVariant(role)}>
	{#if showIcon}
		{#if isAdmin(role)}
			<ShieldCheck class="mr-1 h-3 w-3" />
		{:else}
			<UserCircle class="mr-1 h-3 w-3" />
		{/if}
	{/if}
	{getRoleLabel(role)}
</Badge>
