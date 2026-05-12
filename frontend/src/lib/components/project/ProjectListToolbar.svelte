<script lang="ts">
  import { Label } from '$lib/components/ui/label/index.js';
  import type { ProjectStatusFilter } from '$lib/stores/projects/projectListStore.js';

  type StatusOption = {
    value: ProjectStatusFilter;
    label: string;
  };

  type Props = {
    label: string;
    value: ProjectStatusFilter;
    options: readonly StatusOption[];
    disabled?: boolean;
    onStatusChange: (status: ProjectStatusFilter) => void;
  };

  let { label, value, options, disabled = false, onStatusChange }: Props = $props();

  function isProjectStatusFilter(value: string): value is ProjectStatusFilter {
    return value === 'all' || value === 'planned' || value === 'ongoing' || value === 'completed';
  }

  function handleStatusChange(event: Event): void {
    const target = event.currentTarget;
    if (!(target instanceof HTMLSelectElement)) return;
    if (!isProjectStatusFilter(target.value)) return;

    onStatusChange(target.value);
  }
</script>

<div class="rounded-xl border bg-card p-4 shadow-sm">
  <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
    <Label for="project_status_filter">{label}</Label>

    <select
      id="project_status_filter"
      class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs transition-[color,box-shadow] outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 sm:w-[220px]"
      {disabled}
      {value}
      onchange={handleStatusChange}
    >
      {#each options as option (option.value)}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
  </div>
</div>
