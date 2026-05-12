<script lang="ts">
  import { onMount } from 'svelte';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import UserCreateDialog from '$lib/components/users/UserCreateDialog.svelte';
  import UserDirectoryPagination from '$lib/components/users/UserDirectoryPagination.svelte';
  import { UserDirectoryPageState } from '$lib/components/users/UserDirectoryPageState.svelte.js';
  import UserDirectoryTable from '$lib/components/users/UserDirectoryTable.svelte';
  import UserDirectoryToolbar from '$lib/components/users/UserDirectoryToolbar.svelte';
  import { createTranslator } from '$lib/i18n/translator';

  interface Props {
    data?: {
      user?: {
        can_access_user_directory?: boolean | null;
      } | null;
    };
  }

  let { data }: Props = $props();

  const t = createTranslator();
  const state = new UserDirectoryPageState();

  onMount(() => {
    void state.initialize(data?.user ? Boolean(data.user.can_access_user_directory) : undefined);
    return state.startResendClock();
  });
</script>

<svelte:head>
  <title>{$t('navigation.users')} | Infra Link</title>
</svelte:head>

<Tooltip.Provider>
  <div class="flex flex-col gap-6">
    <EntityListHeader
      title={$t('pages.user_management')}
      description={$t('pages.user_management_desc')}
      backHref="/users"
      backLabel={$t('common.back')}
      createLabel={$t('common.create_user')}
      canCreate={state.pageCapabilities.can_create_user}
      createActive={state.createDialogOpen}
      onCreateClick={() => state.openCreateDialog()}
    />

    <UserDirectoryToolbar {state} />
    <UserDirectoryTable {state} />
    <UserDirectoryPagination {state} />
  </div>

  <UserCreateDialog {state} />
  <ConfirmDialog />
</Tooltip.Provider>
