<script lang="ts">
  import HttpErrorPage from '$lib/components/http-error-page.svelte';
  import type { PageData } from './$types.js';

  let { data }: { data: PageData } = $props();
</script>

<HttpErrorPage status={data.status} from={data.from} />
