<script lang="ts">
    import { onMount } from 'svelte';
    import { goto, invalidateAll } from '$app/navigation';
    import { api } from '$lib/api/client';

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

<div class="flex items-center justify-center h-full p-8">
    <p class="text-muted-foreground">Logging out...</p>
</div>
