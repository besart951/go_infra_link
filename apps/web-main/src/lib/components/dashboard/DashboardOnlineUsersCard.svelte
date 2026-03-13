<script lang="ts">
  import * as Card from '@ui-svelte/components/ui/card/index.js';
  import { Badge } from '@ui-svelte/components/ui/badge/index.js';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import type { DashboardUserPresence } from '$lib/domain/dashboard/index.js';
  import { createTranslator } from '@i18n/translator.js';

  type Props = {
    users: DashboardUserPresence[];
  };

  let { users }: Props = $props();
  const t = createTranslator();

  function formatLastLogin(value?: string): string {
    if (!value) return $t('dashboard.no_recent_login');
    return `${$t('dashboard.last_login_prefix')} ${new Date(value).toLocaleString()}`;
  }
</script>

<Card.Root>
  <Card.Header>
    <Card.Title>{$t('dashboard.online_title')}</Card.Title>
    <Card.Description>{$t('dashboard.online_desc')}</Card.Description>
  </Card.Header>
  <Card.Content>
    {#if users.length === 0}
      <p class="text-sm text-muted-foreground">{$t('dashboard.no_online')}</p>
    {:else}
      <div class="space-y-3">
        {#each users as user (user.id)}
          <div class="flex items-center justify-between gap-3">
            <div class="flex min-w-0 items-center gap-2">
              <UserAvatar firstName={user.first_name} lastName={user.last_name} />
              <div class="min-w-0">
                <p class="truncate text-sm font-medium">{user.first_name} {user.last_name}</p>
                <p class="truncate text-xs text-muted-foreground">
                  {formatLastLogin(user.last_login_at)}
                </p>
              </div>
            </div>
            <Badge variant="success">{$t('dashboard.online_badge')}</Badge>
          </div>
        {/each}
      </div>
    {/if}
  </Card.Content>
</Card.Root>
