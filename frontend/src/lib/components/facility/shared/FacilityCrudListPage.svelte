<script lang="ts" generics="TItem">
  import { onMount } from 'svelte';
  import type { Component, Snippet } from 'svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import type { ListState } from '$lib/application/useCases/listUseCase.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { BacnetReferenceResource } from '$lib/domain/facility/index.js';
  import type { FacilityDeleteImpactResource } from '$lib/domain/facility/index.js';
  import { bacnetReferenceUsageRepository } from '$lib/infrastructure/api/bacnetReferenceUsageRepository.js';
  import { facilityDeleteImpactRepository } from '$lib/infrastructure/api/facilityDeleteImpactRepository.js';
  import {
    facilityReferenceDataCache,
    type FacilityRealtimeResource
  } from '$lib/services/facilityReferenceDataCache.js';
  import type { CrudPageActions } from './crudPageActions.svelte.js';

  const t = createTranslator();

  interface ListStore<T> {
    subscribe: (run: (value: ListState<T>) => void) => () => void;
    load: (searchText?: string) => void | Promise<void>;
    reload: () => void | Promise<void>;
    goToPage: (page: number) => void | Promise<void>;
    search: (searchText: string) => void;
  }

  interface Props {
    title: string;
    description: string;
    createLabel: string;
    permissionResource: string;
    store: ListStore<TItem>;
    actions: CrudPageActions<TItem>;
    form: Component<any>;
    columns: Array<{ key: string; label: string; width?: string }>;
    rowSnippet: Snippet<[TItem]>;
    searchPlaceholder: string;
    emptyMessage: string;
    documentTitle?: string;
    bacnetUsageResource?: BacnetReferenceResource;
    deleteImpactResource?: FacilityDeleteImpactResource;
    getItemId?: (item: TItem) => string;
    realtimeResource?: FacilityRealtimeResource;
  }

  let {
    title,
    description,
    createLabel,
    permissionResource,
    store,
    actions,
    form: Form,
    columns,
    rowSnippet: itemRows,
    searchPlaceholder,
    emptyMessage,
    documentTitle = title,
    bacnetUsageResource,
    deleteImpactResource,
    getItemId,
    realtimeResource
  }: Props = $props();

  let usageRequestID = 0;
  let deleteImpactRequestID = 0;
  let remoteChangePending = $state(false);

  function resolveItemId(item: TItem): string {
    if (getItemId) return getItemId(item);
    const id = (item as { id?: unknown })?.id;
    return typeof id === 'string' ? id : '';
  }

  async function loadBacnetUsage(
    resource: BacnetReferenceResource,
    ids: string[],
    requestID: number
  ) {
    try {
      const counts = await bacnetReferenceUsageRepository.getCounts(resource, ids);
      if (requestID === usageRequestID) {
        actions.setBacnetUsageCounts(counts);
      }
    } catch (error) {
      console.error('Failed to load BACnet reference usage:', error);
      if (requestID === usageRequestID) {
        actions.setBacnetUsageCounts({});
      }
    }
  }

  async function loadDeleteImpacts(
    resource: FacilityDeleteImpactResource,
    ids: string[],
    requestID: number
  ) {
    try {
      const impacts = await facilityDeleteImpactRepository.getImpacts(resource, ids);
      if (requestID !== deleteImpactRequestID) return;
      actions.setReferenceImpactCounts(
        Object.fromEntries(
          impacts.map((impact) => [
            impact.id,
            impact.blockers.reduce((count, blocker) => count + blocker.count, 0)
          ])
        )
      );
    } catch (error) {
      console.error('Failed to load facility delete impact:', error);
      if (requestID === deleteImpactRequestID) actions.setReferenceImpactCounts({});
    }
  }

  async function confirmBacnetImpactUpdate(item: TItem): Promise<boolean> {
    if (bacnetUsageResource) {
      const id = resolveItemId(item);
      if (id) {
        try {
          const counts = await bacnetReferenceUsageRepository.getCounts(bacnetUsageResource, [id]);
          actions.mergeBacnetUsageCounts(counts);
        } catch (error) {
          console.error('Failed to refresh BACnet reference usage:', error);
        }
      }
    }
    return actions.confirmBacnetImpactUpdate(item);
  }

  onMount(() => {
    store.load();

    return facilityReferenceDataCache.subscribeFacilityChanges((event) => {
      if (!realtimeResource || (event.resource !== realtimeResource && event.resource !== 'all'))
        return;
      if (actions.showForm) {
        if (!facilityReferenceDataCache.isChangeFromCurrentUser(event)) {
          remoteChangePending = true;
        }
        return;
      }
      void store.reload();
    });
  });

  $effect(() => {
    if (!bacnetUsageResource) {
      actions.setBacnetUsageCounts({});
      return;
    }

    const ids = ($store.items ?? []).map(resolveItemId).filter(Boolean);
    const requestID = ++usageRequestID;
    if (ids.length === 0) {
      actions.setBacnetUsageCounts({});
      return;
    }

    void loadBacnetUsage(bacnetUsageResource, ids, requestID);
  });

  $effect(() => {
    if (!deleteImpactResource) {
      actions.setReferenceImpactCounts({});
      return;
    }
    const ids = ($store.items ?? []).map(resolveItemId).filter(Boolean);
    const requestID = ++deleteImpactRequestID;
    if (ids.length === 0) {
      actions.setReferenceImpactCounts({});
      return;
    }
    void loadDeleteImpacts(deleteImpactResource, ids, requestID);
  });
</script>

<svelte:head>
  <title>{documentTitle} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <EntityListHeader
    {title}
    {description}
    backHref="/facility"
    backLabel={$t('common.back')}
    {createLabel}
    canCreate={!actions.showForm && canPerform('create', permissionResource)}
    createActive={actions.showForm}
    onCreateClick={() => actions.create()}
  />

  {#if actions.showForm}
    {#if remoteChangePending}
      <div
        class="flex items-center justify-between gap-3 rounded-md border border-warning-border bg-warning-muted px-3 py-2 text-sm text-warning-muted-foreground"
        role="status"
      >
        <span>{$t('facility.realtime_change_pending')}</span>
        <button
          class="font-medium underline underline-offset-2"
          type="button"
          onclick={() => {
            remoteChangePending = false;
            void store.reload();
          }}
        >
          {$t('common.refresh')}
        </button>
      </div>
    {/if}
    <Form
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
      beforeUpdate={confirmBacnetImpactUpdate}
    />
  {/if}

  <PaginatedList
    state={$store}
    {columns}
    {searchPlaceholder}
    {emptyMessage}
    onSearch={(text) => store.search(text)}
    onPageChange={(page) => store.goToPage(page)}
    onReload={() => store.reload()}
  >
    {#snippet rowSnippet(item: TItem)}
      {@render itemRows(item)}
    {/snippet}
  </PaginatedList>
</div>
