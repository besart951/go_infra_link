<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createTranslator } from '$lib/i18n/translator.js';

  const t = createTranslator();

  onMount(() => {
    goto('/projects/list');
  });
</script>

<div class="p-6 text-sm text-muted-foreground">{$t('projects.redirect')}</div>
