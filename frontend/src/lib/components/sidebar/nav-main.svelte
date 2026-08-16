<script lang="ts">
  import * as Collapsible from '$lib/components/ui/collapsible/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import { cn } from '$lib/utils.js';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import { tick, type Component } from 'svelte';

  const COLLAPSED_MENU_CLOSE_DELAY_MS = 180;

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

  let collapsedMenuOpen = $state<Record<string, boolean>>({});
  let collapsedMenuAnchors = $state<Record<string, HTMLButtonElement | null>>({});
  let collapsedMenuContent = $state<Record<string, HTMLDivElement | null>>({});
  let pointerInsideCollapsedMenu: string | undefined;
  let pointerFocusedMenu: string | undefined;
  let closeCollapsedMenuTimeout: ReturnType<typeof setTimeout> | undefined;

  function closeMobileSidebar() {
    if (sidebar.isMobile) {
      sidebar.setOpenMobile(false);
    }
  }

  function getCollapsedMenuId(item: NavItem) {
    return `collapsed-nav-menu-${item.url.replace(/[^a-z0-9]+/gi, '-')}`;
  }

  function clearCollapsedMenuCloseTimeout() {
    if (closeCollapsedMenuTimeout === undefined) return;
    clearTimeout(closeCollapsedMenuTimeout);
    closeCollapsedMenuTimeout = undefined;
  }

  function openCollapsedMenu(title: string) {
    clearCollapsedMenuCloseTimeout();

    for (const menuTitle of Object.keys(collapsedMenuOpen)) {
      if (menuTitle !== title) {
        collapsedMenuOpen[menuTitle] = false;
      }
    }

    collapsedMenuOpen[title] = true;
  }

  function isFocusInsideCollapsedMenu(title: string) {
    const activeElement = document.activeElement;
    const anchor = collapsedMenuAnchors[title];
    const content = collapsedMenuContent[title];

    return activeElement === anchor || content?.contains(activeElement) === true;
  }

  function scheduleCollapsedMenuClose(title: string) {
    clearCollapsedMenuCloseTimeout();
    closeCollapsedMenuTimeout = setTimeout(() => {
      if (pointerInsideCollapsedMenu === title || isFocusInsideCollapsedMenu(title)) {
        return;
      }

      collapsedMenuOpen[title] = false;
    }, COLLAPSED_MENU_CLOSE_DELAY_MS);
  }

  function handleCollapsedMenuPointerEnter(title: string) {
    pointerInsideCollapsedMenu = title;
    openCollapsedMenu(title);
  }

  function handleCollapsedMenuPointerLeave(title: string) {
    if (pointerInsideCollapsedMenu === title) {
      pointerInsideCollapsedMenu = undefined;
    }
    scheduleCollapsedMenuClose(title);
  }

  function handleCollapsedMenuFocus(title: string) {
    if (pointerFocusedMenu === title) {
      pointerFocusedMenu = undefined;
      return;
    }

    openCollapsedMenu(title);
  }

  async function focusCollapsedMenuContent(title: string) {
    openCollapsedMenu(title);
    await tick();
    collapsedMenuContent[title]?.focus();
  }

  function handleCollapsedMenuKeydown(event: KeyboardEvent, title: string) {
    if (event.key !== 'ArrowDown') return;

    event.preventDefault();
    void focusCollapsedMenuContent(title);
  }
</script>

<Sidebar.Group>
  <Sidebar.GroupLabel>Platform</Sidebar.GroupLabel>
  <Sidebar.Menu>
    {#each items as item (item.title)}
      {#if item.items && item.items.length > 0}
        <Sidebar.MenuItem>
          {#if showCollapsedMenu}
            <DropdownMenu.Root
              open={collapsedMenuOpen[item.title] ?? false}
              onOpenChange={(open) => (collapsedMenuOpen[item.title] = open)}
            >
              <Sidebar.MenuButton
                bind:ref={collapsedMenuAnchors[item.title]}
                isActive={item.isActive}
                aria-controls={getCollapsedMenuId(item)}
                aria-expanded={collapsedMenuOpen[item.title] ?? false}
                aria-haspopup="menu"
                class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                onpointerenter={() => handleCollapsedMenuPointerEnter(item.title)}
                onpointerleave={() => handleCollapsedMenuPointerLeave(item.title)}
                onpointerdown={() => (pointerFocusedMenu = item.title)}
                onfocus={() => handleCollapsedMenuFocus(item.title)}
                onblur={() => scheduleCollapsedMenuClose(item.title)}
                onkeydown={(event) => handleCollapsedMenuKeydown(event, item.title)}
              >
                {#if item.icon}
                  <item.icon />
                {/if}
                <span class="sr-only">{item.title}</span>
              </Sidebar.MenuButton>
              <DropdownMenu.Content
                bind:ref={collapsedMenuContent[item.title]}
                id={getCollapsedMenuId(item)}
                customAnchor={collapsedMenuAnchors[item.title]}
                class="w-[17.5rem] overflow-hidden rounded-xl border-sidebar-border/80 bg-sidebar p-1.5 text-sidebar-foreground shadow-xl shadow-black/25"
                side="right"
                align="start"
                sideOffset={10}
                onpointerenter={() => handleCollapsedMenuPointerEnter(item.title)}
                onpointerleave={() => handleCollapsedMenuPointerLeave(item.title)}
                onfocusin={clearCollapsedMenuCloseTimeout}
                onfocusout={() => scheduleCollapsedMenuClose(item.title)}
              >
                <DropdownMenu.Label
                  class="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-semibold text-sidebar-foreground"
                >
                  <span
                    class="flex size-7 items-center justify-center rounded-md bg-sidebar-accent text-sidebar-accent-foreground"
                  >
                    {#if item.icon}
                      <item.icon class="size-4" />
                    {/if}
                  </span>
                  <span>{item.title}</span>
                </DropdownMenu.Label>
                <DropdownMenu.Separator class="my-1.5 bg-sidebar-border/70" />
                {#each item.items as subItem, index (subItem.title)}
                  <DropdownMenu.Item
                    class={cn(
                      'min-h-11 cursor-pointer gap-3 rounded-lg px-2.5 py-2 text-sm font-medium text-sidebar-foreground transition-colors',
                      subItem.isActive
                        ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                        : 'data-highlighted:bg-sidebar-accent data-highlighted:text-sidebar-accent-foreground'
                    )}
                  >
                    {#snippet child({ props })}
                      <a
                        href={subItem.url}
                        aria-current={subItem.isActive ? 'page' : undefined}
                        {...props}
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
                    {/snippet}
                  </DropdownMenu.Item>
                  {#if subItem.dividerAfter && index < item.items.length - 1}
                    <DropdownMenu.Separator class="my-1.5 bg-sidebar-border/70" />
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
