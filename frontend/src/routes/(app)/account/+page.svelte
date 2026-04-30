<script lang="ts">
  import { onMount } from 'svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import AccountAppearanceSection from '$lib/components/account/AccountAppearanceSection.svelte';
  import AccountInformationSection from '$lib/components/account/AccountInformationSection.svelte';
  import AccountNotificationsSection from '$lib/components/account/AccountNotificationsSection.svelte';
  import AccountPasswordSection from '$lib/components/account/AccountPasswordSection.svelte';
  import AccountTabsNav from '$lib/components/account/AccountTabsNav.svelte';
  import type { AccountTab, AccountTabItem } from '$lib/components/account/types.js';
  import {
    getCurrentUser,
    updateCurrentUser,
    updateCurrentUserPassword,
    type UpdateUserRequest,
    type User
  } from '$lib/api/users.js';
  import { listTeams, listTeamMembers } from '$lib/api/teams.js';
  import { getErrorMessage } from '$lib/api/client.js';
  import { GetNotificationPreferenceUseCase } from '$lib/application/useCases/notification/getNotificationPreferenceUseCase.js';
  import { SaveNotificationPreferenceUseCase } from '$lib/application/useCases/notification/saveNotificationPreferenceUseCase.js';
  import { SendNotificationEmailVerificationUseCase } from '$lib/application/useCases/notification/sendNotificationEmailVerificationUseCase.js';
  import { VerifyNotificationEmailUseCase } from '$lib/application/useCases/notification/verifyNotificationEmailUseCase.js';
  import { createTranslator } from '$lib/i18n/translator';
  import { notificationPreferenceRepository } from '$lib/infrastructure/api/notificationPreferenceRepository.js';
  import type {
    NotificationPreferenceFormValues,
    UpsertUserNotificationPreferenceRequest,
    UserNotificationPreference
  } from '$lib/domain/notification/index.js';
  import {
    createNotificationPreferenceFormValues,
    hasVerifiedNotificationEmail,
    normalizeNotificationPreferenceInput
  } from '$lib/domain/notification/index.js';

  const t = createTranslator();
  const getNotificationPreference = new GetNotificationPreferenceUseCase(
    notificationPreferenceRepository
  );
  const saveNotificationPreference = new SaveNotificationPreferenceUseCase(
    notificationPreferenceRepository
  );
  const sendNotificationEmailVerification = new SendNotificationEmailVerificationUseCase(
    notificationPreferenceRepository
  );
  const verifyNotificationEmail = new VerifyNotificationEmailUseCase(
    notificationPreferenceRepository
  );

  let activeTab = $state<AccountTab>('information');
  let currentUser = $state<User | null>(null);

  let firstName = $state('');
  let lastName = $state('');
  let email = $state('');

  let newPassword = $state('');
  let confirmPassword = $state('');

  let isSavingProfile = $state(false);
  let isSavingPassword = $state(false);
  let isSavingNotifications = $state(false);
  let isSendingEmailVerification = $state(false);
  let isVerifyingNotificationEmail = $state(false);
  let isLoadingNotifications = $state(false);
  let isLoading = $state(true);

  let userTeams = $state<string[]>([]);
  let teamsError = $state<string | null>(null);
  let notificationPreference = $state<UserNotificationPreference | null>(null);
  let notificationDraft = $state<NotificationPreferenceFormValues>(
    createNotificationPreferenceFormValues(null)
  );
  let verificationCode = $state('');
  let notificationsError = $state<string | null>(null);

  const permissions = $derived(currentUser?.permissions ?? []);
  const notificationEmailVerified = $derived(hasVerifiedNotificationEmail(notificationPreference));
  const notificationIsDirty = $derived.by(() => {
    const initial = normalizeNotificationPreferenceInput(
      createNotificationPreferenceFormValues(notificationPreference)
    );
    const current = normalizeNotificationPreferenceInput(notificationDraft);
    return JSON.stringify(initial) !== JSON.stringify(current);
  });
  const canSendEmailVerification = $derived(
    Boolean(notificationPreference?.notification_email) &&
      !notificationIsDirty &&
      !isSendingEmailVerification
  );

  const accountTabs: AccountTabItem[] = [
    { value: 'information', label: $t('pages.account_tabs_information') },
    { value: 'notifications', label: $t('pages.account_tabs_notifications') },
    { value: 'password', label: $t('pages.account_tabs_password') },
    { value: 'preferences', label: $t('pages.account_tabs_preferences') }
  ];

  function applyUserToForm(user: User) {
    firstName = user.first_name;
    lastName = user.last_name;
    email = user.email;
  }

  async function loadUserTeams(userId: string) {
    teamsError = null;
    userTeams = [];

    try {
      const teamsResponse = await listTeams({ page: 1, limit: 100, search: '' });
      const memberLists = await Promise.all(
        teamsResponse.items.map(async (team) => ({
          team,
          members: await listTeamMembers(team.id, { page: 1, limit: 1000 })
        }))
      );

      userTeams = memberLists
        .filter((entry) => entry.members.items.some((member) => member.user_id === userId))
        .map((entry) => entry.team.name);
    } catch (err) {
      teamsError = getErrorMessage(err);
    }
  }

  async function loadNotificationPreference() {
    isLoadingNotifications = true;
    notificationsError = null;

    try {
      notificationPreference = await getNotificationPreference.execute();
      notificationDraft = createNotificationPreferenceFormValues(notificationPreference);
    } catch (err) {
      notificationsError = getErrorMessage(err);
    } finally {
      isLoadingNotifications = false;
    }
  }

  async function loadAccount() {
    isLoading = true;
    try {
      const user = await getCurrentUser();
      currentUser = user;
      applyUserToForm(user);
      await Promise.all([loadUserTeams(user.id), loadNotificationPreference()]);
    } catch (err) {
      addToast(getErrorMessage(err), 'error');
    } finally {
      isLoading = false;
    }
  }

  async function handleInformationSubmit(event: SubmitEvent) {
    event.preventDefault();
    if (!currentUser) return;

    isSavingProfile = true;
    try {
      const payload: UpdateUserRequest = {
        first_name: firstName,
        last_name: lastName,
        email
      };
      const updated = await updateCurrentUser(currentUser.id, payload);
      currentUser = updated;
      applyUserToForm(updated);
      addToast($t('messages.account_info_saved'), 'success');
    } catch (err) {
      addToast(getErrorMessage(err), 'error');
    } finally {
      isSavingProfile = false;
    }
  }

  async function handlePasswordSubmit(event: SubmitEvent) {
    event.preventDefault();
    if (!currentUser) return;

    if (newPassword.length < 8) {
      addToast(
        $t('validation.password_too_short', { field: $t('auth.password'), min: 8 }),
        'error'
      );
      return;
    }

    if (newPassword !== confirmPassword) {
      addToast(
        $t('validation.must_match', {
          field1: $t('auth.new_password'),
          field2: $t('auth.confirm_password')
        }),
        'error'
      );
      return;
    }

    isSavingPassword = true;
    try {
      await updateCurrentUserPassword(currentUser.id, newPassword);
      newPassword = '';
      confirmPassword = '';
      addToast($t('messages.account_password_saved'), 'success');
    } catch (err) {
      addToast(getErrorMessage(err), 'error');
    } finally {
      isSavingPassword = false;
    }
  }

  async function handleNotificationSubmit(event: SubmitEvent) {
    event.preventDefault();

    isSavingNotifications = true;
    notificationsError = null;
    try {
      const payload: UpsertUserNotificationPreferenceRequest =
        normalizeNotificationPreferenceInput(notificationDraft);
      notificationPreference = await saveNotificationPreference.execute(payload);
      notificationDraft = createNotificationPreferenceFormValues(notificationPreference);
      verificationCode = '';
      addToast($t('messages.account_notifications_saved'), 'success');
    } catch (err) {
      notificationsError = getErrorMessage(err);
      addToast(notificationsError, 'error');
    } finally {
      isSavingNotifications = false;
    }
  }

  function resetNotificationDraft() {
    notificationDraft = createNotificationPreferenceFormValues(notificationPreference);
    notificationsError = null;
  }

  async function handleSendEmailVerification() {
    isSendingEmailVerification = true;
    notificationsError = null;
    try {
      notificationPreference = await sendNotificationEmailVerification.execute();
      notificationDraft = createNotificationPreferenceFormValues(notificationPreference);
      verificationCode = '';
      addToast($t('messages.account_notification_email_code_sent'), 'success');
    } catch (err) {
      notificationsError = getErrorMessage(err);
      addToast(notificationsError, 'error');
    } finally {
      isSendingEmailVerification = false;
    }
  }

  async function handleVerifyNotificationEmail() {
    if (verificationCode.trim().length !== 6) {
      addToast($t('notifications.preferences.email.code_invalid'), 'error');
      return;
    }

    isVerifyingNotificationEmail = true;
    notificationsError = null;
    try {
      notificationPreference = await verifyNotificationEmail.execute({
        code: verificationCode.trim()
      });
      notificationDraft = createNotificationPreferenceFormValues(notificationPreference);
      verificationCode = '';
      addToast($t('messages.account_notification_email_verified'), 'success');
    } catch (err) {
      notificationsError = getErrorMessage(err);
      addToast(notificationsError, 'error');
    } finally {
      isVerifyingNotificationEmail = false;
    }
  }

  onMount(() => {
    void loadAccount();
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
    <AccountTabsNav bind:activeTab tabs={accountTabs} />

    <section>
      {#if isLoading}
        <div class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
          {$t('common.loading')}
        </div>
      {:else if !currentUser}
        <div class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
          {$t('errors.unknown_error')}
        </div>
      {:else if activeTab === 'information'}
        <AccountInformationSection
          {currentUser}
          bind:firstName
          bind:lastName
          bind:email
          {isSavingProfile}
          {permissions}
          {teamsError}
          {userTeams}
          onSubmit={handleInformationSubmit}
        />
      {:else if activeTab === 'notifications'}
        <AccountNotificationsSection
          {currentUser}
          bind:notificationDraft
          bind:verificationCode
          {notificationsError}
          {notificationEmailVerified}
          {notificationIsDirty}
          {canSendEmailVerification}
          {isLoadingNotifications}
          {isSavingNotifications}
          {isSendingEmailVerification}
          {isVerifyingNotificationEmail}
          onSubmit={handleNotificationSubmit}
          onReset={resetNotificationDraft}
          onSendEmailVerification={handleSendEmailVerification}
          onVerifyNotificationEmail={handleVerifyNotificationEmail}
        />
      {:else if activeTab === 'password'}
        <AccountPasswordSection
          bind:newPassword
          bind:confirmPassword
          {isSavingPassword}
          onSubmit={handlePasswordSubmit}
        />
      {:else if activeTab === 'preferences'}
        <AccountAppearanceSection />
      {/if}
    </section>
  </div>
</div>
