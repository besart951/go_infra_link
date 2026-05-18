<script lang="ts">
  import { Label } from '$lib/components/ui/label/index.js';
  import ProjectPhaseSelect from '$lib/components/project/ProjectPhaseSelect.svelte';
  import type { ProjectStatusFilter } from '$lib/stores/projects/projectListStore.js';

  type StatusOption = {
    value: ProjectStatusFilter;
    label: string;
  };

  type Props = {
    statusLabel: string;
    statusValue: ProjectStatusFilter;
    options: readonly StatusOption[];
    phaseLabel: string;
    phaseValue: string;
    allPhasesLabel: string;
    phaseSearchPlaceholder: string;
    phaseEmptyText: string;
    disabled?: boolean;
    onStatusChange: (status: ProjectStatusFilter) => void;
    onPhaseChange: (phaseId: string) => void;
  };

  let {
    statusLabel,
    statusValue,
    options,
    phaseLabel,
    phaseValue,
    allPhasesLabel,
    phaseSearchPlaceholder,
    phaseEmptyText,
    disabled = false,
    onStatusChange,
    onPhaseChange
  }: Props = $props();

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
  <div class="grid gap-4 md:grid-cols-2">
    <div class="flex flex-col gap-2">
      <Label for="project_status_filter">{statusLabel}</Label>

      <select
        id="project_status_filter"
        class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs transition-[color,box-shadow] outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50"
        {disabled}
        value={statusValue}
        onchange={handleStatusChange}
      >
        {#each options as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </div>

    <div class="flex flex-col gap-2">
      <Label for="project_phase_filter">{phaseLabel}</Label>
      <ProjectPhaseSelect
        id="project_phase_filter"
        value={phaseValue}
        width="w-full"
        placeholder={allPhasesLabel}
        clearText={allPhasesLabel}
        searchPlaceholder={phaseSearchPlaceholder}
        emptyText={phaseEmptyText}
        clearable
        {disabled}
        onValueChange={onPhaseChange}
      />
    </div>
  </div>
</div>
