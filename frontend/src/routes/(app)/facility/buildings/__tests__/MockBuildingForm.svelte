<script lang="ts">
  import type { Building } from '$lib/domain/facility/index.js';

  interface Props {
    initialData?: Building;
    onSuccess?: () => void | Promise<void>;
    onCancel?: () => void;
  }

  let { initialData, onSuccess, onCancel }: Props = $props();
</script>

<section data-testid="building-form">
  <span data-testid="form-mode">{initialData?.id ?? 'create'}</span>
  <button type="button" onclick={() => onSuccess?.()}>form-success</button>
  <button type="button" onclick={() => onCancel?.()}>form-cancel</button>
</section>
