<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { UserDirectoryUser } from '$lib/infrastructure/api/userRepository.js';
  import MoreVertical from '@lucide/svelte/icons/more-vertical';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Send from '@lucide/svelte/icons/send';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import UserCheck from '@lucide/svelte/icons/user-check';
  import UserMinus from '@lucide/svelte/icons/user-minus';
  import type { UserDirectoryPageState } from './UserDirectoryPageState.svelte.js';

  type Props = {
    state: UserDirectoryPageState;
    user: UserDirectoryUser;
  };

  let { state, user }: Props = $props();
  const t = createTranslator();

  const hasRoleAction = $derived(user.capabilities.can_change_role);
  const hasToggleAction = $derived(user.capabilities.can_disable || user.capabilities.can_enable);
  const hasInvitationAction = $derived(state.hasInvitationResendAction(user));
  const hasDeleteAction = $derived(user.capabilities.can_delete);
  const hasRestoreAction = $derived(user.capabilities.can_restore);
  const hasActions = $derived(
    hasRoleAction || hasToggleAction || hasInvitationAction || hasDeleteAction || hasRestoreAction
  );
  const canResendInvitation = $derived(state.canResendInvitation(user, state.resendClock));
  const resendDisabledReason = $derived(
    state.invitationResendDisabledReason(user, state.resendClock)
  );
  const userName = $derived(`${user.first_name} ${user.last_name}`);
</script>

{#if hasActions}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button variant="ghost" size="sm" aria-label={$t('common.actions')} {...props}>
          <MoreVertical class="h-4 w-4" />
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end" class="w-56">
      {#if hasRoleAction}
        <DropdownMenu.Label>{$t('common.change_role')}</DropdownMenu.Label>
        <DropdownMenu.Separator />
        {#each state.roleOptionsFor(user) as roleObj (roleObj.role)}
          <DropdownMenu.Item onclick={() => void state.handleRoleChange(user.id, roleObj.role)}>
            {roleObj.display_name}
          </DropdownMenu.Item>
        {/each}
      {/if}

      {#if hasToggleAction}
        {#if hasRoleAction}
          <DropdownMenu.Separator />
        {/if}
        <DropdownMenu.Item onclick={() => void state.handleToggleActive(user.id, user.is_active)}>
          {#if user.is_active}
            <UserMinus class="mr-2 h-4 w-4" />
            {$t('actions.disable_user')}
          {:else}
            <UserCheck class="mr-2 h-4 w-4" />
            {$t('actions.enable_user')}
          {/if}
        </DropdownMenu.Item>
      {/if}

      {#if hasInvitationAction}
        {#if hasRoleAction || hasToggleAction}
          <DropdownMenu.Separator />
        {/if}
        {#if canResendInvitation}
          <DropdownMenu.Item onclick={() => void state.handleResendInvitation(user.id)}>
            <Send class="mr-2 h-4 w-4" />
            {$t('user.resend_invitation')}
          </DropdownMenu.Item>
        {:else}
          <Tooltip.Root>
            <Tooltip.Trigger class="block w-full">
              <DropdownMenu.Item
                disabled
                class="w-full cursor-not-allowed"
                title={resendDisabledReason ?? undefined}
                aria-label={`${$t('user.resend_invitation')}: ${resendDisabledReason ?? ''}`}
              >
                <Send class="mr-2 h-4 w-4" />
                {$t('user.resend_invitation')}
              </DropdownMenu.Item>
            </Tooltip.Trigger>
            {#if resendDisabledReason}
              <Tooltip.Content side="left" class="max-w-xs text-sm">
                {resendDisabledReason}
              </Tooltip.Content>
            {/if}
          </Tooltip.Root>
        {/if}
      {/if}

      {#if hasRestoreAction}
        {#if hasRoleAction || hasToggleAction || hasInvitationAction}
          <DropdownMenu.Separator />
        {/if}
        <DropdownMenu.Item onclick={() => void state.handleRestoreUser(user.id, userName)}>
          <RotateCcw class="mr-2 h-4 w-4" />
          {$t('actions.restore_user')}
        </DropdownMenu.Item>
      {/if}

      {#if hasDeleteAction && !user.is_deleted}
        {#if hasRoleAction || hasToggleAction || hasInvitationAction}
          <DropdownMenu.Separator />
        {/if}
        <DropdownMenu.Item
          class="text-destructive"
          onclick={() => void state.handleDeleteUser(user.id, userName)}
        >
          <Trash2 class="mr-2 h-4 w-4" />
          {$t('actions.delete_user')}
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
