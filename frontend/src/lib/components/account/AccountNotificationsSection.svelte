<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { createTranslator } from '$lib/i18n/translator';
  import {
    Bell,
    CircleCheck,
    Clock3,
    KeyRound,
    Mail,
    MailCheck,
    MonitorCheck
  } from '@lucide/svelte';
  import type { Component } from 'svelte';
  import type { User } from '$lib/infrastructure/api/userRepository.js';
  import type {
    NotificationChannel,
    NotificationFrequency,
    NotificationPreferenceFormValues
  } from '$lib/domain/notification/index.js';

  type NotificationOption<TValue extends string> = {
    value: TValue;
    label: string;
    description: string;
    icon: Component;
  };

  interface Props {
    currentUser: User;
    notificationDraft: NotificationPreferenceFormValues;
    verificationCode: string;
    notificationsError: string | null;
    notificationEmailVerified: boolean;
    notificationIsDirty: boolean;
    canSendEmailVerification: boolean;
    isLoadingNotifications: boolean;
    isSavingNotifications: boolean;
    isSendingEmailVerification: boolean;
    isVerifyingNotificationEmail: boolean;
    onSubmit: (event: SubmitEvent) => void | Promise<void>;
    onReset: () => void;
    onSendEmailVerification: () => void | Promise<void>;
    onVerifyNotificationEmail: () => void | Promise<void>;
  }

  let {
    currentUser,
    notificationDraft = $bindable(),
    verificationCode = $bindable(),
    notificationsError,
    notificationEmailVerified,
    notificationIsDirty,
    canSendEmailVerification,
    isLoadingNotifications,
    isSavingNotifications,
    isSendingEmailVerification,
    isVerifyingNotificationEmail,
    onSubmit,
    onReset,
    onSendEmailVerification,
    onVerifyNotificationEmail
  }: Props = $props();

  const t = createTranslator();

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
</script>

<form class="rounded-lg border bg-card p-4" onsubmit={onSubmit}>
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
            onclick={onSendEmailVerification}
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
              onclick={onVerifyNotificationEmail}
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
        onclick={onReset}
      >
        {$t('common.reset')}
      </Button>
      <Button type="submit" disabled={isSavingNotifications || !notificationIsDirty}>
        {isSavingNotifications ? $t('common.saving') : $t('common.save_changes')}
      </Button>
    </div>
  {/if}
</form>
