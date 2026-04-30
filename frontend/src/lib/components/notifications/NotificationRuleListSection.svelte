<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import type { NotificationRule } from '$lib/domain/notification/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';

  type SelectOption = {
    id: string;
    label: string;
  };

  interface Props {
    rules: NotificationRule[];
    isLoading: boolean;
    isSubmitting: boolean;
    roleOptions: SelectOption[];
    onDeleteRule: (rule: NotificationRule) => void | Promise<void>;
  }

  let { rules, isLoading, isSubmitting, roleOptions, onDeleteRule }: Props = $props();
  const t = createTranslator();

  function displayRole(role: string): string {
    return roleOptions.find((option) => option.id === role)?.label || role;
  }
</script>

<div class="space-y-2">
  {#if isLoading && rules.length === 0}
    <div class="py-4 text-sm text-muted-foreground">{$t('common.loading')}</div>
  {:else if rules.length === 0}
    <div class="py-4 text-sm text-muted-foreground">{$t('notifications.rules.empty')}</div>
  {:else}
    {#each rules as rule (rule.id)}
      <div class="flex items-start justify-between gap-3 rounded-md border px-3 py-2">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-medium">{rule.name}</span>
            <span class="rounded-full bg-muted px-2 py-0.5 text-xs">{rule.event_key}</span>
          </div>
          <p class="mt-1 text-xs text-muted-foreground">
            {$t(`notifications.rules.recipient_${rule.recipient_type}`)}
            {#if rule.recipient_role}
              · {displayRole(rule.recipient_role)}
            {/if}
            {#if rule.project_id}
              · {$t('notifications.rules.project_id')}: {rule.project_id}
            {/if}
            {#if rule.resource_type}
              · {$t('notifications.rules.resource_type')}: {$t(
                `notifications.rules.resource_types.${rule.resource_type}`
              )}
            {/if}
          </p>
        </div>
        <Button
          variant="ghost"
          size="icon-sm"
          onclick={() => onDeleteRule(rule)}
          disabled={isSubmitting}
        >
          <Trash2Icon class="size-4" />
        </Button>
      </div>
    {/each}
  {/if}
</div>
