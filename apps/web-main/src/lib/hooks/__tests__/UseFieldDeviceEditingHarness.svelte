<script lang="ts">
  import { useFieldDeviceEditing } from '../useFieldDeviceEditing.svelte.js';

  interface Props {
    onReady: (editing: ReturnType<typeof useFieldDeviceEditing>) => void;
  }

  let { onReady }: Props = $props();
  const editing = useFieldDeviceEditing();

  $effect(() => {
    onReady(editing);
  });
</script>
