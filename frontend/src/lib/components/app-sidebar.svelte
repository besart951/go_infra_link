<script lang="ts">
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
	import { page } from "$app/stores";
	import { goto } from "$app/navigation";
	import { Building2, FolderKanban, LogOut, Settings, UserCircle, Users } from "@lucide/svelte";

	const items = [
		{ label: "Dashboard", href: "/", icon: Building2 },
		{ label: "Projects", href: "/projects", icon: FolderKanban },
		{ label: "Teams", href: "/teams", icon: Users },
		{ label: "Users", href: "/users", icon: UserCircle },
	] as const;
</script>

<Sidebar.Root collapsible="icon">
	<Sidebar.Header class="px-2 py-2">
		<div class="flex items-center gap-2 px-2">
			<div
				class="bg-primary text-primary-foreground flex size-8 items-center justify-center rounded-md font-semibold"
			>
				IL
			</div>
			<div class="flex flex-col leading-tight">
				<span class="text-sm font-semibold">Infra Link</span>
				<span class="text-muted-foreground text-xs">Console</span>
			</div>
		</div>
	</Sidebar.Header>

	<Sidebar.Content>
		<Sidebar.Menu>
			{#each items as item (item.href)}
				<Sidebar.MenuItem>
					<Sidebar.MenuButton
						isActive={$page.url.pathname === item.href || $page.url.pathname.startsWith(item.href + "/")}
						onclick={() => goto(item.href)}
						tooltipContent={item.label}
					>
						<svelte:component this={item.icon} />
						<span>{item.label}</span>
					</Sidebar.MenuButton>
				</Sidebar.MenuItem>
			{/each}
		</Sidebar.Menu>
	</Sidebar.Content>

	<Sidebar.Footer>
		<Sidebar.Menu>
			<Sidebar.MenuItem>
				<Sidebar.MenuButton
					isActive={$page.url.pathname === "/settings" || $page.url.pathname.startsWith("/settings/")}
					onclick={() => goto("/settings")}
					tooltipContent="Settings"
				>
					<Settings />
					<span>Settings</span>
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
			<Sidebar.MenuItem>
				<Sidebar.MenuButton onclick={() => goto("/logout")} tooltipContent="Logout">
					<LogOut />
					<span>Logout</span>
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
	</Sidebar.Footer>

	<Sidebar.Rail />
</Sidebar.Root>
