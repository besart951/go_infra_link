<script lang="ts">
  import { onMount } from 'svelte';
  import AppSidebar from '$lib/components/app-sidebar.svelte';
  import { NotificationBell } from '$lib/components/notifications/index.js';
  import Toasts from '$lib/components/toast.svelte';
  import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
  import { Separator } from '$lib/components/ui/separator/index.js';
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import { page } from '$app/stores';
  import type { LayoutData } from './$types.js';
  import { loadAuth } from '$lib/stores/auth.svelte.js';
  import { goto } from '$app/navigation';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { initNetworkStatus, networkStatus } from '$lib/stores/network.js';
  import { initAppearance, setCurrentAppearanceUserId } from '$lib/stores/appearance.js';
  import { getBreadcrumbForPath } from '$lib/navigation/appNavigation.js';

  const translator = createTranslator();

  const { children, data } = $props<{ children?: any; data: LayoutData }>();

  onMount(async () => {
    initNetworkStatus();
    await loadAuth();
    setCurrentAppearanceUserId(data.user?.id ?? null);
    initAppearance(data.user?.id ?? null);
  });

  $effect(() => {
    if (!data.user && data.backendAvailable !== false) {
      goto('/login');
    }
  });

  const breadcrumbForPath = (pathname: string) => getBreadcrumbForPath(pathname, $translator);
</script>

{#if data.user}
  <Sidebar.Provider>
    <AppSidebar user={data.user} teams={data.teams ?? []} projects={data.projects ?? []} />
    <Sidebar.Inset>
      <header class="flex h-16 shrink-0 items-center justify-between gap-2">
        <div class="flex min-w-0 items-center gap-2 px-4">
          <Sidebar.Trigger class="-ms-1" />
          <Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
          <Breadcrumb.Root>
            <Breadcrumb.List>
              {@const breadcrumb = breadcrumbForPath($page.url.pathname)}
              <Breadcrumb.Item class="hidden md:block">
                <Breadcrumb.Link href="/">{$translator('app.brand')}</Breadcrumb.Link>
              </Breadcrumb.Item>
              <Breadcrumb.Separator class="hidden md:block" />
              {#if breadcrumb.parent}
                <Breadcrumb.Item class="hidden md:block">
                  <Breadcrumb.Link href={breadcrumb.parent.href}>
                    {breadcrumb.parent.title}
                  </Breadcrumb.Link>
                </Breadcrumb.Item>
                <Breadcrumb.Separator class="hidden md:block" />
              {/if}
              <Breadcrumb.Item>
                <Breadcrumb.Page>{breadcrumb.current}</Breadcrumb.Page>
              </Breadcrumb.Item>
            </Breadcrumb.List>
          </Breadcrumb.Root>
        </div>
        <div class="flex items-center gap-2 px-4">
          <NotificationBell />
        </div>
      </header>
      {#if !$networkStatus.browserOnline}
        <div class="px-4">
          <div
            class="rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-900"
          >
            {$translator('pages.connection.offline')}
          </div>
        </div>
      {:else if $networkStatus.retrying}
        <div class="px-4">
          <div
            class="rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-900"
          >
            {$translator('pages.connection.retrying', {
              attempt: $networkStatus.retryAttempt,
              max: $networkStatus.retryMax
            })}
          </div>
        </div>
      {:else if $networkStatus.apiUnavailable || data.backendAvailable === false}
        <div class="px-4">
          <div
            class="rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-900"
          >
            {$translator('pages.connection.backend_unavailable_stale')}
          </div>
        </div>
      {/if}
      <div class="flex min-h-0 min-w-0 flex-1 flex-col gap-4 overflow-y-auto p-4 pt-0">
        {@render children?.()}
      </div>
    </Sidebar.Inset>
    <Toasts />
  </Sidebar.Provider>
{:else if data.backendAvailable === false}
  <div class="flex h-screen w-full items-center justify-center p-4">
    <div class="w-full max-w-md rounded-lg border bg-card p-6 shadow-sm">
      <h2 class="mb-2 text-lg font-semibold text-destructive">
        {$translator('pages.backend_unavailable_title')}
      </h2>
      <p class="text-sm text-muted-foreground">
        {$translator('pages.backend_unavailable_desc')}
      </p>
    </div>
  </div>
{/if}
