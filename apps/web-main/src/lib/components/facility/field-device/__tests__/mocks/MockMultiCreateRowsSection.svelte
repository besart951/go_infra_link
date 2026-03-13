<script lang="ts">
  import type {
    FieldDeviceRowData,
    FieldDeviceRowError
  } from '$lib/domain/facility/fieldDeviceMultiCreate.js';

  interface Props {
    rows: FieldDeviceRowData[];
    rowErrors: Map<number, FieldDeviceRowError>;
    onAddRow: () => void;
    onSubmit: () => void;
    onRowBmkChange: (index: number, value: string) => void;
    onRowDescriptionChange: (index: number, value: string) => void;
    onRowTextFixChange: (index: number, value: string) => void;
  }

  let {
    rows,
    onAddRow,
    onSubmit,
    onRowBmkChange,
    onRowDescriptionChange,
    onRowTextFixChange
  }: Props = $props();
</script>

<div data-testid="rows-count">{rows.length}</div>
<button data-testid="add-row" onclick={onAddRow}>add row</button>
<button
  data-testid="set-row-values"
  onclick={() => {
    onRowBmkChange(0, 'BMK-NEW');
    onRowDescriptionChange(0, 'New Description');
    onRowTextFixChange(0, 'TextFix-NEW');
  }}
>
  set row
</button>
<button data-testid="submit-multi-create" onclick={onSubmit}>submit</button>
