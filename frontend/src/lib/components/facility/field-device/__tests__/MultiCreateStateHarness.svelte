<script lang="ts">
  import { FieldDeviceMultiCreateState } from '../multi-create/FieldDeviceMultiCreateState.svelte.js';

  const state = new FieldDeviceMultiCreateState();

  function setSps(value: string) {
    state.handleSpsSystemTypeChange(value);
  }

  function setPreselection(objectDataId: string) {
    state.handlePreselectionChange({
      objectDataId,
      apparatId: 'app-1',
      systemPartId: 'sp-1'
    });
  }
</script>

<div data-testid="selection">
  {state.selection.spsControllerSystemTypeId}|{state.selection.objectDataId}|{state.selection
    .apparatId}|{state.selection.systemPartId}
</div>
<div data-testid="rows-count">{state.rows.length}</div>
<div data-testid="available-numbers">{state.availableNumbers.join(',')}</div>
<div data-testid="loading-available">{String(state.loadingAvailableNumbers)}</div>
<div data-testid="object-data-id">{state.selectedObjectData?.id ?? ''}</div>
<div data-testid="object-data-loading">{String(state.loadingObjectDataPreview)}</div>
<div data-testid="object-data-error">{state.objectDataPreviewError}</div>

<button data-testid="set-sps-1" onclick={() => setSps('sps-1')}>set sps 1</button>
<button data-testid="set-sps-2" onclick={() => setSps('sps-2')}>set sps 2</button>
<button data-testid="set-object-1" onclick={() => setPreselection('obj-1')}>set object 1</button>
<button data-testid="set-object-2" onclick={() => setPreselection('obj-2')}>set object 2</button>
