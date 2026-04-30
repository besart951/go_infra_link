<script lang="ts">
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import type { Component } from 'svelte';
  import { cn } from '$lib/utils.js';

  export type ModuleCardTone = 'user' | 'facility' | 'project' | 'notification';

  export type ModuleCardItem = {
    title: string;
    description: string;
    href: string;
    icon: Component;
    tone: ModuleCardTone | string;
  };

  const toneClasses: Record<ModuleCardTone, string> = {
    user: 'module-card-tone-user',
    facility: 'module-card-tone-facility',
    project: 'module-card-tone-project',
    notification: 'module-card-tone-notification'
  };

  let {
    items,
    emptyMessage,
    columns = 'xl:grid-cols-4'
  }: {
    items: ModuleCardItem[];
    emptyMessage?: string;
    columns?: string;
  } = $props();
</script>

{#if items.length > 0}
  <div class={cn('grid gap-3 sm:grid-cols-2', columns)}>
    {#each items as item (item.href)}
      {@const toneClass = toneClasses[item.tone as ModuleCardTone] ?? toneClasses.project}
      <a
        href={item.href}
        class={cn(
          'group relative overflow-hidden rounded-lg border border-[color:var(--module-card-border)] bg-card p-4 text-card-foreground shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-[color:var(--module-card-border-hover)] hover:bg-[var(--module-card-surface-hover)] hover:shadow-md focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-none',
          toneClass
        )}
      >
        <span class="absolute inset-x-0 top-0 h-0.5 bg-[var(--module-card-stripe)]"></span>
        <span class="flex items-start gap-3">
          <span
            class="flex size-11 shrink-0 items-center justify-center rounded-md bg-[var(--module-card-icon-bg)] text-[color:var(--module-card-icon-fg)] ring-1 ring-[color:var(--module-card-icon-ring)] transition-transform group-hover:animate-in group-hover:animation-duration-300 group-hover:zoom-in-95 group-hover:spin-in-6"
          >
            <item.icon class="size-5" />
          </span>
          <span class="min-w-0 flex-1 space-y-1">
            <span class="block text-sm leading-5 font-semibold">{item.title}</span>
            <span class="block text-sm leading-5 text-muted-foreground">{item.description}</span>
          </span>
          <ChevronRightIcon
            class="mt-1 size-4 shrink-0 text-[color:var(--module-card-accent)] transition-transform duration-200 group-hover:translate-x-0.5"
          />
        </span>
      </a>
    {/each}
  </div>
{:else if emptyMessage}
  <div class="rounded-lg border bg-muted/30 p-4 text-sm text-muted-foreground">
    {emptyMessage}
  </div>
{/if}

<style>
  :global(.module-card-tone-user) {
    --module-card-border: var(--hub-user-border);
    --module-card-border-hover: var(--hub-user-border-hover);
    --module-card-surface-hover: var(--hub-user-surface-hover);
    --module-card-stripe: var(--hub-user-stripe);
    --module-card-icon-bg: var(--hub-user-icon-bg);
    --module-card-icon-fg: var(--hub-user-icon-fg);
    --module-card-icon-ring: var(--hub-user-icon-ring);
    --module-card-accent: var(--hub-user-accent);
  }

  :global(.module-card-tone-facility) {
    --module-card-border: var(--hub-facility-border);
    --module-card-border-hover: var(--hub-facility-border-hover);
    --module-card-surface-hover: var(--hub-facility-surface-hover);
    --module-card-stripe: var(--hub-facility-stripe);
    --module-card-icon-bg: var(--hub-facility-icon-bg);
    --module-card-icon-fg: var(--hub-facility-icon-fg);
    --module-card-icon-ring: var(--hub-facility-icon-ring);
    --module-card-accent: var(--hub-facility-accent);
  }

  :global(.module-card-tone-project) {
    --module-card-border: var(--hub-project-border);
    --module-card-border-hover: var(--hub-project-border-hover);
    --module-card-surface-hover: var(--hub-project-surface-hover);
    --module-card-stripe: var(--hub-project-stripe);
    --module-card-icon-bg: var(--hub-project-icon-bg);
    --module-card-icon-fg: var(--hub-project-icon-fg);
    --module-card-icon-ring: var(--hub-project-icon-ring);
    --module-card-accent: var(--hub-project-accent);
  }

  :global(.module-card-tone-notification) {
    --module-card-border: var(--hub-notification-border);
    --module-card-border-hover: var(--hub-notification-border-hover);
    --module-card-surface-hover: var(--hub-notification-surface-hover);
    --module-card-stripe: var(--hub-notification-stripe);
    --module-card-icon-bg: var(--hub-notification-icon-bg);
    --module-card-icon-fg: var(--hub-notification-icon-fg);
    --module-card-icon-ring: var(--hub-notification-icon-ring);
    --module-card-accent: var(--hub-notification-accent);
  }
</style>
