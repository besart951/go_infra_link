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
  import { bacnetReferenceUsageRepository } from '$lib/infrastructure/api/bacnetReferenceUsageRepository.js';
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
    getItemId?: (item: TItem) => string;
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
    getItemId
  }: Props = $props();

  let usageRequestID = 0;

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
