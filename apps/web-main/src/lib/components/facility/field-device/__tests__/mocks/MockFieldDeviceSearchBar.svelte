<script lang="ts">
  interface Props {
    onBulkDelete: () => void;
    onRefresh: () => void;
  }

  let { onBulkDelete, onRefresh }: Props = $props();
</script>

<button data-testid="bulk-delete" onclick={onBulkDelete}>bulk delete</button>
<button data-testid="refresh" onclick={onRefresh}>refresh</button>
