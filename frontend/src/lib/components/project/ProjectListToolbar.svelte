<script lang="ts">
  import { Label } from '$lib/components/ui/label/index.js';
  import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
  import ProjectPhaseSelect from '$lib/components/project/ProjectPhaseSelect.svelte';
  import type { ProjectStatusFilter } from '$lib/stores/projects/projectListStore.js';

  type StatusOption = {
    value: ProjectStatusFilter;
    label: string;
  };

  type Props = {
    statusLabel: string;
    statusValue: ProjectStatusFilter;
    options: StatusOption[];
    phaseLabel: string;
    phaseValue: string;
    allPhasesLabel: string;
    statusSearchPlaceholder: string;
    statusEmptyText: string;
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
    statusSearchPlaceholder,
    statusEmptyText,
    phaseSearchPlaceholder,
    phaseEmptyText,
    disabled = false,
    onStatusChange,
    onPhaseChange
  }: Props = $props();

  function isProjectStatusFilter(value: string): value is ProjectStatusFilter {
    return value === 'all' || value === 'planned' || value === 'ongoing' || value === 'completed';
  }

  function handleStatusChange(value: string): void {
    if (isProjectStatusFilter(value)) {
      onStatusChange(value);
    }
  }
</script>

<div class="rounded-xl border bg-card p-4 shadow-sm">
  <div class="grid gap-4 md:grid-cols-2">
    <div class="flex flex-col gap-2">
      <Label for="project_status_filter">{statusLabel}</Label>

      <StaticCombobox
        id="project_status_filter"
        items={options}
        value={statusValue}
        labelKey="label"
        idKey="value"
        width="w-full"
        popupWidth="w-full"
        searchPlaceholder={statusSearchPlaceholder}
        emptyText={statusEmptyText}
        {disabled}
        onValueChange={handleStatusChange}
      />
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
