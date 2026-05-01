<script lang="ts" generics="TItem">
  import { onMount } from 'svelte';
  import type { Component, Snippet } from 'svelte';
  import { Plus } from '@lucide/svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import type { ListState } from '$lib/application/useCases/listUseCase.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import type { CrudPageActions } from './crudPageActions.svelte.js';

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
  <div class="flex items-center justify-between gap-4">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{title}</h1>
      <p class="text-sm text-muted-foreground">{description}</p>
    </div>
    {#if !actions.showForm && canPerform('create', permissionResource)}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {createLabel}
      </Button>
    {/if}
  </div>

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
