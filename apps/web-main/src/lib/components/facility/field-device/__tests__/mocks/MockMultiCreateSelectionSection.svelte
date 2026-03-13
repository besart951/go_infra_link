<script lang="ts">
  interface Props {
    onSpsSystemTypeChange: (value: string) => void;
    onPreselectionChange: (value: {
      objectDataId: string;
      apparatId: string;
      systemPartId: string;
    }) => void;
  }

  let { onSpsSystemTypeChange, onPreselectionChange }: Props = $props();
</script>

<button data-testid="set-selection" onclick={() => onSpsSystemTypeChange('sps-1')}>set sps</button>
<button
  data-testid="set-preselection"
  onclick={() =>
    onPreselectionChange({
      objectDataId: 'obj-1',
      apparatId: 'app-1',
      systemPartId: 'sp-1'
    })}
>
  set preselection
</button>
