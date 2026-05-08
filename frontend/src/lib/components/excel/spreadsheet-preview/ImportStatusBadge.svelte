<script lang="ts">
  import type { ImportNodeViewState } from './fieldDeviceImportPresentation.js';

  interface Props {
    node: ImportNodeViewState;
    class?: string;
  }

  let { node, class: className = '' }: Props = $props();
</script>

<span
  class={`shrink-0 rounded border px-2 py-0.5 text-[11px] font-medium ${node.badgeClass} ${className}`}
>
  {node.label}
</span>
