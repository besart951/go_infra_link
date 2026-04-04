<script lang="ts">
  import { useFieldDeviceState } from '../../state/context.svelte.js';

  const state = useFieldDeviceState();
</script>

<div data-testid="table-count">{state.items.length}</div>
{#each state.items as item (item.id)}
  <button data-testid={`select-${item.id}`} onclick={() => state.toggleSelection(item.id)}>
    select
  </button>
  <button data-testid={`delete-${item.id}`} onclick={() => void state.deleteDevice(item)}>
    delete
  </button>
{/each}
