<script lang="ts">
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import LinkIcon from '@lucide/svelte/icons/link';
  import type { components } from '$lib/api/generated/schema.js';
  import type {
    FacilityDetailKind,
    FacilityDetailScope
  } from '$lib/services/facilityDetailCache.js';

  type Relation =
    components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.DetailRelation'];

  let {
    relations = [],
    scope = {},
    onPageChange
  }: {
    relations?: Relation[];
    scope?: FacilityDetailScope;
    onPageChange?: (page: number) => void;
  } = $props();

  const routeKinds: Record<string, FacilityDetailKind | undefined> = {
    buildings: 'buildings',
    'control-cabinets': 'control-cabinets',
    'sps-controllers': 'sps-controllers',
    'sps-controller-system-types': 'sps-controller-system-types',
    'field-devices': 'field-devices'
  };

  function href(resource: string | undefined, id: string | undefined): string | undefined {
    const kind = resource ? routeKinds[resource] : undefined;
    if (!kind || !id) return undefined;
    const prefix = scope.projectId ? `/projects/${scope.projectId}/facility` : '/facility';
    return `${prefix}/${kind}/${id}`;
  }
</script>

{#if relations.length > 0}
  <section class="space-y-3" aria-label="Verknüpfte Elemente">
    {#each relations as relation (`${relation.key}-${relation.resource}`)}
      <Card class="overflow-hidden">
        <CardHeader class="flex-row items-center justify-between gap-3 space-y-0 py-4">
          <CardTitle class="text-base">{relation.label}</CardTitle>
          <Badge variant="secondary">{relation.count ?? 0}</Badge>
        </CardHeader>
        <CardContent class="space-y-1 pb-4">
          {#if (relation.items?.length ?? 0) === 0}
            <p class="px-3 py-2 text-sm text-muted-foreground">Keine verknüpften Elemente.</p>
          {:else}
            {#each relation.items ?? [] as item (item.id)}
              {@const itemHref = href(relation.resource, item.id)}
              {#if itemHref}
                <a
                  class="group flex items-center gap-3 rounded-md px-3 py-2 transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
                  href={itemHref}
                >
                  <LinkIcon class="size-4 shrink-0 text-muted-foreground" />
                  <span class="min-w-0 flex-1">
                    <span class="block truncate text-sm font-medium">{item.label}</span>
                    {#if item.subtitle}
                      <span class="block truncate text-xs text-muted-foreground"
                        >{item.subtitle}</span
                      >
                    {/if}
                  </span>
                  <ChevronRight
                    class="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5"
                  />
                </a>
              {:else}
                <div class="flex items-center gap-3 rounded-md px-3 py-2">
                  <LinkIcon class="size-4 shrink-0 text-muted-foreground" />
                  <span class="min-w-0 flex-1">
                    <span class="block truncate text-sm font-medium">{item.label}</span>
                    {#if item.subtitle}
                      <span class="block truncate text-xs text-muted-foreground"
                        >{item.subtitle}</span
                      >
                    {/if}
                  </span>
                </div>
              {/if}
            {/each}
          {/if}
          {#if (relation.count ?? 0) > (relation.items?.length ?? 0)}
            <div
              class="flex items-center justify-between gap-2 px-3 pt-2 text-xs text-muted-foreground"
            >
              <span>Seite {relation.page ?? 1} von {relation.total_pages ?? 1}</span>
              {#if (relation.total_pages ?? 1) > 1}
                <span class="flex items-center gap-1">
                  <Button
                    size="sm"
                    variant="ghost"
                    disabled={(relation.page ?? 1) <= 1}
                    onclick={() => onPageChange?.((relation.page ?? 1) - 1)}>Zurück</Button
                  >
                  <Button
                    size="sm"
                    variant="ghost"
                    disabled={(relation.page ?? 1) >= (relation.total_pages ?? 1)}
                    onclick={() => onPageChange?.((relation.page ?? 1) + 1)}>Weiter</Button
                  >
                </span>
              {/if}
            </div>
          {/if}
        </CardContent>
      </Card>
    {/each}
  </section>
{/if}
