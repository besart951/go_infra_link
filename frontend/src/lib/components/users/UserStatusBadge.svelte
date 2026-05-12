<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { UserDirectoryUser } from '$lib/infrastructure/api/userRepository.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type Props = {
    user: UserDirectoryUser;
  };

  let { user }: Props = $props();
  const t = createTranslator();
</script>

{#if user.disabled_at}
  <Badge variant="destructive">{$t('common.disabled')}</Badge>
{:else if user.locked_until}
  <Badge variant="warning">{$t('common.locked')}</Badge>
{:else if user.is_active}
  <Badge variant="success">{$t('common.active')}</Badge>
{:else}
  <Badge variant="outline">{$t('common.inactive')}</Badge>
{/if}
