<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/index.js';
	import { theme } from '$lib/stores/theme.js';
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import SunIcon from '@lucide/svelte/icons/sun';
	import MonitorIcon from '@lucide/svelte/icons/monitor';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import UserIcon from '@lucide/svelte/icons/user';
	import { goto } from '$app/navigation';
	import type { User } from '$lib/api/users.js';
	import { onMount } from 'svelte';

	let { user }: { user: User } = $props();
	const sidebar = useSidebar();

	const logout = async () => {
		await goto('/api/auth/logout');
	};

	onMount(() => {
		theme.init();
	});
</script>

<Sidebar.Menu>
	<Sidebar.MenuItem>
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						size="lg"
						class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
						{...props}
					>
						<Avatar.Root class="size-8 rounded-lg">
							<Avatar.Fallback class="rounded-lg"
								>{user.first_name[0]}{user.last_name[0]}</Avatar.Fallback
							>
						</Avatar.Root>
						<div class="grid flex-1 text-start text-sm leading-tight">
							<span class="truncate font-medium">{user.first_name} {user.last_name}</span>
							<span class="truncate text-xs">{user.email}</span>
						</div>
						<ChevronsUpDownIcon class="ms-auto size-4" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
				side={sidebar.isMobile ? 'bottom' : 'right'}
				align="end"
				sideOffset={4}
			>
				<DropdownMenu.Label class="p-0 font-normal">
					<div class="flex items-center gap-2 px-1 py-1.5 text-start text-sm">
						<Avatar.Root class="size-8 rounded-lg">
							<Avatar.Fallback class="rounded-lg"
								>{user.first_name[0]}{user.last_name[0]}</Avatar.Fallback
							>
						</Avatar.Root>
						<div class="grid flex-1 text-start text-sm leading-tight">
							<span class="truncate font-medium">{user.first_name} {user.last_name}</span>
							<span class="truncate text-xs">{user.email}</span>
						</div>
					</div>
				</DropdownMenu.Label>
				<DropdownMenu.Separator />
				<DropdownMenu.Group>
					<DropdownMenu.Item>
						<button onclick={() => theme.setTheme('light')} class="flex w-full items-center">
							<SunIcon class="size-4" />
							<span class="ml-2">Light</span>
						</button>
					</DropdownMenu.Item>
					<DropdownMenu.Item>
						<button onclick={() => theme.setTheme('dark')} class="flex w-full items-center">
							<MoonIcon class="size-4" />
							<span class="ml-2">Dark</span>
						</button>
					</DropdownMenu.Item>
					<DropdownMenu.Item>
						<button onclick={() => theme.setTheme('system')} class="flex w-full items-center">
							<MonitorIcon class="size-4" />
							<span class="ml-2">System</span>
						</button>
					</DropdownMenu.Item>
				</DropdownMenu.Group>
				<DropdownMenu.Separator />
				<DropdownMenu.Group>
					<DropdownMenu.Item>
						<button onclick={() => goto('/account')} class="flex w-full items-center">
							<UserIcon class="size-4" />
							<span class="ml-2">Account</span>
						</button>
					</DropdownMenu.Item>
					<DropdownMenu.Item>
						<button onclick={() => goto('/settings')} class="flex w-full items-center">
							<SettingsIcon class="size-4" />
							<span class="ml-2">Settings</span>
						</button>
					</DropdownMenu.Item>
				</DropdownMenu.Group>
				<DropdownMenu.Separator />
				<DropdownMenu.Item>
					<button onclick={logout} class="flex w-full items-center">
						<LogOutIcon class="size-4" />
						<span class="ml-2">Log out</span>
					</button>
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>