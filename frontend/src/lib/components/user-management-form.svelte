<script lang="ts">
  import { getAllowedRolesForCreation } from '$lib/stores/auth.svelte';
  import { getErrorMessage, getFieldErrors } from '$lib/api/client.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import * as Command from '$lib/components/ui/command/index.js';
  import { Check } from '@lucide/svelte';
  import { cn } from '$lib/utils.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import { userRepository, type AllowedRole } from '$lib/infrastructure/api/userRepository.js';

  interface Props {
    onSuccess?: () => void;
    onCancel?: () => void;
  }

  let { onSuccess, onCancel }: Props = $props();

  const t = createTranslator();

  let email = $state('');
  let selectedRole = $state<AllowedRole | null>(null);
  let openCombobox = $state(false);

  let isSubmitting = $state(false);
  let error = $state('');
  let fieldErrors = $state<Record<string, string>>({});

  const allowedRoles = $derived(getAllowedRolesForCreation());

  async function handleSubmit(e: Event) {
    e.preventDefault();

    if (!selectedRole) {
      error = translate('users.form.role_required');
      return;
    }

    isSubmitting = true;
    error = '';
    fieldErrors = {};

    try {
      await userRepository.invite({
        email,
        role: selectedRole.role
      });

      email = '';
      selectedRole = null;

      if (onSuccess) onSuccess();
    } catch (err) {
      error = getErrorMessage(err);
      fieldErrors = getFieldErrors(err);
    } finally {
      isSubmitting = false;
    }
  }
</script>

<form onsubmit={handleSubmit} class="space-y-4">
  {#if error}
    <div
      class="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive"
    >
      {error}
    </div>
  {/if}

  <div class="space-y-2">
    <Label for="email">{$t('auth.email')}</Label>
    <Input
      type="email"
      id="email"
      bind:value={email}
      required
      class={fieldErrors.email ? 'border-destructive' : ''}
    />
    {#if fieldErrors.email}
      <p class="text-sm text-destructive">{fieldErrors.email}</p>
    {/if}
  </div>

  <div class="space-y-2">
    <Label for="role">{$t('common.role')}</Label>
    <Popover.Root bind:open={openCombobox}>
      <Popover.Trigger>
        {#snippet child({ props })}
          <Button
            {...props}
            variant="outline"
            role="combobox"
            aria-expanded={openCombobox}
            class={cn(
              'w-full justify-between',
              !selectedRole && 'text-muted-foreground',
              fieldErrors.role && 'border-destructive'
            )}
          >
            {selectedRole ? selectedRole.display_name : $t('users.form.select_role_placeholder')}
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="ml-2 h-4 w-4 shrink-0 opacity-50"
            >
              <path d="m7 15 5 5 5-5" />
              <path d="m7 9 5-5 5 5" />
            </svg>
          </Button>
        {/snippet}
      </Popover.Trigger>
      <Popover.Content class="w-full bg-background p-0 text-foreground" side="bottom" align="start">
        <Command.Root>
          <Command.Input placeholder={$t('users.form.search_roles_placeholder')} />
          <Command.Empty>{$t('users.form.no_role_found')}</Command.Empty>
          <Command.List>
            {#each allowedRoles as roleObj (roleObj.role)}
              <Command.Item
                value={roleObj.role}
                onSelect={() => {
                  selectedRole = roleObj;
                  openCombobox = false;
                }}
              >
                <Check
                  class={cn(
                    'mr-2 h-4 w-4',
                    selectedRole?.role !== roleObj.role && 'text-transparent'
                  )}
                />
                {roleObj.display_name}
              </Command.Item>
            {/each}
          </Command.List>
        </Command.Root>
      </Popover.Content>
    </Popover.Root>
    {#if fieldErrors.role}
      <p class="text-sm text-destructive">{fieldErrors.role}</p>
    {/if}
    <p class="text-sm text-muted-foreground">
      {$t('users.form.allowed_roles_help')}
    </p>
  </div>

  <div class="flex justify-end gap-2 pt-2">
    {#if onCancel}
      <Button type="button" variant="outline" onclick={onCancel} disabled={isSubmitting}>
        {$t('common.cancel')}
      </Button>
    {/if}
    <Button type="submit" disabled={isSubmitting}>
      {isSubmitting ? $t('users.form.creating_user') : $t('user.send_invitation')}
    </Button>
  </div>
</form>
