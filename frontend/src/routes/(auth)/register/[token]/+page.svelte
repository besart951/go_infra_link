<script lang="ts">
  import { goto } from '$app/navigation';
  import {
    completeRegistration,
    getRegistrationErrorMessage,
    getRegistrationFieldErrors
  } from '$lib/api/registrations.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import {
    Field,
    FieldContent,
    FieldError,
    FieldGroup,
    FieldLabel
  } from '$lib/components/ui/field/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  let { data } = $props();

  let firstName = $state('');
  let lastName = $state('');
  let password = $state('');
  let privacyAck = $state(false);
  let isSubmitting = $state(false);
  let submitError = $state<string | null>(null);
  let fieldErrors = $state<Record<string, string>>({});
  let visibleError = $derived(submitError ?? data.error);

  function errorText(code: string | null): string | null {
    switch (code) {
      case 'registration_expired':
        return $t('registration.expired');
      case 'registration_invalid':
        return $t('registration.invalid');
      case 'registration_unavailable':
        return $t('registration.unavailable');
      default:
        return code;
    }
  }

  async function handleSubmit(event: Event) {
    event.preventDefault();
    if (!data.registration) return;

    isSubmitting = true;
    submitError = null;
    fieldErrors = {};

    try {
      await completeRegistration(data.token, {
        first_name: firstName,
        last_name: lastName,
        password,
        privacy_ack: privacyAck
      });
      await goto('/login');
    } catch (err) {
      submitError = getRegistrationErrorMessage(err);
      fieldErrors = getRegistrationFieldErrors(err);
    } finally {
      isSubmitting = false;
    }
  }
</script>

<svelte:head>
  <title>{$t('registration.title')} | Infra Link</title>
</svelte:head>

<div class="space-y-6">
  <div class="space-y-2">
    <h1 class="text-2xl font-semibold tracking-tight">{$t('registration.title')}</h1>
    {#if data.registration}
      <p class="text-sm text-muted-foreground">
        {$t('registration.account_for', { email: data.registration.email })}
      </p>
    {:else}
      <p class="text-sm text-muted-foreground">{$t('registration.checking_invitation')}</p>
    {/if}
  </div>

  {#if visibleError}
    <div
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {errorText(visibleError)}
    </div>
  {/if}

  {#if data.registration}
    <form onsubmit={handleSubmit} class="space-y-6">
      <FieldGroup>
        <Field>
          <FieldLabel for="firstName">{$t('registration.first_name')}</FieldLabel>
          <FieldContent>
            <Input id="firstName" bind:value={firstName} autocomplete="given-name" required />
          </FieldContent>
          {#if fieldErrors.first_name}
            <FieldError>{fieldErrors.first_name}</FieldError>
          {/if}
        </Field>

        <Field>
          <FieldLabel for="lastName">{$t('registration.last_name')}</FieldLabel>
          <FieldContent>
            <Input id="lastName" bind:value={lastName} autocomplete="family-name" required />
          </FieldContent>
          {#if fieldErrors.last_name}
            <FieldError>{fieldErrors.last_name}</FieldError>
          {/if}
        </Field>

        <Field>
          <FieldLabel for="password">{$t('registration.password')}</FieldLabel>
          <FieldContent>
            <Input
              id="password"
              type="password"
              bind:value={password}
              autocomplete="new-password"
              minlength={8}
              required
            />
          </FieldContent>
          {#if fieldErrors.password}
            <FieldError>{fieldErrors.password}</FieldError>
          {/if}
        </Field>
      </FieldGroup>

      <div class="rounded-md border bg-muted/30 p-3 text-sm text-muted-foreground">
        {$t('registration.privacy_notice')}
      </div>

      <label class="flex items-start gap-2 text-sm">
        <Checkbox checked={privacyAck} onCheckedChange={(value) => (privacyAck = value === true)} />
        <span>{$t('registration.privacy_ack')}</span>
      </label>
      {#if fieldErrors.privacy_ack}
        <p class="text-sm text-destructive">{fieldErrors.privacy_ack}</p>
      {/if}

      <Button type="submit" class="w-full" disabled={isSubmitting || !privacyAck}>
        {isSubmitting ? $t('registration.submitting') : $t('registration.complete')}
      </Button>
    </form>
  {/if}
</div>
