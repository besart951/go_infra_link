<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { createTranslator } from '$lib/i18n/translator';
  import type { User } from '$lib/infrastructure/api/userRepository.js';

  interface Props {
    currentUser: User;
    firstName: string;
    lastName: string;
    email: string;
    isSavingProfile: boolean;
    permissions: string[];
    teamsError: string | null;
    userTeams: string[];
    onSubmit: (event: SubmitEvent) => void | Promise<void>;
  }

  let {
    currentUser,
    firstName = $bindable(),
    lastName = $bindable(),
    email = $bindable(),
    isSavingProfile,
    permissions,
    teamsError,
    userTeams,
    onSubmit
  }: Props = $props();

  const t = createTranslator();
</script>

<div class="flex flex-col gap-4">
  <form class="rounded-lg border bg-card p-4" onsubmit={onSubmit}>
    <div class="mb-4">
      <h2 class="text-base font-semibold">{$t('pages.account_information_title')}</h2>
      <p class="text-sm text-muted-foreground">{$t('pages.account_information_desc')}</p>
    </div>

    <div class="grid gap-4 sm:grid-cols-2">
      <div class="flex flex-col gap-2">
        <label for="first_name" class="text-sm font-medium">{$t('user.firstname')}</label>
        <Input id="first_name" bind:value={firstName} required minlength={1} maxlength={100} />
      </div>
      <div class="flex flex-col gap-2">
        <label for="last_name" class="text-sm font-medium">{$t('user.lastname')}</label>
        <Input id="last_name" bind:value={lastName} required minlength={1} maxlength={100} />
      </div>
    </div>

    <div class="mt-4 flex flex-col gap-2">
      <label for="email" class="text-sm font-medium">{$t('user.email')}</label>
      <Input id="email" type="email" bind:value={email} required />
    </div>

    <div class="mt-4 flex justify-end">
      <Button type="submit" disabled={isSavingProfile}>
        {isSavingProfile ? $t('common.saving') : $t('common.save_changes')}
      </Button>
    </div>
  </form>

  <div class="rounded-lg border bg-card p-4">
    <h3 class="text-base font-semibold">{$t('pages.account_access_title')}</h3>
    <div class="mt-3 grid gap-4 sm:grid-cols-2">
      <div>
        <p class="text-sm text-muted-foreground">{$t('common.role')}</p>
        <p class="mt-1 font-medium">{currentUser.role}</p>
      </div>
      <div>
        <p class="text-sm text-muted-foreground">
          {$t('roles.permissions.table.permission')}
        </p>
        <div class="mt-2 flex flex-wrap gap-2">
          {#if permissions.length === 0}
            <span class="text-sm text-muted-foreground"
              >{$t('pages.account_permissions_empty')}</span
            >
          {:else}
            {#each permissions as permission (permission)}
              <Badge variant="outline">{permission}</Badge>
            {/each}
          {/if}
        </div>
      </div>
    </div>

    <div class="mt-4">
      <p class="text-sm text-muted-foreground">{$t('navigation.teams')}</p>
      <div class="mt-2 flex flex-wrap gap-2">
        {#if teamsError}
          <span class="text-sm text-muted-foreground">{teamsError}</span>
        {:else if userTeams.length === 0}
          <span class="text-sm text-muted-foreground">{$t('pages.account_teams_empty')}</span>
        {:else}
          {#each userTeams as teamName (teamName)}
            <Badge variant="secondary">{teamName}</Badge>
          {/each}
        {/if}
      </div>
    </div>
  </div>
</div>
