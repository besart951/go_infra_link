<script lang="ts">
  import * as Card from '@ui-svelte/components/ui/card/index.js';
  import { Badge } from '@ui-svelte/components/ui/badge/index.js';
  import { Button, buttonVariants } from '@ui-svelte/components/ui/button/index.js';
  import * as Tooltip from '@ui-svelte/components/ui/tooltip/index.js';
  import CpuIcon from '@lucide/svelte/icons/cpu';
  import NetworkIcon from '@lucide/svelte/icons/network';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import CopyIcon from '@lucide/svelte/icons/copy';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
  import { goto } from '$app/navigation';
  import type { SPSController, SPSControllerSystemType } from '$lib/domain/facility/index.js';
  import { createTranslator } from '@i18n/translator.js';

  type Props = {
    controllers: SPSController[];
    systemTypesByController: Record<string, SPSControllerSystemType[]>;
    canRead: boolean;
    canCreate: boolean;
    canUpdate: boolean;
    canDelete: boolean;
    onCreate: () => void;
    onEdit: (controller: SPSController) => void;
    onCopy: (controller: SPSController) => void;
    onDelete: (controller: SPSController) => void;
  };

  let {
    controllers,
    systemTypesByController,
    canRead,
    canCreate,
    canUpdate,
    canDelete,
    onCreate,
    onEdit,
    onCopy,
    onDelete
  }: Props = $props();

  const t = createTranslator();

  function typesFor(controllerId: string): SPSControllerSystemType[] {
    return systemTypesByController[controllerId] ?? [];
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
      {#if canCreate}
        <Button size="sm" onclick={onCreate}>
          <PlusIcon class="mr-2 size-4" />
          {$t('facility.control_cabinet_detail.create_sps')}
        </Button>
      {/if}
    </div>
  </Card.Header>

  <Card.Content>
    <Tooltip.Provider>
      {#if !canRead}
        <div
          class="rounded-md border border-border bg-muted/30 px-3 py-2 text-sm text-muted-foreground"
        >
          {$t('facility.control_cabinet_detail.no_read_permission')}
        </div>
      {:else if controllers.length === 0}
        <div
          class="rounded-md border border-dashed border-border px-4 py-8 text-center text-sm text-muted-foreground"
        >
          {$t('facility.control_cabinet_detail.no_sps')}
        </div>
      {:else}
        <div class="space-y-3">
          {#each controllers as controller (controller.id)}
            <div class="rounded-lg border border-border/70 bg-muted/20 p-4">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p class="text-base font-semibold text-foreground">{controller.device_name}</p>
                  <div class="mt-1 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                    <Badge variant="secondary">GA: {controller.ga_device ?? '-'}</Badge>
                    <Badge variant="outline">IP: {controller.ip_address ?? '-'}</Badge>
                  </div>
                </div>

                <div class="flex items-center gap-1">
                  <Tooltip.Root>
                    <Tooltip.Trigger
                      class={buttonVariants({ variant: 'ghost', size: 'sm' })}
                      onclick={() => goto(`/facility/sps-controllers/${controller.id}`)}
                    >
                      <ArrowRightIcon class="size-4" />
                    </Tooltip.Trigger>
                    <Tooltip.Content>{$t('facility.view')}</Tooltip.Content>
                  </Tooltip.Root>

                  {#if canUpdate}
                    <Tooltip.Root>
                      <Tooltip.Trigger
                        class={buttonVariants({ variant: 'ghost', size: 'sm' })}
                        onclick={() => onEdit(controller)}
                      >
                        <PencilIcon class="size-4" />
                      </Tooltip.Trigger>
                      <Tooltip.Content>{$t('common.edit')}</Tooltip.Content>
                    </Tooltip.Root>

                    <Tooltip.Root>
                      <Tooltip.Trigger
                        class={buttonVariants({ variant: 'ghost', size: 'sm' })}
                        onclick={() => onCopy(controller)}
                      >
                        <CopyIcon class="size-4" />
                      </Tooltip.Trigger>
                      <Tooltip.Content>{$t('common.copy')}</Tooltip.Content>
                    </Tooltip.Root>
                  {/if}

                  {#if canDelete}
                    <Tooltip.Root>
                      <Tooltip.Trigger
                        class={`${buttonVariants({ variant: 'ghost', size: 'sm' })} text-destructive`}
                        onclick={() => onDelete(controller)}
                      >
                        <Trash2Icon class="size-4" />
                      </Tooltip.Trigger>
                      <Tooltip.Content>{$t('common.delete')}</Tooltip.Content>
                    </Tooltip.Root>
                  {/if}
                </div>
              </div>

              <div class="mt-4 rounded-md border border-border/70 bg-background/60 p-3">
                <p
                  class="mb-2 flex items-center gap-2 text-xs font-medium tracking-wide text-muted-foreground uppercase"
                >
                  <NetworkIcon class="size-3.5" />
                  {$t('facility.control_cabinet_detail.system_types')}
                </p>
                {#if typesFor(controller.id).length === 0}
                  <p class="text-sm text-muted-foreground">
                    {$t('facility.control_cabinet_detail.no_system_types')}
                  </p>
                {:else}
                  <div class="grid gap-2 sm:grid-cols-2">
                    {#each typesFor(controller.id) as systemType (systemType.id)}
                      <div class="rounded-md border border-border/60 bg-muted/20 px-3 py-2 text-sm">
                        <p class="font-medium text-foreground">
                          {systemType.system_type_name ?? systemType.system_type_id}
                        </p>
                        <p class="text-xs text-muted-foreground">
                          {systemType.number != null
                            ? `${$t('facility.control_cabinet_detail.number')}: ${systemType.number}`
                            : ''}
                          {#if systemType.document_name}
                            {systemType.number != null ? ' • ' : ''}
                            {systemType.document_name}
                          {/if}
                        </p>
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
