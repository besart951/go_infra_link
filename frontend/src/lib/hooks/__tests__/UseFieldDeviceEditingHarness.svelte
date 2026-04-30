<script lang="ts">
  import { useFieldDeviceEditing } from '../useFieldDeviceEditing.svelte.js';

  type EditingOptions = NonNullable<Parameters<typeof useFieldDeviceEditing>[0]>;

  interface Props {
    onReady: (editing: ReturnType<typeof useFieldDeviceEditing>) => void;
    projectId?: string;
    onSharedStateChange?: EditingOptions['onSharedStateChange'];
    onSaveSuccess?: EditingOptions['onSaveSuccess'];
  }

  let { onReady, projectId, onSharedStateChange, onSaveSuccess }: Props = $props();
  const editing = useFieldDeviceEditing({
    projectId: () => projectId,
    onSharedStateChange: (state) => onSharedStateChange?.(state),
    onSaveSuccess: (deviceIds) => onSaveSuccess?.(deviceIds)
  });

  $effect(() => {
    onReady(editing);
  });
</script>
