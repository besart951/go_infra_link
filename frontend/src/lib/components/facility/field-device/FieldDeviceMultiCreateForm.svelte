<script lang="ts">
  import { onDestroy } from 'svelte';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { AlertCircle } from '@lucide/svelte';
  import MultiCreateSelectionSection from './multi-create/MultiCreateSelectionSection.svelte';
  import MultiCreateRowsSection from './multi-create/MultiCreateRowsSection.svelte';
  import { FieldDeviceMultiCreateState } from './multi-create/FieldDeviceMultiCreateState.svelte.js';

  import type { FieldDevice } from '$lib/domain/facility/index.js';

  interface Props {
    projectId?: string;
    systemTypeRefreshKey?: string | number;
    onSuccess?: (createdDevices: FieldDevice[]) => void;
    onCancel?: () => void;
  }

  let { projectId, systemTypeRefreshKey, onSuccess, onCancel }: Props = $props();

  const state = new FieldDeviceMultiCreateState({
    projectId: () => projectId,
    onSuccess: () => onSuccess
  });

  onDestroy(function () {
    state.destroy();
  });
</script>

<div>
  {#if state.globalError}
    <Alert.Root variant="destructive">
      <AlertCircle class="size-4" />
      <Alert.Description>{state.globalError}</Alert.Description>
    </Alert.Root>
  {/if}

  <MultiCreateSelectionSection {state} {systemTypeRefreshKey} />

  {#if state.showRowsSection}
    <MultiCreateRowsSection {state} {onCancel} />
  {/if}
</div>
