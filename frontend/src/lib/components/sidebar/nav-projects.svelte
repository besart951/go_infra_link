<script lang="ts">
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/context.svelte.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import FolderIcon from '@lucide/svelte/icons/folder';
	import ForwardIcon from '@lucide/svelte/icons/forward';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import type { Component } from 'svelte';

	interface ProjectItem {
		id: string;
		name: string;
		url: string;
		icon?: Component;
		status?: string;
	}

	let {
		projects,
		onViewProject,
		onShareProject,
		onCreate
	}: {
		projects: ProjectItem[];
		onViewProject?: (id: string) => void;
		onShareProject?: (id: string) => void;
		onCreate?: () => void;
	} = $props();

	const sidebar = useSidebar();

	function closeMobileSidebar() {
		if (sidebar.isMobile) {
			sidebar.setOpenMobile(false);
		}
	}
</script>

<Sidebar.Group class="group-data-[collapsible=icon]:hidden">
	<Sidebar.GroupLabel>Projects</Sidebar.GroupLabel>
	<Sidebar.Menu>
		{#each projects as item (item.id)}
			<Sidebar.MenuItem>
				<Sidebar.MenuButton>
					{#snippet child({ props })}
						<a href={item.url} onclick={closeMobileSidebar} {...props}>
							{#if item.icon}
								<item.icon />
							{:else}
								<FolderIcon />
							{/if}
							<span>{item.name}</span>
						</a>
					{/snippet}
				</Sidebar.MenuButton>
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						{#snippet child({ props })}
							<Sidebar.MenuAction showOnHover {...props}>
								<EllipsisIcon />
								<span class="sr-only">More</span>
							</Sidebar.MenuAction>
						{/snippet}
					</DropdownMenu.Trigger>
					<DropdownMenu.Content
						class="w-48 rounded-lg"
						side={sidebar.isMobile ? 'bottom' : 'right'}
						align={sidebar.isMobile ? 'end' : 'start'}
					>
						<DropdownMenu.Item onclick={() => onViewProject?.(item.id)}>
							<FolderIcon class="text-muted-foreground" />
							<span>View Project</span>
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => onShareProject?.(item.id)}>
							<ForwardIcon class="text-muted-foreground" />
							<span>Share Project</span>
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Sidebar.MenuItem>
		{/each}
		{#if onCreate}
			<Sidebar.MenuItem>
				<Sidebar.MenuButton class="text-sidebar-foreground/70" onclick={onCreate}>
					<PlusIcon class="text-sidebar-foreground/70" />
					<span>New Project</span>
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		{/if}
	</Sidebar.Menu>
</Sidebar.Group>
