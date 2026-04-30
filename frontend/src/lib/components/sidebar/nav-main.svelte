<script lang="ts">
  import * as Collapsible from '$lib/components/ui/collapsible/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import { cn } from '$lib/utils.js';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import type { Component } from 'svelte';

  interface NavSubItem {
    title: string;
    url: string;
    dividerAfter?: boolean;
    isActive?: boolean;
  }

  interface NavItem {
    title: string;
    url: string;
    icon?: Component;
    isActive?: boolean;
    items?: NavSubItem[];
  }

  let { items }: { items: NavItem[] } = $props();
  const sidebar = Sidebar.useSidebar();
  const showCollapsedMenu = $derived(sidebar.state === 'collapsed' && !sidebar.isMobile);

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
        <Sidebar.MenuItem>
          {#if showCollapsedMenu}
            <DropdownMenu.Root>
              <DropdownMenu.Trigger>
                {#snippet child({ props })}
                  <Sidebar.MenuButton
                    {...props}
                    isActive={item.isActive}
                    class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                  >
                    {#if item.icon}
                      <item.icon />
                    {/if}
                    <span class="sr-only">{item.title}</span>
                  </Sidebar.MenuButton>
                {/snippet}
              </DropdownMenu.Trigger>
              <DropdownMenu.Content
                class="w-56 rounded-lg"
                side="right"
                align="start"
                sideOffset={8}
              >
                <DropdownMenu.Label class="text-xs text-muted-foreground">
                  {item.title}
                </DropdownMenu.Label>
                <DropdownMenu.Separator />
                {#each item.items as subItem, index (subItem.title)}
                  <DropdownMenu.Item
                    class={cn(
                      'cursor-pointer',
                      subItem.isActive && 'bg-accent font-medium text-accent-foreground'
                    )}
                  >
                    {#snippet child({ props })}
                      <a
                        href={subItem.url}
                        aria-current={subItem.isActive ? 'page' : undefined}
                        {...props}
                      >
                        <span>{subItem.title}</span>
                      </a>
                    {/snippet}
                  </DropdownMenu.Item>
                  {#if subItem.dividerAfter && index < item.items.length - 1}
                    <DropdownMenu.Separator />
                  {/if}
                {/each}
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          {:else}
            <Collapsible.Root open={item.isActive} class="group/collapsible">
              <Collapsible.Trigger>
                {#snippet child({ props })}
                  <Sidebar.MenuButton
                    {...props}
                    isActive={item.isActive}
                    tooltipContent={item.title}
                  >
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
                  {#each item.items as subItem, index (subItem.title)}
                    <Sidebar.MenuSubItem>
                      <Sidebar.MenuSubButton
                        href={subItem.url}
                        onclick={closeMobileSidebar}
                        isActive={subItem.isActive}
                        aria-current={subItem.isActive ? 'page' : undefined}
                      >
                        <span>{subItem.title}</span>
                      </Sidebar.MenuSubButton>
                    </Sidebar.MenuSubItem>
                    {#if subItem.dividerAfter && index < item.items.length - 1}
                      <Sidebar.MenuSubItem>
                        <Sidebar.Separator class="my-1" />
                      </Sidebar.MenuSubItem>
                    {/if}
                  {/each}
                </Sidebar.MenuSub>
              </Collapsible.Content>
            </Collapsible.Root>
          {/if}
        </Sidebar.MenuItem>
      {:else}
        <Sidebar.MenuItem>
          <Sidebar.MenuButton isActive={item.isActive} tooltipContent={item.title}>
            {#snippet child({ props })}
              <a
                href={item.url}
                onclick={closeMobileSidebar}
                aria-current={item.isActive ? 'page' : undefined}
                {...props}
              >
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
