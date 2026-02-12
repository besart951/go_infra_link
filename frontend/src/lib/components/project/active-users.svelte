<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import type { UserPresence } from '$lib/stores/websocket.svelte.js';
	import { Users } from 'lucide-svelte';

	interface Props {
		users: UserPresence[];
		maxVisible?: number;
	}

	let { users, maxVisible = 3 }: Props = $props();

	const visibleUsers = $derived(users.slice(0, maxVisible));
	const remainingCount = $derived(Math.max(0, users.length - maxVisible));

	function getInitials(user: UserPresence): string {
		const first = user.first_name?.[0] ?? '';
		const last = user.last_name?.[0] ?? '';
		return `${first}${last}`.toUpperCase();
	}

	function getFullName(user: UserPresence): string {
		return `${user.first_name} ${user.last_name}`.trim();
	}

	function getRandomColor(userId: string): string {
		// Generate a consistent color based on user ID
		const colors = [
			'bg-blue-500',
			'bg-green-500',
			'bg-purple-500',
			'bg-pink-500',
			'bg-orange-500',
			'bg-cyan-500',
			'bg-indigo-500',
			'bg-teal-500'
		];
		const hash = userId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
		return colors[hash % colors.length];
	}
</script>

{#if users.length > 0}
	<div class="flex items-center gap-3">
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<Users class="size-4" />
			<span>{users.length} {users.length === 1 ? 'user' : 'users'} active</span>
		</div>

		<div class="flex -space-x-2">
			{#each visibleUsers as user (user.user_id)}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<Avatar.Root class="size-8 border-2 border-background ring-2 ring-primary/20">
							<Avatar.Fallback class={`text-xs font-semibold text-white ${getRandomColor(user.user_id)}`}>
								{getInitials(user)}
							</Avatar.Fallback>
						</Avatar.Root>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<div class="space-y-1">
							<p class="font-medium">{getFullName(user)}</p>
							<p class="text-xs text-muted-foreground">{user.email}</p>
						</div>
					</Tooltip.Content>
				</Tooltip.Root>
			{/each}

			{#if remainingCount > 0}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<div
							class="flex size-8 items-center justify-center rounded-full border-2 border-background bg-muted text-xs font-semibold ring-2 ring-primary/20"
						>
							+{remainingCount}
						</div>
					</Tooltip.Trigger>
					<Tooltip.Content>
						<div class="space-y-1">
							<p class="font-medium">+{remainingCount} more {remainingCount === 1 ? 'user' : 'users'}</p>
						</div>
					</Tooltip.Content>
				</Tooltip.Root>
			{/if}
		</div>
	</div>
{/if}
