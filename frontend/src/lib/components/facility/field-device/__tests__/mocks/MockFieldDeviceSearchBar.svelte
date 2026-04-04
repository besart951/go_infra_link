<script lang="ts">
  import { useFieldDeviceState } from '../../state/context.svelte.js';

  const state = useFieldDeviceState();
</script>

<button data-testid="bulk-delete" onclick={() => void state.bulkDeleteSelected()}
  >bulk delete</button
>
<button data-testid="refresh" onclick={() => void state.reload()}>refresh</button>
