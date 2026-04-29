<script lang="ts">
  import { page } from '$app/stores';
  import HttpErrorPage from '$lib/components/http-error-page.svelte';

  const status = $derived($page.status ?? 500);
  const fallbackMessage = $derived($page.error?.message);
</script>

<HttpErrorPage {status} {fallbackMessage} />
