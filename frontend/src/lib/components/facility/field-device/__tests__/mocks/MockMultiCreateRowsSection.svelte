<script lang="ts">
  interface Props {
    state: {
      rows: Array<{ apparatNr: number | null }>;
      addRow: () => void;
      handleSubmit: () => void | Promise<void>;
      handleRowBmkChange: (index: number, value: string) => void;
      handleRowDescriptionChange: (index: number, value: string) => void;
      handleRowTextFixChange: (index: number, value: string) => void;
    };
    onCancel?: () => void;
  }

  let { state }: Props = $props();
</script>

<div data-testid="rows-count">{state.rows.length}</div>
<button data-testid="add-row" onclick={() => state.addRow()}>add row</button>
<button
  data-testid="set-row-values"
  onclick={() => {
    state.handleRowBmkChange(0, 'BMK-NEW');
    state.handleRowDescriptionChange(0, 'New Description');
    state.handleRowTextFixChange(0, 'TextFix-NEW');
  }}
>
  set row
</button>
<button data-testid="submit-multi-create" onclick={() => state.handleSubmit()}>submit</button>
