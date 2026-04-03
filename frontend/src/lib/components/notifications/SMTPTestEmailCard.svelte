<script lang="ts">
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import type {
    SendSMTPTestEmailRequest,
    SMTPSettings,
    SMTPTestEmailFormValues
  } from '$lib/domain/notification/index.js';
  import {
    createSMTPTestEmailFormValues,
    normalizeSMTPTestEmailInput,
    validateSMTPTestEmailInput
  } from '$lib/domain/notification/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  interface Props {
    settings?: SMTPSettings | null;
    defaultRecipient?: string;
    isSubmitting?: boolean;
    formError?: string | null;
    fieldErrors?: Record<string, string>;
    onSubmit?: (payload: SendSMTPTestEmailRequest) => Promise<void> | void;
  }

  let {
    settings = null,
    defaultRecipient = '',
    isSubmitting = false,
    formError = null,
    fieldErrors = {},
    onSubmit
  }: Props = $props();

  const t = createTranslator();

  let draft = $state<SMTPTestEmailFormValues>(createSMTPTestEmailFormValues(''));
  let localErrors = $state<Record<string, string>>({});
  let previousDefaultRecipient = $state('');

  $effect(() => {
    if (defaultRecipient !== previousDefaultRecipient) {
      if (!draft.to || draft.to === previousDefaultRecipient) {
        draft.to = defaultRecipient;
      }
      previousDefaultRecipient = defaultRecipient;
    }
  });

  function errorFor(field: string): string | undefined {
    const key = fieldErrors[field] ?? localErrors[field];
    return key ? $t(key) : undefined;
  }

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();

    const payload = normalizeSMTPTestEmailInput(draft);
    localErrors = validateSMTPTestEmailInput(payload);

    if (Object.keys(localErrors).length > 0) {
      return;
    }

    await onSubmit?.(payload);
  }
</script>

<Card.Root>
  <Card.Header class="gap-3">
    <Card.Title>{$t('notifications.test.title')}</Card.Title>
    <Card.Description>{$t('notifications.test.description')}</Card.Description>
  </Card.Header>

  <Card.Content class="space-y-4">
    {#if !settings}
      <Alert.Root>
        <Alert.Description>{$t('notifications.test.missing_config')}</Alert.Description>
      </Alert.Root>
    {:else}
      {#if formError}
        <Alert.Root variant="destructive">
          <Alert.Description>{formError}</Alert.Description>
        </Alert.Root>
      {/if}

      {#if !settings.enabled}
        <Alert.Root>
          <Alert.Description>{$t('notifications.test.disabled_hint')}</Alert.Description>
        </Alert.Root>
      {/if}

      <form class="space-y-4" onsubmit={handleSubmit}>
        <div class="space-y-2">
          <Label for="smtp_test_to">{$t('notifications.test.to')}</Label>
          <Input
            id="smtp_test_to"
            type="email"
            bind:value={draft.to}
            disabled={isSubmitting}
            placeholder={$t('notifications.test.to_placeholder')}
            aria-invalid={Boolean(errorFor('to'))}
          />
          {#if errorFor('to')}
            <p class="text-sm text-destructive">{errorFor('to')}</p>
          {/if}
        </div>

        <div class="space-y-2">
          <Label for="smtp_test_subject">
            {$t('notifications.test.subject')}
            <span class="ms-2 text-xs font-normal text-muted-foreground"
              >{$t('pages.optional')}</span
            >
          </Label>
          <Input
            id="smtp_test_subject"
            bind:value={draft.subject}
            disabled={isSubmitting}
            placeholder={$t('notifications.test.subject_placeholder')}
          />
        </div>

        <div class="space-y-2">
          <Label for="smtp_test_body">
            {$t('notifications.test.body')}
            <span class="ms-2 text-xs font-normal text-muted-foreground"
              >{$t('pages.optional')}</span
            >
          </Label>
          <Textarea
            id="smtp_test_body"
            bind:value={draft.body}
            disabled={isSubmitting}
            rows={6}
            placeholder={$t('notifications.test.body_placeholder')}
          />
        </div>

        <div class="flex justify-end">
          <Button type="submit" disabled={isSubmitting}>
            {$t(isSubmitting ? 'common.saving' : 'notifications.test.send')}
          </Button>
        </div>
      </form>
    {/if}
  </Card.Content>
</Card.Root>
