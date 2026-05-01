<script lang="ts">
  import * as Card from '$lib/components/ui/card/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
  import * as ButtonGroup from '$lib/components/ui/button-group/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import CpuIcon from '@lucide/svelte/icons/cpu';
  import NetworkIcon from '@lucide/svelte/icons/network';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import CopyIcon from '@lucide/svelte/icons/copy';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
  import type { SPSController } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useControlCabinetDetailState } from './state/context.svelte.js';

  const state = useControlCabinetDetailState();
  const t = createTranslator();

  function handleCreateClick(): void {
    state.startSpsCreate();
  }

  function createViewHandler(controllerId: string): () => Promise<void> {
    return async function (): Promise<void> {
      await state.goToSpsController(controllerId);
    };
  }

  function createEditHandler(controller: SPSController): () => void {
    return function (): void {
      state.startSpsEdit(controller);
    };
  }

  function createCopyHandler(controller: SPSController): () => Promise<void> {
    return async function (): Promise<void> {
      await state.copySps(controller);
    };
  }

  function createDeleteHandler(controller: SPSController): () => Promise<void> {
    return async function (): Promise<void> {
      await state.deleteSps(controller);
    };
  }

  function systemTypeSummary(systemType: {
    number?: number | null;
    document_name?: string | null;
  }): string {
    if (systemType.number != null && systemType.document_name) {
      return `${$t('facility.control_cabinet_detail.number')}: ${systemType.number} • ${systemType.document_name}`;
    }

    if (systemType.number != null) {
      return `${$t('facility.control_cabinet_detail.number')}: ${systemType.number}`;
    }

    return systemType.document_name ?? '';
  }
</script>

<Card.Root class="border-primary/20 bg-card">
  <Card.Header>
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <Card.Title class="flex items-center gap-2">
          <CpuIcon class="size-5 text-primary" />
          {$t('facility.control_cabinet_detail.sps_title')}
        </Card.Title>
        <Card.Description>{$t('facility.control_cabinet_detail.sps_desc')}</Card.Description>
      </div>
      {#if state.canCreateSps}
        <Button size="sm" onclick={handleCreateClick}>
          <PlusIcon class="mr-2 size-4" />
          {$t('facility.control_cabinet_detail.create_sps')}
        </Button>
      {/if}
    </div>
  </Card.Header>

  <Card.Content>
    <Tooltip.Provider>
      {#if !state.canReadSps}
        <div
          class="rounded-md border border-border bg-muted/30 px-3 py-2 text-sm text-muted-foreground"
        >
          {$t('facility.control_cabinet_detail.no_read_permission')}
        </div>
      {:else if state.controllers.length === 0}
        <div
          class="rounded-md border border-dashed border-border px-4 py-8 text-center text-sm text-muted-foreground"
        >
          {$t('facility.control_cabinet_detail.no_sps')}
        </div>
      {:else}
        <div class="space-y-3">
          {#each state.controllers as controller (controller.id)}
            <div class="rounded-lg border border-border/70 bg-muted/20 p-4">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p class="text-base font-semibold text-foreground">{controller.device_name}</p>
                  <div class="mt-1 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                    <Badge variant="secondary">GA: {controller.ga_device ?? '-'}</Badge>
                    <Badge variant="outline">IP: {controller.ip_address ?? '-'}</Badge>
                  </div>
                </div>

                <ButtonGroup.Root class="shrink-0">
                  <Tooltip.Root>
                    <Tooltip.Trigger
                      class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
                      onclick={createViewHandler(controller.id)}
                      aria-label={$t('facility.view')}
                    >
                      <ArrowRightIcon class="size-4" />
                    </Tooltip.Trigger>
                    <Tooltip.Content>{$t('facility.view')}</Tooltip.Content>
                  </Tooltip.Root>

                  {#if state.canUpdateSps}
                    <Tooltip.Root>
                      <Tooltip.Trigger
                        class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
                        onclick={createEditHandler(controller)}
                        aria-label={$t('common.edit')}
                      >
                        <PencilIcon class="size-4" />
                      </Tooltip.Trigger>
                      <Tooltip.Content>{$t('common.edit')}</Tooltip.Content>
                    </Tooltip.Root>

                    <Tooltip.Root>
                      <Tooltip.Trigger
                        class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
                        onclick={createCopyHandler(controller)}
                        aria-label={$t('common.copy')}
                      >
                        <CopyIcon class="size-4" />
                      </Tooltip.Trigger>
                      <Tooltip.Content>{$t('common.copy')}</Tooltip.Content>
                    </Tooltip.Root>
                  {/if}

                  {#if state.canDeleteSps}
                    <Tooltip.Root>
                      <Tooltip.Trigger
                        class={`${buttonVariants({
                          variant: 'outline',
                          size: 'icon-sm'
                        })} text-destructive hover:bg-destructive/10 hover:text-destructive`}
                        onclick={createDeleteHandler(controller)}
                        aria-label={$t('common.delete')}
                      >
                        <Trash2Icon class="size-4" />
                      </Tooltip.Trigger>
                      <Tooltip.Content>{$t('common.delete')}</Tooltip.Content>
                    </Tooltip.Root>
                  {/if}
                </ButtonGroup.Root>
              </div>

              <div class="mt-4 rounded-md border border-border/70 bg-background/60 p-3">
                <p
                  class="mb-2 flex items-center gap-2 text-xs font-medium tracking-wide text-muted-foreground uppercase"
                >
                  <NetworkIcon class="size-3.5" />
                  {$t('facility.control_cabinet_detail.system_types')}
                </p>
                {#if state.getSystemTypesForController(controller.id).length === 0}
                  <p class="text-sm text-muted-foreground">
                    {$t('facility.control_cabinet_detail.no_system_types')}
                  </p>
                {:else}
                  <div class="grid gap-2 sm:grid-cols-2">
                    {#each state.getSystemTypesForController(controller.id) as systemType (systemType.id)}
                      <div class="rounded-md border border-border/60 bg-muted/20 px-3 py-2 text-sm">
                        <p class="font-medium text-foreground">
                          {systemType.system_type_name ?? systemType.system_type_id}
                        </p>
                        <p class="text-xs text-muted-foreground">{systemTypeSummary(systemType)}</p>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </Tooltip.Provider>
  </Card.Content>
</Card.Root>
