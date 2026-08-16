<script lang="ts">
  import { onMount } from 'svelte';
  import { goto, invalidateAll } from '$app/navigation';
  import { api } from '$lib/api/client';
  import { systemNotificationState } from '$lib/components/notifications/SystemNotificationState.svelte.js';
  import { createTranslator } from '$lib/i18n/translator';
  import { facilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';
  import { facilityDetailCache } from '$lib/services/facilityDetailCache.js';

  const t = createTranslator();

  onMount(async () => {
    facilityReferenceDataCache.stop({ immediate: true });
    facilityDetailCache.stop();
    systemNotificationState.disconnect({ immediate: true });

    try {
      await api('/auth/logout', { method: 'POST' });
      await invalidateAll();
    } catch (e) {
      console.error('Logout failed', e);
    } finally {
      await goto('/login');
    }
  });
</script>

<svelte:head>
  <title>{$t('navigation.logout')} | Infra Link</title>
</svelte:head>

<div class="flex h-full items-center justify-center p-8">
  <p class="text-muted-foreground">{$t('messages.logout_in_progress')}</p>
</div>
