<script lang="ts">
  import LoadingRows from '../table-loading-rows.svelte';

  let { loading, delayMs = 150 } = $props<{
    loading: boolean;
    delayMs?: number;
  }>();
</script>

<table>
  <tbody>
    <LoadingRows {loading} {delayMs} rowCount={2} columnCount={3} />
  </tbody>
</table>
