<script lang="ts">
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import UserManagementForm from '$lib/components/user-management-form.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { UserDirectoryPageState } from './UserDirectoryPageState.svelte.js';

  type Props = {
    state: UserDirectoryPageState;
  };

  let { state }: Props = $props();
  const t = createTranslator();
</script>

<Dialog.Root bind:open={state.createDialogOpen}>
  <Dialog.Content class="sm:max-w-lg">
    <Dialog.Header>
      <Dialog.Title>{$t('common.create_user')}</Dialog.Title>
      <Dialog.Description>{$t('user.invitation_create_description')}</Dialog.Description>
    </Dialog.Header>
    <UserManagementForm
      onSuccess={() => {
        void state.handleUserCreated();
      }}
      onCancel={() => state.closeCreateDialog()}
    />
  </Dialog.Content>
</Dialog.Root>
