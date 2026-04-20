<script lang="ts">
  interface Props {
    state: {
      handleSpsSystemTypeChange: (value: string) => void;
      handlePreselectionChange: (value: {
        objectDataId: string;
        apparatId: string;
        systemPartId: string;
      }) => void;
    };
    systemTypeRefreshKey?: string | number;
  }

  let { state }: Props = $props();
</script>

<button data-testid="set-selection" onclick={() => state.handleSpsSystemTypeChange('sps-1')}>
  set sps
</button>
<button
  data-testid="set-preselection"
  onclick={() =>
    state.handlePreselectionChange({
      objectDataId: 'obj-1',
      apparatId: 'app-1',
      systemPartId: 'sp-1'
    })}
>
  set preselection
</button>
