<script lang="ts">
	import * as Collapsible from '$lib/components/ui/collapsible/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import type { Component } from 'svelte';

	interface NavItem {
		title: string;
		url: string;
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		icon?: any;
		isActive?: boolean;
		items?: { title: string; url: string }[];
	}

	let { items }: { items: NavItem[] } = $props();
	const sidebar = Sidebar.useSidebar();

	function closeMobileSidebar() {
		if (sidebar.isMobile) {
			sidebar.setOpenMobile(false);
		}
	}
</script>

<Sidebar.Group>
	<Sidebar.GroupLabel>Platform</Sidebar.GroupLabel>
	<Sidebar.Menu>
		{#each items as item (item.title)}
			{#if item.items && item.items.length > 0}
				<Collapsible.Root open={item.isActive} class="group/collapsible">
					<Sidebar.MenuItem>
						<Collapsible.Trigger>
							{#snippet child({ props })}
								<Sidebar.MenuButton {...props} tooltipContent={item.title}>
									{#if item.icon}
										<item.icon />
									{/if}
									<span>{item.title}</span>
									<ChevronRightIcon
										class="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"
									/>
								</Sidebar.MenuButton>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content>
							<Sidebar.MenuSub>
								{#each item.items as subItem (subItem.title)}
									<Sidebar.MenuSubItem>
										<Sidebar.MenuSubButton href={subItem.url} onclick={closeMobileSidebar}>
											<span>{subItem.title}</span>
										</Sidebar.MenuSubButton>
									</Sidebar.MenuSubItem>
								{/each}
							</Sidebar.MenuSub>
						</Collapsible.Content>
					</Sidebar.MenuItem>
				</Collapsible.Root>
			{:else}
				<Sidebar.MenuItem>
					<Sidebar.MenuButton isActive={item.isActive} tooltipContent={item.title}>
						{#snippet child({ props })}
							<a href={item.url} onclick={closeMobileSidebar} {...props}>
								{#if item.icon}
									<item.icon />
								{/if}
								<span>{item.title}</span>
							</a>
						{/snippet}
					</Sidebar.MenuButton>
				</Sidebar.MenuItem>
			{/if}
		{/each}
	</Sidebar.Menu>
</Sidebar.Group>
