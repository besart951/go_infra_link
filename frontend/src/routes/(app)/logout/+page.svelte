<script lang="ts">
    import { onMount } from 'svelte';
    import { goto, invalidateAll } from '$app/navigation';
    import { api } from '$lib/api/client';
    import { createTranslator } from '$lib/i18n/translator';

    const t = createTranslator();

    onMount(async () => {
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

<div class="flex items-center justify-center h-full p-8">
    <p class="text-muted-foreground">{$t('messages.logout_in_progress')}</p>
</div>
