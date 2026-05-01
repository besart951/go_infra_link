<script lang="ts" generics="TItem">
  import { onMount } from 'svelte';
  import type { Component, Snippet } from 'svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import type { ListState } from '$lib/application/useCases/listUseCase.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
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
    documentTitle = title
  }: Props = $props();

  onMount(() => {
    store.load();
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
    backLabel={$t('hub.back_to_overview')}
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
