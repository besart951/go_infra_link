<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import {
    getCurrentUser,
    updateCurrentUser,
    updateCurrentUserPassword,
    type User,
    type UpdateUserRequest
  } from '$lib/api/users.js';
  import { listTeams, listTeamMembers } from '$lib/api/teams.js';
  import { getErrorMessage } from '$lib/api/client.js';
  import { GetNotificationPreferenceUseCase } from '$lib/application/useCases/notification/getNotificationPreferenceUseCase.js';
  import { SaveNotificationPreferenceUseCase } from '$lib/application/useCases/notification/saveNotificationPreferenceUseCase.js';
  import { SendNotificationEmailVerificationUseCase } from '$lib/application/useCases/notification/sendNotificationEmailVerificationUseCase.js';
  import { VerifyNotificationEmailUseCase } from '$lib/application/useCases/notification/verifyNotificationEmailUseCase.js';
  import { createTranslator } from '$lib/i18n/translator';
  import { notificationPreferenceRepository } from '$lib/infrastructure/api/notificationPreferenceRepository.js';
  import {
    CONTRAST_PREFERENCE_STEP,
    DEFAULT_CONTRAST_PREFERENCE,
    FONT_STACKS,
    MAX_CONTRAST_PREFERENCE,
    MIN_CONTRAST_PREFERENCE,
    contrastPreference,
    fontPreference,
    setContrastPreference,
    setFontPreference,
    setThemePreference,
    themePreference,
    type FontPreference,
    type ThemePreference
  } from '$lib/stores/appearance.js';
  import {
    Bell,
    CircleCheck,
    Clock3,
    Contrast,
    KeyRound,
    LaptopMinimal,
    Mail,
    MailCheck,
    MonitorCheck,
    Moon,
    RotateCcw,
    Sun,
    Type
  } from '@lucide/svelte';
  import type {
    NotificationChannel,
    NotificationFrequency,
    NotificationPreferenceFormValues,
    UpsertUserNotificationPreferenceRequest,
    UserNotificationPreference
  } from '$lib/domain/notification/index.js';
  import {
    createNotificationPreferenceFormValues,
    hasVerifiedNotificationEmail,
    normalizeNotificationPreferenceInput
  } from '$lib/domain/notification/index.js';

  type AccountTab = 'information' | 'notifications' | 'password' | 'preferences';
  type ThemeOption = {
    value: ThemePreference;
    label: string;
    description: string;
    icon: typeof LaptopMinimal;
  };
  type FontOption = {
    value: FontPreference;
    label: string;
    description: string;
    stack: string;
  };
  type NotificationOption<TValue extends string> = {
    value: TValue;
    label: string;
    description: string;
    icon: typeof LaptopMinimal;
  };

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

  const options: ThemeOption[] = [
    {
      value: 'system',
      label: $t('pages.settings_theme_system'),
      description: $t('pages.settings_theme_system_desc'),
      icon: LaptopMinimal
    },
    {
      value: 'light',
      label: $t('pages.settings_theme_light'),
      description: $t('pages.settings_theme_light_desc'),
      icon: Sun
    },
    {
      value: 'dark',
      label: $t('pages.settings_theme_dark'),
      description: $t('pages.settings_theme_dark_desc'),
      icon: Moon
    }
  ];

  const fontOptions: FontOption[] = [
    {
      value: 'noto',
      label: $t('pages.settings_font_noto'),
      description: $t('pages.settings_font_noto_desc'),
      stack: FONT_STACKS.noto
    },
    {
      value: 'system',
      label: $t('pages.settings_font_system'),
      description: $t('pages.settings_font_system_desc'),
      stack: FONT_STACKS.system
    },
    {
      value: 'serif',
      label: $t('pages.settings_font_serif'),
      description: $t('pages.settings_font_serif_desc'),
      stack: FONT_STACKS.serif
    },
    {
      value: 'mono',
      label: $t('pages.settings_font_mono'),
      description: $t('pages.settings_font_mono_desc'),
      stack: FONT_STACKS.mono
    }
  ];

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

  const channelOptions = $derived<NotificationOption<NotificationChannel>[]>([
    {
      value: 'email',
      label: $t('notifications.preferences.channels.email.label'),
      description: $t('notifications.preferences.channels.email.description'),
      icon: Mail
    },
    {
      value: 'system',
      label: $t('notifications.preferences.channels.system.label'),
      description: $t('notifications.preferences.channels.system.description'),
      icon: MonitorCheck
    },
    {
      value: 'both',
      label: $t('notifications.preferences.channels.both.label'),
      description: $t('notifications.preferences.channels.both.description'),
      icon: MailCheck
    }
  ]);

  const frequencyOptions = $derived<NotificationOption<NotificationFrequency>[]>([
    {
      value: 'immediate',
      label: $t('notifications.preferences.frequencies.immediate.label'),
      description: $t('notifications.preferences.frequencies.immediate.description'),
      icon: Bell
    },
    {
      value: 'hourly',
      label: $t('notifications.preferences.frequencies.hourly.label'),
      description: $t('notifications.preferences.frequencies.hourly.description'),
      icon: Clock3
    },
    {
      value: 'daily',
      label: $t('notifications.preferences.frequencies.daily.label'),
      description: $t('notifications.preferences.frequencies.daily.description'),
      icon: Clock3
    },
    {
      value: 'weekly',
      label: $t('notifications.preferences.frequencies.weekly.label'),
      description: $t('notifications.preferences.frequencies.weekly.description'),
      icon: Clock3
    }
  ]);

  const accountTabs: { value: AccountTab; label: string }[] = [
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
        .filter((entry) => entry.members.items.some((m) => m.user_id === userId))
        .map((entry) => entry.team.name);
    } catch (err) {
      teamsError = getErrorMessage(err);
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

  function handleContrastInput(event: Event) {
    setContrastPreference(Number((event.currentTarget as HTMLInputElement).value));
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
    loadAccount();
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
    <aside>
      <div class="rounded-lg border bg-card p-2">
        <nav class="flex flex-col gap-1">
          {#each accountTabs as tab (tab.value)}
            <button
              type="button"
              class={activeTab === tab.value
                ? 'rounded-md bg-muted px-3 py-2 text-left text-sm font-medium'
                : 'rounded-md px-3 py-2 text-left text-sm text-muted-foreground hover:bg-muted/60'}
              onclick={() => (activeTab = tab.value)}
            >
              {tab.label}
            </button>
          {/each}
        </nav>
      </div>
    </aside>

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
        <div class="flex flex-col gap-4">
          <form class="rounded-lg border bg-card p-4" onsubmit={handleInformationSubmit}>
            <div class="mb-4">
              <h2 class="text-base font-semibold">{$t('pages.account_information_title')}</h2>
              <p class="text-sm text-muted-foreground">{$t('pages.account_information_desc')}</p>
            </div>

            <div class="grid gap-4 sm:grid-cols-2">
              <div class="flex flex-col gap-2">
                <label for="first_name" class="text-sm font-medium">{$t('user.firstname')}</label>
                <Input
                  id="first_name"
                  bind:value={firstName}
                  required
                  minlength={1}
                  maxlength={100}
                />
              </div>
              <div class="flex flex-col gap-2">
                <label for="last_name" class="text-sm font-medium">{$t('user.lastname')}</label>
                <Input
                  id="last_name"
                  bind:value={lastName}
                  required
                  minlength={1}
                  maxlength={100}
                />
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
                  <span class="text-sm text-muted-foreground"
                    >{$t('pages.account_teams_empty')}</span
                  >
                {:else}
                  {#each userTeams as teamName (teamName)}
                    <Badge variant="secondary">{teamName}</Badge>
                  {/each}
                {/if}
              </div>
            </div>
          </div>
        </div>
      {:else if activeTab === 'notifications'}
        <form class="rounded-lg border bg-card p-4" onsubmit={handleNotificationSubmit}>
          <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div class="min-w-0 space-y-1">
              <h2 class="text-base font-semibold">{$t('pages.account_notifications_title')}</h2>
              <p class="text-sm text-muted-foreground">{$t('pages.account_notifications_desc')}</p>
            </div>
            <Badge variant={notificationIsDirty ? 'outline' : 'secondary'}>
              {$t(
                isLoadingNotifications
                  ? 'notifications.status.loading'
                  : notificationIsDirty
                    ? 'notifications.status.unsaved'
                    : 'notifications.status.synced'
              )}
            </Badge>
          </div>

          {#if notificationsError}
            <div
              class="mt-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
            >
              {notificationsError}
            </div>
          {/if}

          {#if isLoadingNotifications}
            <div class="mt-4 rounded-md border bg-muted/30 p-4 text-sm text-muted-foreground">
              {$t('common.loading')}
            </div>
          {:else}
            <div class="mt-5 grid gap-5">
              <section class="grid gap-3 rounded-md border bg-muted/20 p-4">
                <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div class="flex min-w-0 items-center gap-2">
                    <Mail class="size-4 shrink-0 text-muted-foreground" />
                    <h3 class="text-sm font-semibold">
                      {$t('notifications.preferences.email.title')}
                    </h3>
                  </div>
                  <Badge variant={notificationEmailVerified ? 'secondary' : 'outline'}>
                    {$t(
                      notificationEmailVerified
                        ? 'notifications.preferences.email.verified'
                        : notificationDraft.notification_email
                          ? 'notifications.preferences.email.unverified'
                          : 'notifications.preferences.email.missing'
                    )}
                  </Badge>
                </div>

                <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
                  <div class="space-y-2">
                    <label for="notification_email" class="text-sm font-medium">
                      {$t('notifications.preferences.email.label')}
                    </label>
                    <Input
                      id="notification_email"
                      type="email"
                      bind:value={notificationDraft.notification_email}
                      placeholder={currentUser.email}
                      disabled={isSavingNotifications}
                    />
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    disabled={!canSendEmailVerification}
                    onclick={handleSendEmailVerification}
                  >
                    <MailCheck class="size-4" />
                    {$t(
                      isSendingEmailVerification
                        ? 'notifications.preferences.email.sending_code'
                        : 'notifications.preferences.email.send_code'
                    )}
                  </Button>
                </div>

                {#if notificationDraft.notification_email && !notificationEmailVerified}
                  <div class="grid gap-3 lg:grid-cols-[14rem_auto] lg:items-end">
                    <div class="space-y-2">
                      <label for="notification_email_code" class="text-sm font-medium">
                        {$t('notifications.preferences.email.code_label')}
                      </label>
                      <Input
                        id="notification_email_code"
                        inputmode="numeric"
                        maxlength={6}
                        bind:value={verificationCode}
                        disabled={isVerifyingNotificationEmail}
                      />
                    </div>
                    <Button
                      type="button"
                      disabled={isVerifyingNotificationEmail ||
                        notificationIsDirty ||
                        verificationCode.trim().length !== 6}
                      onclick={handleVerifyNotificationEmail}
                    >
                      <KeyRound class="size-4" />
                      {$t(
                        isVerifyingNotificationEmail
                          ? 'notifications.preferences.email.verifying'
                          : 'notifications.preferences.email.verify'
                      )}
                    </Button>
                  </div>
                {:else if notificationEmailVerified}
                  <div class="flex items-center gap-2 text-sm text-muted-foreground">
                    <CircleCheck class="size-4 text-emerald-600" />
                    <span>{$t('notifications.preferences.email.ready')}</span>
                  </div>
                {/if}
              </section>

              <section class="grid gap-3">
                <div class="flex items-center gap-2">
                  <MailCheck class="size-4 text-muted-foreground" />
                  <h3 class="text-sm font-semibold">
                    {$t('notifications.preferences.channel_title')}
                  </h3>
                </div>

                <div class="grid gap-2 lg:grid-cols-3">
                  {#each channelOptions as option (option.value)}
                    {@const active = notificationDraft.channel === option.value}
                    <Button
                      type="button"
                      variant={active ? 'default' : 'outline'}
                      class="h-full min-h-24 items-start justify-start gap-3 px-4 py-3 text-left whitespace-normal"
                      aria-pressed={active}
                      disabled={isSavingNotifications}
                      onclick={() => (notificationDraft.channel = option.value)}
                    >
                      <option.icon class="mt-0.5 size-4 shrink-0" />
                      <span class="flex min-w-0 flex-col items-start gap-1">
                        <span class="leading-tight">{option.label}</span>
                        <span
                          class={active
                            ? 'text-xs leading-snug wrap-break-word text-primary-foreground/80'
                            : 'text-xs leading-snug wrap-break-word text-muted-foreground'}
                        >
                          {option.description}
                        </span>
                      </span>
                    </Button>
                  {/each}
                </div>
              </section>

              <section class="grid gap-3">
                <div class="flex items-center gap-2">
                  <Clock3 class="size-4 text-muted-foreground" />
                  <h3 class="text-sm font-semibold">
                    {$t('notifications.preferences.frequency_title')}
                  </h3>
                </div>

                <div class="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
                  {#each frequencyOptions as option (option.value)}
                    {@const active = notificationDraft.frequency === option.value}
                    <Button
                      type="button"
                      variant={active ? 'default' : 'outline'}
                      class="h-full min-h-24 items-start justify-start gap-3 px-4 py-3 text-left whitespace-normal"
                      aria-pressed={active}
                      disabled={isSavingNotifications}
                      onclick={() => (notificationDraft.frequency = option.value)}
                    >
                      <option.icon class="mt-0.5 size-4 shrink-0" />
                      <span class="flex min-w-0 flex-col items-start gap-1">
                        <span class="leading-tight">{option.label}</span>
                        <span
                          class={active
                            ? 'text-xs leading-snug wrap-break-word text-primary-foreground/80'
                            : 'text-xs leading-snug wrap-break-word text-muted-foreground'}
                        >
                          {option.description}
                        </span>
                      </span>
                    </Button>
                  {/each}
                </div>
              </section>
            </div>

            <div class="mt-5 flex flex-col gap-2 border-t pt-4 sm:flex-row sm:justify-end">
              <Button
                type="button"
                variant="outline"
                disabled={isSavingNotifications || !notificationIsDirty}
                onclick={resetNotificationDraft}
              >
                {$t('common.reset')}
              </Button>
              <Button type="submit" disabled={isSavingNotifications || !notificationIsDirty}>
                {isSavingNotifications ? $t('common.saving') : $t('common.save_changes')}
              </Button>
            </div>
          {/if}
        </form>
      {:else if activeTab === 'password'}
        <form class="rounded-lg border bg-card p-4" onsubmit={handlePasswordSubmit}>
          <div class="mb-4">
            <h2 class="text-base font-semibold">{$t('pages.account_password_title')}</h2>
            <p class="text-sm text-muted-foreground">{$t('pages.account_password_desc')}</p>
          </div>

          <div class="grid gap-4 sm:max-w-md">
            <div class="flex flex-col gap-2">
              <label for="new_password" class="text-sm font-medium">{$t('auth.new_password')}</label
              >
              <Input
                id="new_password"
                type="password"
                bind:value={newPassword}
                required
                minlength={8}
              />
            </div>
            <div class="flex flex-col gap-2">
              <label for="confirm_password" class="text-sm font-medium">
                {$t('auth.confirm_password')}
              </label>
              <Input
                id="confirm_password"
                type="password"
                bind:value={confirmPassword}
                required
                minlength={8}
              />
            </div>
          </div>

          <div class="mt-4 flex justify-end">
            <Button type="submit" disabled={isSavingPassword}>
              {isSavingPassword ? $t('common.saving') : $t('pages.account_password_save')}
            </Button>
          </div>
        </form>
      {:else if activeTab === 'preferences'}
        <div class="rounded-lg border bg-card p-4">
          <div class="flex flex-col gap-1">
            <h2 class="text-base font-semibold">{$t('pages.settings_appearance')}</h2>
            <p class="text-sm text-muted-foreground">{$t('pages.settings_appearance_desc')}</p>
          </div>

          <section class="mt-4 grid gap-2 sm:grid-cols-3">
            {#each options as opt (opt.value)}
              {@const active = $themePreference === opt.value}
              <Button
                variant={active ? 'default' : 'outline'}
                class="h-full items-start justify-start gap-3 px-4 py-3 text-left whitespace-normal"
                aria-pressed={active}
                onclick={() => setThemePreference(opt.value)}
              >
                <opt.icon class="mt-0.5 size-4 shrink-0" />
                <span class="flex min-w-0 flex-col items-start gap-0.5 text-left">
                  <span class="leading-tight">{opt.label}</span>
                  <span
                    class={active
                      ? 'text-xs leading-snug wrap-break-word text-primary-foreground/80'
                      : 'text-xs leading-snug wrap-break-word text-muted-foreground'}
                  >
                    {opt.description}
                  </span>
                </span>
              </Button>
            {/each}
          </section>

          <section class="mt-6 border-t pt-5">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div class="flex min-w-0 items-start gap-2">
                <Contrast class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                <div class="min-w-0 space-y-1">
                  <h3 class="text-sm font-semibold">{$t('pages.settings_contrast')}</h3>
                  <p class="text-sm text-muted-foreground">{$t('pages.settings_contrast_desc')}</p>
                </div>
              </div>
              <div class="flex shrink-0 items-center gap-2">
                <span class="rounded-md bg-muted px-2 py-1 text-sm font-medium tabular-nums">
                  {$contrastPreference}%
                </span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={$contrastPreference === DEFAULT_CONTRAST_PREFERENCE}
                  onclick={() => setContrastPreference(DEFAULT_CONTRAST_PREFERENCE)}
                >
                  <RotateCcw class="size-4" />
                  {$t('pages.settings_contrast_reset')}
                </Button>
              </div>
            </div>

            <div class="mt-4">
              <label for="appearance_contrast" class="sr-only">
                {$t('pages.settings_contrast')}
              </label>
              <input
                id="appearance_contrast"
                type="range"
                min={MIN_CONTRAST_PREFERENCE}
                max={MAX_CONTRAST_PREFERENCE}
                step={CONTRAST_PREFERENCE_STEP}
                value={$contrastPreference}
                class="h-2 w-full cursor-pointer appearance-none rounded-full bg-muted accent-primary"
                oninput={handleContrastInput}
              />
              <div class="mt-2 flex justify-between text-xs text-muted-foreground">
                <span>{$t('pages.settings_contrast_min')}</span>
                <span>{$t('pages.settings_contrast_default')}</span>
                <span>{$t('pages.settings_contrast_max')}</span>
              </div>
            </div>
          </section>

          <section class="mt-6 border-t pt-5">
            <div class="flex min-w-0 items-start gap-2">
              <Type class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
              <div class="min-w-0 space-y-1">
                <h3 class="text-sm font-semibold">{$t('pages.settings_font')}</h3>
                <p class="text-sm text-muted-foreground">{$t('pages.settings_font_desc')}</p>
              </div>
            </div>

            <div class="mt-4 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
              {#each fontOptions as font (font.value)}
                {@const active = $fontPreference === font.value}
                <Button
                  type="button"
                  variant={active ? 'default' : 'outline'}
                  class="h-full min-h-28 items-start justify-start gap-3 px-4 py-3 text-left whitespace-normal"
                  aria-pressed={active}
                  onclick={() => setFontPreference(font.value)}
                >
                  <Type class="mt-0.5 size-4 shrink-0" />
                  <span class="flex min-w-0 flex-col items-start gap-1 text-left">
                    <span class="leading-tight" style={`font-family: ${font.stack}`}>
                      {font.label}
                    </span>
                    <span
                      class={active
                        ? 'text-xs leading-snug wrap-break-word text-primary-foreground/80'
                        : 'text-xs leading-snug wrap-break-word text-muted-foreground'}
                    >
                      {font.description}
                    </span>
                    <span
                      class={active
                        ? 'font-mono text-[11px] leading-snug wrap-break-word text-primary-foreground/70'
                        : 'font-mono text-[11px] leading-snug wrap-break-word text-muted-foreground'}
                    >
                      {font.stack}
                    </span>
                  </span>
                </Button>
              {/each}
            </div>
          </section>
        </div>
      {/if}
    </section>
  </div>
</div>
