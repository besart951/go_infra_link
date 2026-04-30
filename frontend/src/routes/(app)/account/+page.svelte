<script lang="ts">
  import { onMount } from 'svelte';
  import { AccountPageState } from '$lib/components/account/AccountPageState.svelte.js';
  import AccountAppearanceSection from '$lib/components/account/AccountAppearanceSection.svelte';
  import AccountInformationSection from '$lib/components/account/AccountInformationSection.svelte';
  import AccountNotificationsSection from '$lib/components/account/AccountNotificationsSection.svelte';
  import AccountPasswordSection from '$lib/components/account/AccountPasswordSection.svelte';
  import AccountTabsNav from '$lib/components/account/AccountTabsNav.svelte';
  import type { AccountTabItem } from '$lib/components/account/types.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
  const state = new AccountPageState();

  const accountTabs: AccountTabItem[] = [
    { value: 'information', label: $t('pages.account_tabs_information') },
    { value: 'notifications', label: $t('pages.account_tabs_notifications') },
    { value: 'password', label: $t('pages.account_tabs_password') },
    { value: 'preferences', label: $t('pages.account_tabs_preferences') }
  ];

  onMount(() => {
    void state.loadAccount();
  });
</script>

<svelte:head>
  <title>{$t('navigation.account')} | {$t('app.brand')}</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <div>
    <h1 class="text-3xl font-bold tracking-tight">{$t('navigation.account')}</h1>
    <p class="mt-1 text-muted-foreground">{$t('pages.account_desc')}</p>
  </div>

  <div class="grid gap-6 md:grid-cols-[220px_minmax(0,1fr)]">
    <AccountTabsNav bind:activeTab={state.activeTab} tabs={accountTabs} />

    <section>
      {#if state.isLoading}
        <div class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
          {$t('common.loading')}
        </div>
      {:else if !state.currentUser}
        <div class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
          {$t('errors.unknown_error')}
        </div>
      {:else if state.activeTab === 'information'}
        <AccountInformationSection
          currentUser={state.currentUser}
          bind:firstName={state.firstName}
          bind:lastName={state.lastName}
          bind:email={state.email}
          isSavingProfile={state.isSavingProfile}
          permissions={state.permissions}
          teamsError={state.teamsError}
          userTeams={state.userTeams}
          onSubmit={(event) => state.handleInformationSubmit(event)}
        />
      {:else if state.activeTab === 'notifications'}
        <AccountNotificationsSection
          currentUser={state.currentUser}
          bind:notificationDraft={state.notificationDraft}
          bind:verificationCode={state.verificationCode}
          notificationsError={state.notificationsError}
          notificationEmailVerified={state.notificationEmailVerified}
          notificationIsDirty={state.notificationIsDirty}
          canSendEmailVerification={state.canSendEmailVerification}
          isLoadingNotifications={state.isLoadingNotifications}
          isSavingNotifications={state.isSavingNotifications}
          isSendingEmailVerification={state.isSendingEmailVerification}
          isVerifyingNotificationEmail={state.isVerifyingNotificationEmail}
          onSubmit={(event) => state.handleNotificationSubmit(event)}
          onReset={() => state.resetNotificationDraft()}
          onSendEmailVerification={() => state.handleSendEmailVerification()}
          onVerifyNotificationEmail={() => state.handleVerifyNotificationEmail()}
        />
      {:else if state.activeTab === 'password'}
        <AccountPasswordSection
          bind:newPassword={state.newPassword}
          bind:confirmPassword={state.confirmPassword}
          isSavingPassword={state.isSavingPassword}
          onSubmit={(event) => state.handlePasswordSubmit(event)}
        />
      {:else if state.activeTab === 'preferences'}
        <AccountAppearanceSection />
      {/if}
    </section>
  </div>
</div>
