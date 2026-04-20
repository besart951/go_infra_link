<script lang="ts">
  let { state } = $props<{ state?: { items?: unknown[] } }>();
</script>

<div data-testid="paginated-list" data-count={String(state?.items?.length ?? 0)}></div>
