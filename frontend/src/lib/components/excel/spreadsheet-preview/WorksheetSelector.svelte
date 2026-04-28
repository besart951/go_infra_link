<script lang="ts">
  interface Props {
    worksheets: string[];
    selectedWorksheetName: string;
    disabled?: boolean;
    onSelect: (name: string) => void;
  }

  let { worksheets, selectedWorksheetName, disabled = false, onSelect }: Props = $props();

  function handleChange(event: Event): void {
    const target = event.currentTarget as HTMLSelectElement;
    onSelect(target.value);
  }
</script>

<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
  <div>
    <h2 class="text-base font-semibold">Arbeitsblätter</h2>
    <p class="text-sm text-muted-foreground">{worksheets.length} Arbeitsblätter gefunden</p>
  </div>

  <label class="flex flex-col gap-1 text-sm sm:min-w-72">
    <span class="font-medium">Aktives Arbeitsblatt</span>
    <select
      class="h-9 rounded-md border border-input bg-background px-3 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50"
      value={selectedWorksheetName}
      onchange={handleChange}
      disabled={disabled || worksheets.length === 0}
    >
      {#each worksheets as worksheet}
        <option value={worksheet}>{worksheet}</option>
      {/each}
    </select>
  </label>
</div>
