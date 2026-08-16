<script lang="ts">
  import * as Collapsible from '$lib/components/ui/collapsible/index.js';
  import * as NavigationMenu from '$lib/components/ui/navigation-menu/index.js';
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import { cn } from '$lib/utils.js';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import type { Component } from 'svelte';

  interface NavSubItem {
    title: string;
    url: string;
    icon: Component;
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

  {#if showCollapsedMenu}
    <NavigationMenu.Root
      viewport={false}
      orientation="vertical"
      class="w-full max-w-none items-start justify-start"
    >
      <NavigationMenu.List class="w-full flex-col items-stretch justify-start gap-1">
        {#each items as item (item.title)}
          <NavigationMenu.Item class="w-full">
            {#if item.items && item.items.length > 0}
              <NavigationMenu.Trigger
                data-active={item.isActive}
                class="peer/menu-button flex size-8! w-full items-center justify-center overflow-hidden rounded-md bg-transparent! p-2! text-start text-sm ring-sidebar-ring outline-hidden transition-[width,height,padding,transform,filter,box-shadow] duration-100 ease-in-out hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-2 active:bg-sidebar-accent active:text-sidebar-accent-foreground data-[active=true]:bg-sidebar-accent data-[active=true]:text-sidebar-accent-foreground data-[state=open]:hover:bg-sidebar-accent data-[state=open]:hover:text-sidebar-accent-foreground [&>svg:first-child]:size-4 [&>svg:first-child]:shrink-0 [&>svg:last-child]:hidden"
              >
                {#if item.icon}
                  <item.icon />
                {/if}
                <span class="sr-only">{item.title}</span>
              </NavigationMenu.Trigger>
              <NavigationMenu.Content
                class="start-full! top-0! z-50 mt-0! w-[17.5rem] overflow-hidden rounded-xl border-sidebar-border/80 bg-sidebar p-1.5 text-sidebar-foreground shadow-xl shadow-black/25"
              >
                <div class="rounded-lg px-2.5 py-2 text-sm font-semibold">
                  {item.title}
                </div>
                <div class="my-1.5 h-px bg-sidebar-border/70"></div>
                <ul class="space-y-1">
                  {#each item.items as subItem, index (subItem.title)}
                    <li>
                      <a
                        href={subItem.url}
                        onclick={closeMobileSidebar}
                        aria-current={subItem.isActive ? 'page' : undefined}
                        class={cn(
                          'flex min-h-11 items-center gap-3 rounded-lg px-2.5 py-2 text-sm font-medium text-sidebar-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:bg-sidebar-accent focus-visible:text-sidebar-accent-foreground focus-visible:outline-none',
                          subItem.isActive && 'bg-sidebar-accent text-sidebar-accent-foreground'
                        )}
                      >
                        <span
                          class={cn(
                            'flex size-7 shrink-0 items-center justify-center rounded-md bg-sidebar-accent/70 text-sidebar-accent-foreground',
                            subItem.isActive && 'bg-primary text-primary-foreground'
                          )}
                        >
                          <subItem.icon class="size-4" />
                        </span>
                        <span>{subItem.title}</span>
                      </a>
                    </li>
                    {#if subItem.dividerAfter && index < item.items.length - 1}
                      <li aria-hidden="true" class="my-1.5 h-px bg-sidebar-border/70"></li>
                    {/if}
                  {/each}
                </ul>
              </NavigationMenu.Content>
            {:else}
              <NavigationMenu.Link
                href={item.url}
                onclick={closeMobileSidebar}
                data-active={item.isActive}
                aria-current={item.isActive ? 'page' : undefined}
                class="peer/menu-button flex size-8! w-full items-center justify-center overflow-hidden rounded-md bg-transparent! p-2! text-start text-sm ring-sidebar-ring outline-hidden transition-[width,height,padding,transform,filter,box-shadow] duration-100 ease-in-out hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-2 data-[active=true]:bg-sidebar-accent data-[active=true]:text-sidebar-accent-foreground [&>svg]:size-4 [&>svg]:shrink-0"
              >
                {#if item.icon}
                  <item.icon />
                {/if}
                <span class="sr-only">{item.title}</span>
              </NavigationMenu.Link>
            {/if}
          </NavigationMenu.Item>
        {/each}
      </NavigationMenu.List>
    </NavigationMenu.Root>
  {:else}
    <Sidebar.Menu>
      {#each items as item (item.title)}
        {#if item.items && item.items.length > 0}
          <Sidebar.MenuItem>
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
  {/if}
</Sidebar.Group>
